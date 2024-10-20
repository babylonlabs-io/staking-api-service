package v1queuehandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	queueClient "github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/rs/zerolog/log"
)

func (h *V1QueueHandler) WithdrawStakingHandler(ctx context.Context, messageBody string) *types.Error {
	var withdrawnStakingEvent queueClient.WithdrawStakingEvent
	err := json.Unmarshal([]byte(messageBody), &withdrawnStakingEvent)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal the message body into withdrawnStakingEvent")
		return types.NewError(http.StatusBadRequest, types.BadRequest, err)
	}

	// Check if the delegation is in the right state to process the withdrawn event.
	del, delErr := h.Service.GetDelegation(ctx, withdrawnStakingEvent.StakingTxHashHex)
	// Requeue if found any error. Including not found error
	if delErr != nil {
		return delErr
	}
	state := del.State

	stakingTxHashHex := withdrawnStakingEvent.GetStakingTxHashHex()

	if utils.Contains(utils.OutdatedStatesForWithdraw(), state) {
		// Ignore the message as the delegation state is withdrawn. Nothing to do anymore
		log.Ctx(ctx).Debug().Str("stakingTxHashHex", stakingTxHashHex).
			Msg("delegation state is outdated for withdrawn event")
		return nil
	}
	// Requeue if the current state is not in the qualified states to transition to withdrawn
	// We will wait for the unbonded message to be processed first.
	if !utils.Contains(utils.QualifiedStatesToWithdraw(), state) {
		errMsg := "delegation is not in the qualified state to transition to withdrawn"
		log.Ctx(ctx).Warn().Str("stakingTxHashHex", stakingTxHashHex).
			Str("state", state.ToString()).Msg(errMsg)
		return types.NewErrorWithMsg(http.StatusForbidden, types.Forbidden, errMsg)
	}

	// Transition to withdrawn state
	// Please refer to the README.md for the details on the event processing workflow
	transitionErr := h.Service.TransitionToWithdrawnState(
		ctx, withdrawnStakingEvent.StakingTxHashHex,
	)
	if transitionErr != nil {
		return transitionErr
	}

	return nil
}
