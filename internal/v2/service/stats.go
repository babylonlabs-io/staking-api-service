package v2service

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

type OverallStatsPublic struct {
	ActiveTvl               int64  `json:"active_tvl"`
	ActiveDelegations       int64  `json:"active_delegations"`
	ActiveFinalityProviders uint64 `json:"active_finality_providers"`
	TotalFinalityProviders  uint64 `json:"total_finality_providers"`
	// This represents the total active tvl on BTC chain which includes
	// both phase-1 and phase-2 active tvl
	TotalActiveTvl int64 `json:"total_active_tvl"`
	// This represents the total active delegations on BTC chain which includes
	// both phase-1 and phase-2 active delegations
	TotalActiveDelegations int64 `json:"total_active_delegations"`
}

type StakerStatsPublic struct {
	ActiveTvl               int64 `json:"active_tvl"`
	ActiveDelegations       int64 `json:"active_delegations"`
	UnbondingTvl            int64 `json:"unbonding_tvl"`
	UnbondingDelegations    int64 `json:"unbonding_delegations"`
	WithdrawableTvl         int64 `json:"withdrawable_tvl"`
	WithdrawableDelegations int64 `json:"withdrawable_delegations"`
	SlashedTvl              int64 `json:"slashed_tvl"`
	SlashedDelegations      int64 `json:"slashed_delegations"`
}

func (s *V2Service) GetOverallStats(
	ctx context.Context,
) (*OverallStatsPublic, *types.Error) {
	overallStats, err := s.dbClients.V2DBClient.GetOverallStats(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching overall stats")
		return nil, types.NewInternalServiceError(err)
	}

	// TODO: ideally this should not be fetched from the indexer db
	finalityProviders, err := s.dbClients.IndexerDBClient.GetFinalityProviders(ctx)
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

	// Fetch phase-1 overall stats to calculate the total active tvl and
	// total active delegations
	phase1Stats, err := s.dbClients.V1DBClient.GetOverallStats(ctx)
	if err != nil {
		log.Ctx(ctx).Error().
			Err(err).Msg("error while fetching phase-1 overall stats")
		return nil, types.NewInternalServiceError(err)
	}

	if phase1Stats.ActiveTvl < 0 || phase1Stats.ActiveDelegations < 0 {
		log.Ctx(ctx).Error().
			Err(err).Msg("phase-1 overall stats are negative")
		metrics.RecordManualInterventionRequired("negative_stats_error")
		// Set the stats to 0 if they are negative as we do not want to
		// show negative stats in the UI.
		phase1Stats.ActiveTvl = 0
		phase1Stats.ActiveDelegations = 0
	}

	return &OverallStatsPublic{
		ActiveTvl:               overallStats.ActiveTvl,
		ActiveDelegations:       overallStats.ActiveDelegations,
		TotalActiveTvl:          overallStats.ActiveTvl + phase1Stats.ActiveTvl,
		TotalActiveDelegations:  overallStats.ActiveDelegations + phase1Stats.ActiveDelegations,
		ActiveFinalityProviders: uint64(activeFinalityProvidersCount),
		TotalFinalityProviders:  uint64(len(finalityProviders)),
	}, nil
}

func (s *V2Service) GetStakerStats(
	ctx context.Context,
	stakerPKHex string,
	stakerBabylonAddress *string,
) (*StakerStatsPublic, *types.Error) {
	states := []indexertypes.DelegationState{
		indexertypes.StateActive,
		indexertypes.StateUnbonding,
		indexertypes.StateWithdrawable,
		indexertypes.StateSlashed,
	}
	delegations, err := s.dbClients.IndexerDBClient.GetDelegationsInStates(
		ctx, stakerPKHex, stakerBabylonAddress, states,
	)
	if err != nil {
		logEvent := log.Ctx(ctx).Error().Err(err).
			Str("stakerPKHex", stakerPKHex)

		// Safely add babylon address to log if it exists
		if stakerBabylonAddress != nil {
			logEvent = logEvent.Str("stakerBabylonAddress", *stakerBabylonAddress)
		}

		logEvent.Msg("error while fetching staker stats")
		return nil, types.NewInternalServiceError(err)
	}

	var stats StakerStatsPublic

	for _, delegation := range delegations {
		amount := int64(delegation.StakingAmount)

		switch delegation.State {
		case indexertypes.StateActive:
			stats.ActiveTvl += amount
			stats.ActiveDelegations++
		case indexertypes.StateUnbonding:
			stats.UnbondingTvl += amount
			stats.UnbondingDelegations++
		case indexertypes.StateWithdrawable:
			stats.WithdrawableTvl += amount
			stats.WithdrawableDelegations++
		case indexertypes.StateSlashed:
			stats.SlashedTvl += amount
			stats.SlashedDelegations++
		}
	}

	return &stats, nil
}

// ProcessActiveDelegationStats calculates the active delegation stats and updates the database.
func (s *V2Service) ProcessActiveDelegationStats(ctx context.Context, stakingTxHashHex, stakerPkHex string, fpBtcPkHexes []string, amount uint64) *types.Error {
	// Fetch existing or initialize the stats lock document if not exist
	statsLockDocument, err := s.dbClients.V2DBClient.GetOrCreateStatsLock(
		ctx, stakingTxHashHex, types.Active.ToString(),
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
			Msg("error while fetching stats lock document")
		return types.NewInternalServiceError(err)
	}

	if !statsLockDocument.FinalityProviderStats {
		err = s.dbClients.V2DBClient.IncrementFinalityProviderStats(
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
		err = s.dbClients.V2DBClient.HandleActiveStakerStats(
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
		err = s.dbClients.V2DBClient.IncrementOverallStats(
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
	statsLockDocument, err := s.dbClients.V2DBClient.GetOrCreateStatsLock(
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
		err = s.dbClients.V2DBClient.SubtractFinalityProviderStats(
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

		err = s.dbClients.V2DBClient.HandleUnbondingStakerStats(
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
		err = s.dbClients.V2DBClient.SubtractOverallStats(
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
	statsLockDocument, err := s.dbClients.V2DBClient.GetOrCreateStatsLock(
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
		err = s.dbClients.V2DBClient.HandleWithdrawableStakerStats(
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
	statsLockDocument, err := s.dbClients.V2DBClient.GetOrCreateStatsLock(
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
		err = s.dbClients.V2DBClient.HandleWithdrawnStakerStats(
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
