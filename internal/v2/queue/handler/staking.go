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

	// Check if delegation already exists
	exist, delError := h.Service.IsDelegationPresent(ctx, activeStakingEvent.StakingTxHashHex)
	if delError != nil {
		return delError
	}
	if exist {
		// Ignore the message as the delegation already exists. This is a duplicate message
		log.Ctx(ctx).Debug().Str("StakingTxHashHex", activeStakingEvent.StakingTxHashHex).
			Msg("delegation already exists")
		return nil
	}

	// // Perform the address lookup conversion
	// addressLookupErr := h.performAddressLookupConversion(ctx, activeStakingEvent.StakerBtcPkHex, types.Active)
	// if addressLookupErr != nil {
	// 	return addressLookupErr
	// }

	// Perform the stats calculation
	statsErr := h.Service.ProcessStakingStatsCalculation(
		ctx,
		activeStakingEvent.StakingTxHashHex,
		activeStakingEvent.StakerBtcPkHex,
		activeStakingEvent.FinalityProviderBtcPksHex,
		types.Active,
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
	statsErr := h.Service.ProcessStakingStatsCalculation(
		ctx,
		unbondingStakingEvent.StakingTxHashHex,
		unbondingStakingEvent.StakerBtcPkHex,
		unbondingStakingEvent.FinalityProviderBtcPksHex,
		types.Unbonding,
		unbondingStakingEvent.StakingAmount,
	)
	if statsErr != nil {
		log.Ctx(ctx).Error().Err(statsErr).Msg("Failed to process staking stats calculation")
		return statsErr
	}
	return nil
}

// Convert the staker's public key into corresponding BTC addresses for
// database lookup. This is performed only for active delegation events to
// prevent duplicated database writes.
func (h *V2QueueHandler) performAddressLookupConversion(ctx context.Context, stakerPkHex string, state types.DelegationState) *types.Error {
	// Perform the address lookup conversion only for active delegation events
	// to prevent duplicated database writes
	if state == types.Active {
		addErr := h.Service.ProcessAndSaveBtcAddresses(ctx, stakerPkHex)
		if addErr != nil {
			log.Ctx(ctx).Error().Err(addErr).Msg("Failed to process and save btc addresses")
			return addErr
		}
	}
	return nil
}
