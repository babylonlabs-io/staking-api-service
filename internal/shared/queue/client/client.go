package queueclient

import (
	"context"
	"net/http"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/tracing"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"

	"github.com/babylonlabs-io/staking-queue-client/client"
	queueConfig "github.com/babylonlabs-io/staking-queue-client/config"
	"github.com/rs/zerolog/log"
)

type Queue struct {
	ProcessingTimeout time.Duration
	MaxRetryAttempts  int32
	StatsQueueClient  client.QueueClient
}

func New(ctx context.Context, cfg *queueConfig.QueueConfig, service *services.Services) *Queue {
	statsQueueClient, err := client.NewQueueClient(
		cfg, client.StakingStatsQueueName,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating StatsQueueClient")
	}

	return &Queue{
		ProcessingTimeout: time.Duration(cfg.QueueProcessingTimeout) * time.Second,
		MaxRetryAttempts:  cfg.MsgMaxRetryAttempts,
		StatsQueueClient:  statsQueueClient,
	}
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
