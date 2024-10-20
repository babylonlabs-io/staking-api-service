package queuehandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	queueclient "github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/rs/zerolog/log"
)

type QueueHandler struct {
	emitStatsEvent func(ctx context.Context, messageBody string) error
}

type MessageHandler func(ctx context.Context, messageBody string) *types.Error
type UnprocessableMessageHandler func(ctx context.Context, messageBody, receipt string) *types.Error

func New(
	emitStatsEvent func(ctx context.Context, messageBody string) error,
) *QueueHandler {
	return &QueueHandler{
		emitStatsEvent: emitStatsEvent,
	}
}

func (qh *QueueHandler) EmitStatsEvent(ctx context.Context, statsEvent queueclient.StatsEvent) *types.Error {
	jsonData, err := json.Marshal(statsEvent)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to marshal the stats event")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
	}

	err = qh.emitStatsEvent(ctx, string(jsonData))

	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit the stats event")
		return types.NewError(http.StatusInternalServerError, types.InternalServiceError, err)
	}
	return nil
}
