package queueclient

import (
	"context"
	"net/http"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/tracing"
	queuehandler "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/rs/zerolog/log"
)

func StartQueueMessageProcessing(
	queueClient client.QueueClient,
	handler queuehandler.MessageHandler, unprocessableHandler queuehandler.UnprocessableMessageHandler,
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
