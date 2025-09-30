package v2service

import (
	"context"

	"cosmossdk.io/math"
	"errors"
	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/pkg"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc"
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
	// Represents the APR for BTC staking as a decimal (e.g., 0.035 = 3.5%)
	BTCStakingAPR float64 `json:"btc_staking_apr"`
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

	// Fetch all finality providers stats
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

	// Calculate the APR for BTC staking on Babylon Genesis
	// The APR is calculated based on the activeTvl of the overall stats
	btcStakingAPR, errAprCalculation := s.getBTCStakingAPR(
		ctx, overallStats.ActiveTvl,
	)
	if errAprCalculation != nil {
		log.Ctx(ctx).Error().Err(errAprCalculation).
			Msg("error while calculating BTC staking APR")
		// in case of error we use zero value in BTCStakingAPR
	}

	return &OverallStatsPublic{
		ActiveTvl:               overallStats.ActiveTvl,
		ActiveDelegations:       overallStats.ActiveDelegations,
		TotalActiveTvl:          overallStats.ActiveTvl + phase1Stats.ActiveTvl,
		TotalActiveDelegations:  overallStats.ActiveDelegations + phase1Stats.ActiveDelegations,
		ActiveFinalityProviders: uint64(activeFinalityProvidersCount),
		TotalFinalityProviders:  uint64(len(finalityProviders)),
		BTCStakingAPR:           btcStakingAPR,
	}, nil
}

func (s *V2Service) getBTCStakingAPR(
	ctx context.Context, activeTvl int64,
) (float64, error) {
	// Skip calculation if activeTvl is 0
	if activeTvl <= 0 {
		return 0, nil
	}

	// Convert the activeTvl which is in satoshis to BTC as APR is calculated per
	// BTC
	btcTvl := float64(activeTvl) / 1e8

	btcPrice, err := s.sharedService.GetLatestBTCPrice(ctx)
	if err != nil {
		return 0, err
	}

	babyPrice, err := s.sharedService.GetLatestBABYPrice(ctx)
	if err != nil {
		return 0, err
	}

	// Calculate the APR of the BTC staking on Babylon Genesis
	// APR = (400,000,000 * BABY Price) / (Total BTC Staked * BTC price)
	rewards, err := s.getAnnualBabyRewardsForBTCStaking(ctx)
	if err != nil {
		return 0, err
	}

	btcStakingAPR := (rewards * babyPrice) / (btcTvl * btcPrice)

	return btcStakingAPR, nil
}

func (s *V2Service) getAnnualBabyRewardsForBTCStaking(ctx context.Context) (float64, error) {
	bbnClient := s.bbnClient

	var annualProvisions, stakingRewards math.LegacyDec
	var provisionsErr, stakingRewardsErr error

	wg := conc.NewWaitGroup()
	wg.Go(func() {
		annualProvisions, provisionsErr = bbnClient.AnnualProvisions(ctx)
	})
	wg.Go(func() {
		stakingRewards, stakingRewardsErr = bbnClient.BTCStakingRewardsPortion(ctx)
	})
	wg.Wait()

	// if one of methods failed - combine all errors and return as one error
	if provisionsErr != nil || stakingRewardsErr != nil {
		err := errors.Join(provisionsErr, stakingRewardsErr)
		return 0, err
	}

	annualRewards, err := annualProvisions.Mul(stakingRewards).QuoInt64(pkg.UbbnPerBaby).Float64()
	return annualRewards, err
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

// todo test this method
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

// todo test this method
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
