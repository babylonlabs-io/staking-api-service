package v2queuehandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2queueschema "github.com/babylonlabs-io/staking-api-service/internal/v2/queue/schema"
	"github.com/rs/zerolog/log"
)

// ActiveStakingHandler processes active staking events
func (h *V2QueueHandler) ActiveStakingHandler(ctx context.Context, messageBody string) *types.Error {
	// acknowledge the message
	var activeStakingEvent v2queueschema.ActiveStakingEvent
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

	// Perform the address lookup conversion
	addressLookupErr := h.performAddressLookupConversion(ctx, activeStakingEvent.StakerBtcPkHex, types.Active)
	if addressLookupErr != nil {
		return addressLookupErr
	}

	// Perform the stats calculation
	statsErr := h.Service.ProcessStakingStatsCalculation(
		ctx, activeStakingEvent.StakingTxHashHex,
		activeStakingEvent.StakerBtcPkHex,
		activeStakingEvent.FinalityProviderBtcPkHex,
		types.Active,
		activeStakingEvent.StakingValue,
	)
	if statsErr != nil {
		log.Ctx(ctx).Error().Err(statsErr).Msg("Failed to process staking stats calculation")
		return statsErr
	}

	return nil
}

// HandleVerifiedStaking processes verified staking events
func (h *V2QueueHandler) VerifiedStakingHandler(ctx context.Context, messageBody string) *types.Error {
	var verifiedStakingEvent v2queueschema.VerifiedStakingEvent
	err := json.Unmarshal([]byte(messageBody), &verifiedStakingEvent)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal the message body into VerifiedStakingEvent")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
	}
	return nil
}

// PendingStakingHandler processes pending staking events
func (h *V2QueueHandler) PendingStakingHandler(ctx context.Context, messageBody string) *types.Error {
	var pendingStakingEvent v2queueschema.PendingStakingEvent
	err := json.Unmarshal([]byte(messageBody), &pendingStakingEvent)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal the message body into PendingStakingEvent")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
	}
	return nil
}

// UnbondingStakingHandler processes unbonding staking events
func (h *V2QueueHandler) UnbondingStakingHandler(ctx context.Context, messageBody string) *types.Error {
	var unbondingStakingEvent v2queueschema.UnbondingStakingEvent
	err := json.Unmarshal([]byte(messageBody), &unbondingStakingEvent)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal the message body into UnbondingStakingEvent")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
	}
	return nil
}

// ExpiredStakingHandler processes expired (withdrawable) staking events
func (h *V2QueueHandler) ExpiredStakingHandler(ctx context.Context, messageBody string) *types.Error {
	var expiredStakingEvent v2queueschema.ExpiredStakingEvent
	err := json.Unmarshal([]byte(messageBody), &expiredStakingEvent)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal the message body into ExpiredStakingEvent")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
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
