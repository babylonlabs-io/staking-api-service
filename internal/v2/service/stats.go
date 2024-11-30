package v2service

import (
	"context"

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

// ProcessActiveDelegationStats calculates the active delegation stats and updates the database.
func (s *V2Service) ProcessActiveDelegationStats(ctx context.Context, stakingTxHashHex, stakerPkHex string, fpBtcPkHexes []string, amount uint64) *types.Error {
	// Fetch existing or initialize the stats lock document if not exist
	statsLockDocument, err := s.DbClients.V2DBClient.GetOrCreateStatsLock(
		ctx, stakingTxHashHex, types.Active.ToString(),
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
			Msg("error while fetching stats lock document")
		return types.NewInternalServiceError(err)
	}

	if !statsLockDocument.FinalityProviderStats {
		err = s.DbClients.V2DBClient.IncrementFinalityProviderStats(
			ctx, stakingTxHashHex, fpBtcPkHexes, amount,
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
		// TODO: Convert the staker public key to multiple BTC addresses and save
		// them in the database.
		// https://github.com/babylonlabs-io/staking-api-service/issues/162

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

	return nil
}

// ProcessUnbondingDelegationStats calculates the unbonding delegation stats and updates the database.
func (s *V2Service) ProcessUnbondingDelegationStats(ctx context.Context, stakingTxHashHex, stakerPkHex string, fpBtcPkHexes []string, amount uint64) *types.Error {
	// Fetch existing or initialize the stats lock document if not exist
	statsLockDocument, err := s.DbClients.V2DBClient.GetOrCreateStatsLock(
		ctx, stakingTxHashHex, types.Unbonding.ToString(),
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
			Msg("error while fetching stats lock document")
		return types.NewInternalServiceError(err)
	}

	// Subtract from the finality stats
	if !statsLockDocument.FinalityProviderStats {
		err = s.DbClients.V2DBClient.SubtractFinalityProviderStats(
			ctx, stakingTxHashHex, fpBtcPkHexes, amount,
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

	return nil
}

func (s *V2Service) ProcessSlashedFpStats(ctx context.Context, fpBtcPkHex string) *types.Error {
	slashedFpDelegations, err := s.DbClients.IndexerDBClient.GetSlashedFpDelegations(ctx, fpBtcPkHex)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("fpBtcPkHex", fpBtcPkHex).Msg("error while fetching slashed fp delegations")
		return types.NewInternalServiceError(err)
	}

	for _, delegation := range slashedFpDelegations {
		statsLockDocument, err := s.DbClients.V2DBClient.GetOrCreateStatsLock(
			ctx, delegation.StakingTxHashHex, types.Unbonding.ToString(),
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", delegation.StakingTxHashHex).
				Msg("error while fetching stats lock document")
			return types.NewInternalServiceError(err)
		}

		// Subtract from the finality stats
		if !statsLockDocument.FinalityProviderStats {
			err = s.DbClients.V2DBClient.SubtractFinalityProviderStats(
				ctx, delegation.StakingTxHashHex, delegation.FinalityProviderBtcPksHex, delegation.StakingAmount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", delegation.StakingTxHashHex).
					Msg("error while subtracting finality stats")
				return types.NewInternalServiceError(err)
			}
		}

		if !statsLockDocument.StakerStats {
			err = s.DbClients.V2DBClient.SubtractStakerStats(
				ctx, delegation.StakingTxHashHex, delegation.StakerBtcPkHex, delegation.StakingAmount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", delegation.StakingTxHashHex).
					Msg("error while subtracting staker stats")
				return types.NewInternalServiceError(err)
			}
		}
		// Subtract from the overall stats.
		// The overall stats should be the last to be updated as it has dependency
		// on staker stats.
		if !statsLockDocument.OverallStats {
			err = s.DbClients.V2DBClient.SubtractOverallStats(
				ctx, delegation.StakingTxHashHex, delegation.StakerBtcPkHex, delegation.StakingAmount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", delegation.StakingTxHashHex).
					Msg("error while subtracting overall stats")
				return types.NewInternalServiceError(err)
			}
		}
	}

	return nil
}
