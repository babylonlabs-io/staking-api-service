package v1queuehandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	queueClient "github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/rs/zerolog/log"
)

// StatsHandler handles the processing of stats event from the message queue.
// This handler is responsible for processing non-critical events such as
// statistics calculations and other metadata-related tasks.
// It performs the following operations:
//  1. If the event corresponds to an active delegation, it transforms the staker's public key
//     into corresponding BTC addresses for lookup purposes, and saves them to the database.
//  2. Executes the staking statistics calculation using the provided event data.
//
// If any step fails, it logs the error and returns a corresponding error response.
// which will be sent back to the message queue for later retry
func (h *V1QueueHandler) StatsHandler(ctx context.Context, messageBody string) *types.Error {
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
	// For backwards compatibility reason, we will check msg version to determine
	// if we need to look up the overflow status from db
	// in version 1, we added the new field IsOverflow to the event
	// Below code will be removed after service being fully rollout
	isOverflow := statsEvent.IsOverflow
	if statsEvent.SchemaVersion < 1 {
		// Look up the overflow status from the database
		overflow, overflowErr := h.Service.GetDelegation(ctx, statsEvent.GetStakingTxHashHex())
		if overflowErr != nil {
			log.Ctx(ctx).Error().Err(overflowErr).Msg("Failed to get overflow status")
			return overflowErr
		}
		isOverflow = overflow.IsOverflow
	}

	// Perform the address lookup conversion
	addressLookupErr := h.performAddressLookupConversion(ctx, statsEvent.StakerPkHex, state)
	if addressLookupErr != nil {
		return addressLookupErr
	}

	// Perform the stats calculation only if the event is not an overflow event
	if !isOverflow {
		// Perform the stats calculation
		statsErr := h.Service.ProcessStakingStatsCalculation(
			ctx, statsEvent.StakingTxHashHex,
			statsEvent.StakerPkHex,
			statsEvent.FinalityProviderPkHex,
			state,
			statsEvent.StakingValue,
		)
		if statsErr != nil {
			log.Ctx(ctx).Error().Err(statsErr).Msg("Failed to process staking stats calculation")
			return statsErr
		}
	}
	return nil
}

// Convert the staker's public key into corresponding BTC addresses for
// database lookup. This is performed only for active delegation events to
// prevent duplicated database writes.
func (h *V1QueueHandler) performAddressLookupConversion(ctx context.Context, stakerPkHex string, state types.DelegationState) *types.Error {
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
