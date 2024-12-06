package queue

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/tracing"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2queuehandler "github.com/babylonlabs-io/staking-api-service/internal/v2/queue/handler"
	"github.com/babylonlabs-io/staking-queue-client/client"
	queueConfig "github.com/babylonlabs-io/staking-queue-client/config"
	"github.com/rs/zerolog/log"
)

type Queues struct {
	Handlers                    *v2queuehandler.V2QueueHandler
	processingTimeout           time.Duration
	maxRetryAttempts            int32
	ActiveStakingQueueClient    client.QueueClient
	UnbondingStakingQueueClient client.QueueClient
}

func New(cfg *queueConfig.QueueConfig, service *services.Services) *Queues {
	activeStakingQueueClient, err := client.NewQueueClient(
		cfg, client.ActiveStakingQueueName,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating ActiveStakingQueueClient")
	}

	unbondingStakingQueueClient, err := client.NewQueueClient(
		cfg, client.UnbondingStakingQueueName,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating UnbondingStakingQueueClient")
	}

	handlers := v2queuehandler.NewV2QueueHandler(service)
	return &Queues{
		Handlers:                    handlers,
		processingTimeout:           cfg.QueueProcessingTimeout,
		maxRetryAttempts:            cfg.MsgMaxRetryAttempts,
		ActiveStakingQueueClient:    activeStakingQueueClient,
		UnbondingStakingQueueClient: unbondingStakingQueueClient,
	}
}

// Start all message processing
func (q *Queues) StartReceivingMessages() {
	// start processing messages from the active staking queue
	startQueueMessageProcessing(
		q.ActiveStakingQueueClient,
		q.Handlers.ActiveStakingHandler, q.Handlers.HandleUnprocessedMessage,
		q.maxRetryAttempts, q.processingTimeout,
	)
	startQueueMessageProcessing(
		q.UnbondingStakingQueueClient,
		q.Handlers.UnbondingStakingHandler, q.Handlers.HandleUnprocessedMessage,
		q.maxRetryAttempts, q.processingTimeout,
	)
	// ...add more queues here
}

// Turn off all message processing
func (q *Queues) StopReceivingMessages() {
	activeQueueErr := q.ActiveStakingQueueClient.Stop()
	if activeQueueErr != nil {
		log.Error().Err(activeQueueErr).
			Str("queueName", q.ActiveStakingQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}
	unbondingQueueErr := q.UnbondingStakingQueueClient.Stop()
	if unbondingQueueErr != nil {
		log.Error().Err(unbondingQueueErr).
			Str("queueName", q.UnbondingStakingQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}
	// ...add more queues here
}

func startQueueMessageProcessing(
	queueClient client.QueueClient,
	handler v2queuehandler.MessageHandler,
	unprocessableHandler v2queuehandler.UnprocessableMessageHandler,
	maxRetryAttempts int32, processingTimeout time.Duration,
) {
	messagesChan, err := queueClient.ReceiveMessages()
	log.Info().Str("queueName", queueClient.GetQueueName()).Msg("start receiving messages from queue")
	if err != nil {
		log.Fatal().Err(err).Str("queueName", queueClient.GetQueueName()).Msg("error setting up message channel from queue")
	}

	go func() {
		for message := range messagesChan {
			attempts := message.GetRetryAttempts()
			// For each message, create a new context with a deadline or timeout
			ctx, cancel := context.WithTimeout(context.Background(), processingTimeout)
			ctx = attachLoggerContext(ctx, message, queueClient)
			// Attach the tracingInfo for the message processing
			_, err := tracing.WrapWithSpan[any](ctx, "message_processing", func() (any, *types.Error) {
				timer := metrics.StartEventProcessingDurationTimer(queueClient.GetQueueName(), attempts)
				// Process the message
				err := handler(ctx, message.Body)
				if err != nil {
					timer(err.StatusCode)
				} else {
					timer(http.StatusOK)
				}
				return nil, err
			})
			if err != nil {
				recordErrorLog(err)
				// We will retry the message if it has not exceeded the max retry attempts
				// otherwise, we will dump the message into db for manual inspection and remove from the queue
				if attempts > maxRetryAttempts {
					log.Ctx(ctx).Error().Err(err).
						Msg("exceeded retry attempts, message will be dumped into db for manual inspection")
					metrics.RecordUnprocessableEntity(queueClient.GetQueueName())
					saveUnprocessableMsgErr := unprocessableHandler(ctx, message.Body, message.Receipt)
					if saveUnprocessableMsgErr != nil {
						log.Ctx(ctx).Error().Err(saveUnprocessableMsgErr).
							Msg("error while saving unprocessable message")
						metrics.RecordQueueOperationFailure("unprocessableHandler", queueClient.GetQueueName())
						cancel()
						continue
					}
				} else {
					log.Ctx(ctx).Error().Err(err).
						Msg("error while processing message from queue, will be requeued")
					reQueueErr := queueClient.ReQueueMessage(ctx, message)
					if reQueueErr != nil {
						log.Ctx(ctx).Error().Err(reQueueErr).
							Msg("error while requeuing message")
						metrics.RecordQueueOperationFailure("reQueueMessage", queueClient.GetQueueName())
					}
					cancel()
					continue
				}
			}

			delErr := queueClient.DeleteMessage(message.Receipt)
			if delErr != nil {
				log.Ctx(ctx).Error().Err(delErr).
					Msg("error while deleting message from queue")
				metrics.RecordQueueOperationFailure("deleteMessage", queueClient.GetQueueName())
			}

			tracingInfo := ctx.Value(tracing.TracingInfoKey)
			logEvent := log.Ctx(ctx).Debug()
			if tracingInfo != nil {
				logEvent = logEvent.Interface("tracingInfo", tracingInfo)
			}
			logEvent.Msg("message processed successfully")
			cancel()
		}
		log.Info().Str("queueName", queueClient.GetQueueName()).Msg("stopped receiving messages from queue")
	}()
}

func attachLoggerContext(ctx context.Context, message client.QueueMessage, queueClient client.QueueClient) context.Context {
	ctx = tracing.AttachTracingIntoContext(ctx)

	traceId := ctx.Value(tracing.TraceIdKey)
	return log.With().
		Str("receipt", message.Receipt).
		Str("queueName", queueClient.GetQueueName()).
		Interface("traceId", traceId).
		Logger().WithContext(ctx)
}

func recordErrorLog(err *types.Error) {
	if err.StatusCode >= http.StatusInternalServerError {
		log.Error().Err(err).Msg("event processing failed with 5xx error")
	} else {
		log.Warn().Err(err).Msg("event processing failed with 4xx error")
	}
}

func (q *Queues) IsConnectionHealthy() error {
	var errorMessages []string

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checkQueue := func(name string, client client.QueueClient) {
		if err := client.Ping(ctx); err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("%s is not healthy: %v", name, err))
		}
	}

	checkQueue("ActiveStakingQueueClient", q.ActiveStakingQueueClient)
	checkQueue("UnbondingStakingQueueClient", q.UnbondingStakingQueueClient)

	if len(errorMessages) > 0 {
		return fmt.Errorf(strings.Join(errorMessages, "; "))
	}
	return nil
}
