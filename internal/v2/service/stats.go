package v2service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

type OverallStatsPublic struct {
	ActiveTvl         int64  `json:"active_tvl"`
	TotalTvl          int64  `json:"total_tvl"`
	ActiveDelegations int64  `json:"active_delegations"`
	TotalDelegations  int64  `json:"total_delegations"`
	ActiveStakers     uint64 `json:"active_stakers"`
	TotalStakers      uint64 `json:"total_stakers"`
}

type StakerStatsPublic struct {
	StakerPkHex       string `json:"staker_pk_hex"`
	ActiveTvl         int64  `json:"active_tvl"`
	TotalTvl          int64  `json:"total_tvl"`
	ActiveDelegations int64  `json:"active_delegations"`
	TotalDelegations  int64  `json:"total_delegations"`
}

func (s *V2Service) GetOverallStats(ctx context.Context) (*OverallStatsPublic, *types.Error) {
	overallStats, err := s.DbClients.V2DBClient.GetOverallStats(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching overall stats")
		return nil, types.NewInternalServiceError(err)
	}

	return &OverallStatsPublic{
		ActiveTvl:         overallStats.ActiveTvl,
		TotalTvl:          overallStats.TotalTvl,
		ActiveDelegations: overallStats.ActiveDelegations,
		TotalDelegations:  overallStats.TotalDelegations,
		TotalStakers:      overallStats.TotalStakers,
	}, nil
}

func (s *V2Service) GetStakerStats(ctx context.Context, stakerPKHex string) (*StakerStatsPublic, *types.Error) {
	stakerStats, err := s.DbClients.V2DBClient.GetStakerStats(ctx, stakerPKHex)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("stakerPKHex", stakerPKHex).Msg("error while fetching staker stats")
		return nil, types.NewInternalServiceError(err)
	}

	return &StakerStatsPublic{
		StakerPkHex:       stakerStats.StakerPkHex,
		ActiveTvl:         stakerStats.ActiveTvl,
		TotalTvl:          stakerStats.TotalTvl,
		ActiveDelegations: stakerStats.ActiveDelegations,
		TotalDelegations:  stakerStats.TotalDelegations,
	}, nil
}

// ProcessStakingStatsCalculation calculates the staking stats and updates the database.
// This method tolerates duplicated calls, only the first call will be processed.
func (s *V2Service) ProcessStakingStatsCalculation(
	ctx context.Context,
	stakingTxHashHex, stakerPkHex string,
	finalityProviderBtcPksHex []string,
	state types.DelegationState, amount uint64,
) *types.Error {
	// Fetch existing or initialize the stats lock document if not exist
	statsLockDocument, err := s.DbClients.V2DBClient.GetOrCreateStatsLock(
		ctx, stakingTxHashHex, state.ToString(),
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
			Msg("error while fetching stats lock document")
		return types.NewInternalServiceError(err)
	}
	switch state {
	case types.Active:
		// TODO: Add finality provider stats calculation
		if !statsLockDocument.FinalityProviderStats {
			err = s.DbClients.V2DBClient.IncrementFinalityProviderStats(
				ctx, stakingTxHashHex, finalityProviderBtcPksHex, amount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
					Msg("error while incrementing finality stats")
				return types.NewInternalServiceError(err)
			}
		}

		if !statsLockDocument.StakerStats {
			// TODO: https://github.com/babylonlabs-io/staking-api-service/issues/162
			// // Convert the staker public key to multiple BTC addresses and save
			// // them in the database.
			// if addressConversionErr := s.ProcessAndSaveBtcAddresses(
			// 	ctx, stakerPkHex,
			// ); addressConversionErr != nil {
			// 	log.Ctx(ctx).Error().Err(addressConversionErr).
			// 		Str("stakingTxHashHex", stakingTxHashHex).
			// 		Msg("error while processing and saving btc addresses")
			// 	return types.NewInternalServiceError(addressConversionErr)
			// }
			err = s.DbClients.V2DBClient.IncrementStakerStats(
				ctx, stakingTxHashHex, stakerPkHex, amount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
					Msg("error while incrementing staker stats")
				return types.NewInternalServiceError(err)
			}
		}

		// Add to the overall stats
		// The overall stats should be the last to be updated as it has dependency
		// on staker stats.
		if !statsLockDocument.OverallStats {
			err = s.DbClients.V2DBClient.IncrementOverallStats(
				ctx, stakingTxHashHex, stakerPkHex, amount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					// This is a duplicate call, ignore it
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
					Msg("error while incrementing overall stats")
				return types.NewInternalServiceError(err)
			}
		}
	case types.Unbonded:
		// Subtract from the finality stats
		if !statsLockDocument.FinalityProviderStats {
			err = s.DbClients.V2DBClient.SubtractFinalityProviderStats(
				ctx, stakingTxHashHex, finalityProviderBtcPksHex, amount,
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
			err = s.DbClients.V2DBClient.SubtractStakerStats(
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
			err = s.DbClients.V2DBClient.SubtractOverallStats(
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
	default:
		return types.NewErrorWithMsg(
			http.StatusBadRequest,
			types.BadRequest,
			fmt.Sprintf("invalid delegation state for stats calculation: %s", state),
		)
	}
	return nil
}
