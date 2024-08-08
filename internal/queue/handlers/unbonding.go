package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"github.com/babylonlabs-io/staking-api-service/internal/utils"
	queueClient "github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/rs/zerolog/log"
)

func (h *QueueHandler) UnbondingStakingHandler(ctx context.Context, messageBody string) *types.Error {
	var unbondingStakingEvent queueClient.UnbondingStakingEvent
	err := json.Unmarshal([]byte(messageBody), &unbondingStakingEvent)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal the message body into unbondingStakingEvent")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
	}

	// Check if the delegation is in the right state to process the unbonding event
	del, delErr := h.Services.GetDelegation(ctx, unbondingStakingEvent.StakingTxHashHex)
	// Requeue if found any error. Including not found error
	if delErr != nil {
		return delErr
	}
	state := del.State
	if utils.Contains(utils.OutdatedStatesForUnbonding(), state) {
		// Ignore the message as the delegation state already passed the unbonding state. This is an outdated duplication
		log.Ctx(ctx).Debug().Str("StakingTxHashHex", unbondingStakingEvent.StakingTxHashHex).
			Msg("delegation state is outdated for unbonding event")
		return nil
	}

	expireCheckErr := h.Services.ProcessExpireCheck(
		ctx, unbondingStakingEvent.StakingTxHashHex,
		unbondingStakingEvent.UnbondingStartHeight,
		unbondingStakingEvent.UnbondingTimeLock,
		types.UnbondingTxType,
	)
	if expireCheckErr != nil {
		return expireCheckErr
	}

	// We only emit the stats event for the staking tx that is not an overflow event
	if !del.IsOverflow {
		// Perform the async stats calculation by emit the stats event
		// NOTE: We no longer perform the stats calculation for timelock expired event
		// This is based on the assumption that phase 1 launch date + min timelock will be over the lauch of phase 2 date
		statsError := h.EmitStatsEvent(ctx, queueClient.NewStatsEvent(
			del.StakingTxHashHex,
			del.StakerPkHex,
			del.FinalityProviderPkHex,
			del.StakingValue,
			types.Unbonded.ToString(),
		))
		if statsError != nil {
			log.Ctx(ctx).Error().Err(statsError).Str("stakingTxHashHex", del.StakingTxHashHex).
				Msg("Failed to emit stats event for unbonding staking")
			return statsError
		}
	}

	// Save the unbonding staking delegation. This is the final step in the unbonding staking event processing
	// Please refer to the README.md for the details on the unbonding staking event processing workflow
	transitionErr := h.Services.TransitionToUnbondingState(
		ctx, unbondingStakingEvent.StakingTxHashHex, unbondingStakingEvent.UnbondingStartHeight,
		unbondingStakingEvent.UnbondingTimeLock, unbondingStakingEvent.UnbondingOutputIndex,
		unbondingStakingEvent.UnbondingTxHex, unbondingStakingEvent.UnbondingStartTimestamp,
	)
	if transitionErr != nil {
		return transitionErr
	}

	return nil
}
