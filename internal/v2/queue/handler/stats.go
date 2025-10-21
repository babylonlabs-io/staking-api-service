package v2queuehandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	queueClient "github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/rs/zerolog/log"
)

// ActiveStakingHandler processes active staking events
func (h *V2QueueHandler) ActiveStakingHandler(ctx context.Context, messageBody string) *types.Error {
	// acknowledge the message
	var activeStakingEvent queueClient.StakingEvent
	err := json.Unmarshal([]byte(messageBody), &activeStakingEvent)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal the message body into ActiveStakingEvent")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
	}

	// Mark as v1 delegation as transitioned if it exists
	if err := h.Services.V2Service.MarkV1DelegationAsTransitioned(
		ctx,
		activeStakingEvent.StakingTxHashHex,
		activeStakingEvent.StakerBtcPkHex,
		activeStakingEvent.FinalityProviderBtcPksHex[0], // only a single FP is supported for now
		activeStakingEvent.StakingAmount,
	); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to mark v1 delegation as transitioned")
		return types.NewError(http.StatusInternalServerError, types.InternalServiceError, err)
	}

	// Perform the address lookup conversion
	addErr := h.Services.V1Service.ProcessAndSaveBtcAddresses(ctx, activeStakingEvent.StakerBtcPkHex)
	if addErr != nil {
		log.Ctx(ctx).Error().Err(addErr).Msg("Failed to process and save btc addresses")
		return addErr
	}

	statsErr := h.Services.V2Service.ProcessActiveDelegationStats(
		ctx,
		activeStakingEvent.StakingTxHashHex,
		activeStakingEvent.StakerBtcPkHex,
		activeStakingEvent.FinalityProviderBtcPksHex,
		activeStakingEvent.StakingAmount,
	)
	if statsErr != nil {
		log.Ctx(ctx).Error().Err(statsErr).Msg("Failed to process staking stats calculation")
		return statsErr
	}

	return nil
}

// UnbondingStakingHandler processes unbonding staking events
func (h *V2QueueHandler) UnbondingStakingHandler(ctx context.Context, messageBody string) *types.Error {
	var unbondingStakingEvent queueClient.StakingEvent
	err := json.Unmarshal([]byte(messageBody), &unbondingStakingEvent)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal the message body into UnbondingStakingEvent")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
	}

	// Perform the stats calculation
	statsErr := h.Services.V2Service.ProcessUnbondingDelegationStats(
		ctx,
		unbondingStakingEvent.StakingTxHashHex,
		unbondingStakingEvent.StakerBtcPkHex,
		unbondingStakingEvent.FinalityProviderBtcPksHex,
		unbondingStakingEvent.StakingAmount,
		unbondingStakingEvent.StateHistory,
	)
	if statsErr != nil {
		log.Ctx(ctx).Error().Err(statsErr).Msg("Failed to process staking stats calculation")
		return statsErr
	}
	return nil
}
