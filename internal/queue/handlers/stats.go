package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
	queueClient "github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/rs/zerolog/log"
)

func (h *QueueHandler) StatsHandler(ctx context.Context, messageBody string) *types.Error {
	var statsEvent queueClient.StatsEvent
	err := json.Unmarshal([]byte(messageBody), &statsEvent)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal the message body into statsEvent")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
	}

	state, err := types.FromStringToDelegationState(statsEvent.State)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to convert statsEvent.State to DelegationState")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
	}

	// Perform the stats calculation
	statsErr := h.Services.ProcessStakingStatsCalculation(
		ctx, statsEvent.StakingTxHashHex,
		statsEvent.StakerPkHex,
		statsEvent.FinalityProviderPkHex,
		state,
		statsEvent.StakingValue,
	)
	if statsErr != nil {
		log.Error().Err(statsErr).Msg("Failed to process staking stats calculation")
		return statsErr
	}
	return nil
}
