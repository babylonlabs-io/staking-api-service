package v2service

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

type OverallStatsPublic struct {
	ActiveTvl               int64  `json:"active_tvl"`
	ActiveDelegations       int64  `json:"active_delegations"`
	ActiveStakers           uint64 `json:"active_stakers"`
	ActiveFinalityProviders uint64 `json:"active_finality_providers"`
	TotalFinalityProviders  uint64 `json:"total_finality_providers"`
}

type StakerStatsPublic struct {
	StakerPkHex             string `json:"staker_pk_hex"`
	ActiveTvl               int64  `json:"active_tvl"`
	ActiveDelegations       int64  `json:"active_delegations"`
	UnbondingTvl            int64  `json:"unbonding_tvl"`
	UnbondingDelegations    int64  `json:"unbonding_delegations"`
	WithdrawableTvl         int64  `json:"withdrawable_tvl"`
	WithdrawableDelegations int64  `json:"withdrawable_delegations"`
}

func (s *V2Service) GetOverallStats(ctx context.Context) (*OverallStatsPublic, *types.Error) {
	overallStats, err := s.DbClients.V2DBClient.GetOverallStats(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching overall stats")
		return nil, types.NewInternalServiceError(err)
	}

	// TODO: count fetch will affect the performance of the API
	activeStakersCount, err := s.DbClients.V2DBClient.GetActiveStakersCount(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching active stakers count")
		return nil, types.NewInternalServiceError(err)
	}

	// TODO: ideally this should not be fetched from the indexer db
	finalityProviders, err := s.DbClients.IndexerDBClient.GetFinalityProviders(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching finality providers")
		return nil, types.NewInternalServiceError(err)
	}

	activeFinalityProvidersCount := 0
	for _, fp := range finalityProviders {
		if fp.State == indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_ACTIVE {
			activeFinalityProvidersCount++
		}
	}

	return &OverallStatsPublic{
		ActiveTvl:               overallStats.ActiveTvl,
		ActiveDelegations:       overallStats.ActiveDelegations,
		ActiveStakers:           uint64(activeStakersCount),
		ActiveFinalityProviders: uint64(activeFinalityProvidersCount),
		TotalFinalityProviders:  uint64(len(finalityProviders)),
	}, nil
}

