package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
	queueClient "github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/rs/zerolog/log"
)

func (h *QueueHandler) BtcInfoHandler(ctx context.Context, messageBody string) *types.Error {
	var btcInfo queueClient.BtcInfoEvent
	err := json.Unmarshal([]byte(messageBody), &btcInfo)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal the message body into btcInfo")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
	}

	statsErr := h.Services.ProcessBtcInfoStats(
		ctx, btcInfo.Height, btcInfo.ConfirmedTvl, btcInfo.UnconfirmedTvl,
	)
	if statsErr != nil {
		log.Error().Err(statsErr).Msg("Failed to process unconfirmed tvl stats")
		return types.NewInternalServiceError(statsErr)
	}
	return nil
}
