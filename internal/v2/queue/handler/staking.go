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
