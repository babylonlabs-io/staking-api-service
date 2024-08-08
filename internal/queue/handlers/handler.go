package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/services"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/rs/zerolog/log"
)

type QueueHandler struct {
	Services       *services.Services
	emitStatsEvent func(ctx context.Context, messageBody string) error
}

type MessageHandler func(ctx context.Context, messageBody string) *types.Error
type UnprocessableMessageHandler func(ctx context.Context, messageBody, receipt string) *types.Error

func NewQueueHandler(
	services *services.Services,
	emitStats func(ctx context.Context, messageBody string) error,
) *QueueHandler {
	return &QueueHandler{
		Services:       services,
		emitStatsEvent: emitStats,
	}
}

func (qh *QueueHandler) HandleUnprocessedMessage(ctx context.Context, messageBody, receipt string) *types.Error {
	return qh.Services.SaveUnprocessableMessages(ctx, messageBody, receipt)
}

func (qh *QueueHandler) EmitStatsEvent(ctx context.Context, statsEvent client.StatsEvent) *types.Error {
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
