package service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

// ProcessLegacyStatsDeduction is used to keep phase-1 stats up to date with
// unbonding and migration of delegations into phase-2 by reducing the TVL and
// other stats.
// This method tolerates duplicated calls, only the first call will be processed.
func (s *Service) ProcessLegacyStatsDeduction(
	ctx context.Context, stakingTxHashHex, stakerPkHex, fpPkHex string, amount uint64,
) *types.Error {
	// Fetch existing or initialize the stats lock document if not exist
	// same type "unbonded" is used for unbonding and migration as staker can
	// only one action on the same delegation
	// Note: due to legacy data was using "unbonded" as the identified for
	// unbonding transaction in lock db. We will continue to use it for now.
	statsLockDocument, err := s.DbClients.V1DBClient.GetOrCreateStatsLock(
		ctx, stakingTxHashHex, types.Unbonded.ToString(),
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
			Msg("error while fetching stats lock document")
		return types.NewInternalServiceError(err)
	}
	// Subtract from the finality stats
	if !statsLockDocument.FinalityProviderStats {
		err = s.DbClients.V1DBClient.SubtractFinalityProviderStats(
			ctx, stakingTxHashHex, fpPkHex, amount,
		)
		if err != nil {
			if db.IsNotFoundError(err) {
				return nil
			}
			log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
				Msg("error while subtracting finality stats")
			return types.NewInternalServiceError(err)
		}
	}
	if !statsLockDocument.StakerStats {
		err = s.DbClients.V1DBClient.SubtractStakerStats(
			ctx, stakingTxHashHex, stakerPkHex, amount,
		)
		if err != nil {
			if db.IsNotFoundError(err) {
				return nil
			}
			log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
				Msg("error while subtracting staker stats")
			return types.NewInternalServiceError(err)
		}
	}
	// Subtract from the overall stats.
	// The overall stats should be the last to be updated as it has dependency
	// on staker stats.
	if !statsLockDocument.OverallStats {
		err = s.DbClients.V1DBClient.SubtractOverallStats(
			ctx, stakingTxHashHex, stakerPkHex, amount,
		)
		if err != nil {
			if db.IsNotFoundError(err) {
				return nil
			}
			log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
				Msg("error while subtracting overall stats")
			return types.NewInternalServiceError(err)
		}
	}
	return nil
}