func (s *V2Service) GetStakerStats(ctx context.Context, stakerPKHex string) (*StakerStatsPublic, *types.Error) {
	stakerStats, err := s.DbClients.V2DBClient.GetStakerStats(ctx, stakerPKHex)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("stakerPKHex", stakerPKHex).Msg("error while fetching staker stats")
		return nil, types.NewInternalServiceError(err)
	}

	return &StakerStatsPublic{
		StakerPkHex:             stakerStats.StakerPkHex,
		ActiveTvl:               stakerStats.ActiveTvl,
		ActiveDelegations:       stakerStats.ActiveDelegations,
		UnbondingTvl:            stakerStats.UnbondingTvl,
		UnbondingDelegations:    stakerStats.UnbondingDelegations,
		WithdrawableTvl:         stakerStats.WithdrawableTvl,
		WithdrawableDelegations: stakerStats.WithdrawableDelegations,
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
		err = s.DbClients.V2DBClient.HandleActiveStakerStats(
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
			ctx, stakingTxHashHex, amount,
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

	log.Debug().
		Str("stakingTxHashHex", stakingTxHashHex).
		Str("stakerPkHex", stakerPkHex).
		Msg("Finished processing active delegation stats")

	return nil
}

// ProcessUnbondingDelegationStats calculates the unbonding delegation stats
func (s *V2Service) ProcessUnbondingDelegationStats(
	ctx context.Context,
	stakingTxHashHex string,
	stakerPkHex string,
	fpBtcPkHexes []string,
	amount uint64,
	stateHistory []string,
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
		log.Debug().
			Str("stakingTxHashHex", stakingTxHashHex).
			Str("stakerPkHex", stakerPkHex).
			Str("event_type", "unbonding").
			Msg("Handling unbonding staker stats")

		err = s.DbClients.V2DBClient.HandleUnbondingStakerStats(
			ctx, stakingTxHashHex, stakerPkHex, amount, stateHistory,
		)
		if err != nil {
			if db.IsNotFoundError(err) {
				return nil
			}
			log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
				Msg("error while handling unbonding staker stats")
			return types.NewInternalServiceError(err)
		}
	}
	// Subtract from the overall stats.
	// The overall stats should be the last to be updated as it has dependency
	// on staker stats.
	if !statsLockDocument.OverallStats {
		err = s.DbClients.V2DBClient.SubtractOverallStats(
			ctx, stakingTxHashHex, amount,
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

	log.Debug().
		Str("stakingTxHashHex", stakingTxHashHex).
		Str("stakerPkHex", stakerPkHex).
		Str("event_type", "unbonding").
		Msg("Finished processing unbonding delegation stats")

	return nil
}

func (s *V2Service) ProcessWithdrawableDelegationStats(
	ctx context.Context,
	stakingTxHashHex,
	stakerPkHex string,
	amount uint64,
	stateHistory []string,
) *types.Error {
	statsLockDocument, err := s.DbClients.V2DBClient.GetOrCreateStatsLock(
		ctx,
		stakingTxHashHex,
		types.Withdrawable.ToString(),
	)
	if err != nil {
		log.Ctx(ctx).Error().
			Err(err).
			Str("staking_tx_hash", stakingTxHashHex).
			Msg("Failed to fetch stats lock document")
		return types.NewInternalServiceError(err)
	}

	if !statsLockDocument.StakerStats {
		log.Debug().
			Str("stakingTxHashHex", stakingTxHashHex).
			Str("stakerPkHex", stakerPkHex).
			Msg("Handling withdrawable staker stats")
		err = s.DbClients.V2DBClient.HandleWithdrawableStakerStats(
			ctx, stakingTxHashHex, stakerPkHex, amount, stateHistory,
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("stakingTxHashHex", stakingTxHashHex).
				Str("stakerPkHex", stakerPkHex).
				Msg("error while handling withdrawable staker stats")
			if db.IsNotFoundError(err) {
				return nil
			}
			return types.NewInternalServiceError(err)
		}
	}

	log.Debug().
		Str("stakingTxHashHex", stakingTxHashHex).
		Str("stakerPkHex", stakerPkHex).
		Msg("Finished processing withdrawable delegation stats")

	return nil
}

func (s *V2Service) ProcessWithdrawnDelegationStats(
	ctx context.Context,
	stakingTxHashHex,
	stakerPkHex string,
	amount uint64,
	stateHistory []string,
) *types.Error {
	statsLockDocument, err := s.DbClients.V2DBClient.GetOrCreateStatsLock(
		ctx,
		stakingTxHashHex,
		types.Withdrawn.ToString(),
	)
	if err != nil {
		log.Ctx(ctx).Error().
			Err(err).
			Str("staking_tx_hash", stakingTxHashHex).
			Msg("Failed to fetch stats lock document")
		return types.NewInternalServiceError(err)
	}

	if !statsLockDocument.StakerStats {
		log.Debug().
			Str("stakingTxHashHex", stakingTxHashHex).
			Str("stakerPkHex", stakerPkHex).
			Msg("Handling withdrawn staker stats")
		err = s.DbClients.V2DBClient.HandleWithdrawnStakerStats(
			ctx, stakingTxHashHex, stakerPkHex, amount, stateHistory,
		)
		if err != nil {
			if db.IsNotFoundError(err) {
				return nil
			}
			log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
				Msg("error while handling withdrawn delegation")
			return types.NewInternalServiceError(err)
		}
	}

	log.Debug().
		Str("stakingTxHashHex", stakingTxHashHex).
		Str("stakerPkHex", stakerPkHex).
		Msg("Finished processing withdrawn delegation stats")

	return nil
}
