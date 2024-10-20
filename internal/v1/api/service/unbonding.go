package v1service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	"github.com/rs/zerolog/log"
)

// UnbondDelegation verifies the unbonding request and saves the unbonding tx into the DB.
// It returns an error if the delegation is not eligible for unbonding or if the unbonding request is invalid.
// If successful, it will change the delegation state to `unbonding_requested`
func (s *V1Service) UnbondDelegation(
	ctx context.Context,
	stakingTxHashHex,
	unbondingTxHashHex,
	unbondingTxHex,
	signatureHex string) *types.Error {
	// 1. check the delegation is eligible for unbonding
	delegationDoc, err := s.Service.DbClients.V1DBClient.FindDelegationByTxHashHex(ctx, stakingTxHashHex)
	if err != nil {
		if ok := db.IsNotFoundError(err); ok {
			log.Warn().Err(err).Msg("delegation not found, hence not eligible for unbonding")
			return types.NewErrorWithMsg(http.StatusForbidden, types.NotFound, "delegation not found")
		}
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching delegation")
		return types.NewError(http.StatusInternalServerError, types.InternalServiceError, err)
	}

	if delegationDoc.State != types.Active {
		log.Ctx(ctx).Warn().
			Str("stakingTxHashHex", stakingTxHashHex).
			Str("state", delegationDoc.State.ToString()).
			Msg("delegation state is not active, hence not eligible for unbonding")
		return types.NewErrorWithMsg(http.StatusForbidden, types.Forbidden, "delegation state is not active")
	}

	paramsVersion := s.GetVersionedGlobalParamsByHeight(delegationDoc.StakingTx.StartHeight)
	if paramsVersion == nil {
		log.Ctx(ctx).Error().Msg("failed to get global params")
		return types.NewErrorWithMsg(
			http.StatusInternalServerError, types.InternalServiceError,
			"failed to get global params based on the staking tx height",
		)
	}

	// 2. verify the unbonding request
	if err := utils.VerifyUnbondingRequest(
		delegationDoc.StakingTxHashHex,
		unbondingTxHashHex,
		unbondingTxHex,
		delegationDoc.StakerPkHex,
		delegationDoc.FinalityProviderPkHex,
		signatureHex,
		delegationDoc.StakingTx.TimeLock,
		delegationDoc.StakingTx.OutputIndex,
		delegationDoc.StakingValue,
		paramsVersion,
		s.Service.Cfg.Server.BTCNetParam,
	); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg(fmt.Sprintf("unbonding request did not pass unbonding request verification, staking tx hash: %s, unbonding tx hash: %s",
			delegationDoc.StakingTxHashHex, unbondingTxHashHex))
		return types.NewError(http.StatusForbidden, types.ValidationError, err)
	}

	// 3. save unbonding tx into DB
	err = s.Service.DbClients.V1DBClient.SaveUnbondingTx(ctx, stakingTxHashHex, unbondingTxHashHex, unbondingTxHex, signatureHex)
	if err != nil {
		if ok := db.IsDuplicateKeyError(err); ok {
			log.Ctx(ctx).Warn().Err(err).Msg("unbonding request already been submitted into the system")
			return types.NewError(http.StatusForbidden, types.Forbidden, err)
		} else if ok := db.IsNotFoundError(err); ok {
			log.Ctx(ctx).Warn().Err(err).Msg("no active delegation found for unbonding request")
			return types.NewError(http.StatusForbidden, types.Forbidden, err)
		}
		log.Ctx(ctx).Error().Err(err).Msg("failed to save unbonding tx")
		return types.NewError(http.StatusInternalServerError, types.InternalServiceError, err)
	}
	return nil
}

func (s *V1Service) IsEligibleForUnbondingRequest(ctx context.Context, stakingTxHashHex string) *types.Error {
	delegationDoc, err := s.Service.DbClients.V1DBClient.FindDelegationByTxHashHex(ctx, stakingTxHashHex)
	if err != nil {
		if ok := db.IsNotFoundError(err); ok {
			log.Ctx(ctx).Warn().Err(err).Msg("delegation not found, hence not eligible for unbonding")
			return types.NewErrorWithMsg(http.StatusForbidden, types.NotFound, "delegation not found")
		}
		log.Error().Err(err).Msg("error while fetching delegation")
		return types.NewError(http.StatusInternalServerError, types.InternalServiceError, err)
	}

	if delegationDoc.State != types.Active {
		log.Ctx(ctx).Warn().Msg("delegation state is not active, hence not eligible for unbonding")
		return types.NewErrorWithMsg(http.StatusForbidden, types.Forbidden, "delegation state is not active")
	}
	return nil
}

// TransitionToUnbondingState process the actual confirmed unbonding tx by updating the delegation state to `unbonding`
// It returns true if the delegation is found and successfully transitioned to unbonding state.
func (s *V1Service) TransitionToUnbondingState(
	ctx context.Context, stakingTxHashHex string,
	unbondingStartHeight, unbondingTimelock, unbondingOutputIndex uint64,
	unbondingTxHex string, unbondingStartTimestamp int64,
) *types.Error {
	err := s.Service.DbClients.V1DBClient.TransitionToUnbondingState(ctx, stakingTxHashHex, unbondingStartHeight, unbondingTimelock, unbondingOutputIndex, unbondingTxHex, unbondingStartTimestamp)
	if err != nil {
		if ok := db.IsNotFoundError(err); ok {
			log.Ctx(ctx).Warn().Str("stakingTxHashHex", stakingTxHashHex).Err(err).Msg("delegation not found or no longer eligible for unbonding")
			return types.NewErrorWithMsg(http.StatusForbidden, types.NotFound, "delegation not found or no longer eligible for unbonding")
		}
		log.Ctx(ctx).Error().Str("stakingTxHashHex", stakingTxHashHex).Err(err).Msg("failed to transition to unbonding state")
		return types.NewError(http.StatusInternalServerError, types.InternalServiceError, err)
	}
	return nil
}
