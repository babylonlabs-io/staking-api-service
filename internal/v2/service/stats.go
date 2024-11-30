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

// processStatsSubtraction handles the common logic for subtracting stats
func (s *V2Service) processStatsSubtraction(
	ctx context.Context,
	stakingTxHashHex string,
	stakerPkHex string,
	fpBtcPkHexes []string,
	amount uint64,
) *types.Error {
	statsLockDocument, err := s.DbClients.V2DBClient.GetOrCreateStatsLock(
		ctx,
		stakingTxHashHex,
		types.Unbonding.ToString(), // use same state for both slashed and unbonding
	)
	if err != nil {
		log.Ctx(ctx).Error().
			Err(err).
			Str("staking_tx_hash", stakingTxHashHex).
			Msg("Failed to fetch stats lock document")
		return types.NewInternalServiceError(err)
	}

	// Helper for common error handling
	handleSubtractionError := func(err error, operation string) *types.Error {
		if err != nil {
			if db.IsNotFoundError(err) {
				return nil
			}
			log.Ctx(ctx).Error().
				Err(err).
				Str("staking_tx_hash", stakingTxHashHex).
				Str("operation", operation).
				Msg("Failed to subtract stats")
			return types.NewInternalServiceError(err)
		}
		return nil
	}

	// Subtract from finality provider stats
	if !statsLockDocument.FinalityProviderStats {
		if err := s.DbClients.V2DBClient.SubtractFinalityProviderStats(
			ctx, stakingTxHashHex, fpBtcPkHexes, amount,
		); err != nil {
			return handleSubtractionError(err, "finality_provider_stats")
		}
	}

	// Subtract from staker stats
	if !statsLockDocument.StakerStats {
		if err := s.DbClients.V2DBClient.SubtractStakerStats(
			ctx, stakingTxHashHex, stakerPkHex, amount,
		); err != nil {
			return handleSubtractionError(err, "staker_stats")
		}
	}

	// Subtract from overall stats
	if !statsLockDocument.OverallStats {
		if err := s.DbClients.V2DBClient.SubtractOverallStats(
			ctx, stakingTxHashHex, stakerPkHex, amount,
		); err != nil {
			return handleSubtractionError(err, "overall_stats")
		}
	}

	return nil
}

// ProcessUnbondingDelegationStats calculates the unbonding delegation stats
func (s *V2Service) ProcessUnbondingDelegationStats(
	ctx context.Context,
	stakingTxHashHex string,
	stakerPkHex string,
	fpBtcPkHexes []string,
	amount uint64,
) *types.Error {
	return s.processStatsSubtraction(ctx, stakingTxHashHex, stakerPkHex, fpBtcPkHexes, amount)
}

// ProcessSlashedFpStats processes stats for slashed finality providers
func (s *V2Service) ProcessSlashedFpStats(
	ctx context.Context,
	fpBtcPkHex string,
) *types.Error {
	slashedDelegations, err := s.DbClients.IndexerDBClient.GetSlashedFpDelegations(ctx, fpBtcPkHex)
	if err != nil {
		log.Ctx(ctx).Error().
			Err(err).
			Str("finality_provider_pk_hex", fpBtcPkHex).
			Msg("Failed to fetch slashed delegations")
		return types.NewInternalServiceError(err)
	}

	for _, delegation := range slashedDelegations {
		if err := s.processStatsSubtraction(
			ctx,
			delegation.StakingTxHashHex,
			delegation.StakerBtcPkHex,
			delegation.FinalityProviderBtcPksHex,
			delegation.StakingAmount,
		); err != nil {
			return err
		}
	}

	return nil
}
