package v2service

import (
	"context"
	"errors"
	"fmt"

	cosmosMath "cosmossdk.io/math"

	costakingTypes "github.com/babylonlabs-io/babylon/v4/x/costaking/types"
	incentiveTypes "github.com/babylonlabs-io/babylon/v4/x/incentive/types"
	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/pkg"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

type apr struct {
	BtcStaking  float64 `json:"btc_staking_apr"`
	BabyStaking float64 `json:"baby_staking_apr"`
	CoStaking   float64 `json:"co_staking_apr"`
	Total       float64 `json:"total_apr"`
}
type StakingAPRPublic struct {
	Current                      apr     `json:"current"`
	AdditionalBabyNeededForBoost float64 `json:"additional_baby_needed_for_boost"`
	Boost                        apr     `json:"boost"`
}

// GetStakingAPR calculates personalized apr based on user's BTC and BABY stake
// btcStaked: total satoshis (confirmed + pending)
// babyStaked: total ubbn (confirmed + pending)
func (s *V2Service) GetStakingAPR(ctx context.Context, btcStaked, babyStaked int64) (*StakingAPRPublic, *types.Error) {
	// Fetch prices
	btcPrice, err := s.sharedService.GetLatestBTCPrice(ctx)
	if err != nil {
		return nil, types.NewInternalServiceError(fmt.Errorf("failed to get latest btc price: %w", err))
	}

	babyPrice, err := s.sharedService.GetLatestBABYPrice(ctx)
	if err != nil {
		return nil, types.NewInternalServiceError(fmt.Errorf("failed to get latest baby price: %w", err))
	}

	// Calculate BTC staking APR (this is the same for everyone)
	overallStats, err := s.dbClients.V2DBClient.GetOverallStats(ctx)
	if err != nil {
		return nil, types.NewInternalServiceError(fmt.Errorf("failed to get overall stats: %w", err))
	}

	btcStakingAPR, err := s.calculateBTCStakingAPR(ctx, overallStats.ActiveTvl, btcPrice, babyPrice)
	if err != nil {
		return nil, types.NewInternalServiceError(fmt.Errorf("failed to calculate btc staking apr: %w", err))
	}

	// Calculate BABY staking APR (this is the same for everyone)
	babyStakingAPR, err := s.calculateBabyStakingAPR(ctx)
	if err != nil {
		return nil, types.NewInternalServiceError(fmt.Errorf("failed to calculate baby staking apr: %w", err))
	}

	// Fetch co-staking data in parallel
	var totalCoStakingRewardSupply float64
	var globalTotalScore int64
	var scoreRatio int64
	var rewardSupplyErr, totalScoreErr, paramsErr error

	var wg conc.WaitGroup
	wg.Go(func() {
		totalCoStakingRewardSupply, rewardSupplyErr = s.calculateTotalCoStakingRewardSupply(ctx)
	})
	wg.Go(func() {
		totalScoreInt, err := s.bbnClient.CostakingTotalScore(ctx)
		if err != nil {
			totalScoreErr = err
			return
		}
		if !totalScoreInt.IsNil() {
			globalTotalScore = totalScoreInt.Int64()
		}
	})
	wg.Go(func() {
		params, err := s.bbnClient.CostakingParams(ctx)
		if err != nil {
			paramsErr = err
			return
		}
		scoreRatio = params.ScoreRatioBtcByBaby.Int64()
	})
	wg.Wait()

	if err := errors.Join(rewardSupplyErr, totalScoreErr, paramsErr); err != nil {
		return nil, types.NewInternalServiceError(fmt.Errorf("failed to fetch co-staking data: %w", err))
	}

	// Calculate current APR with user's current stake
	currentCoStakingAPR := s.calculateUserCoStakingAPR(
		btcStaked, babyStaked, globalTotalScore, scoreRatio,
		totalCoStakingRewardSupply, btcPrice, babyPrice,
	)

	// Calculate additional BABY needed for 100% eligibility
	requiredBabyForFullEligibility := btcStaked * scoreRatio
	additionalBabyNeeded := float64(max(0, requiredBabyForFullEligibility-babyStaked))

	// Calculate boost APR (at 100% eligibility)
	boostCoStakingAPR := s.calculateBoostCoStakingAPR(
		btcStaked, babyStaked, globalTotalScore, scoreRatio,
		totalCoStakingRewardSupply, btcPrice, babyPrice,
	)

	// Convert from ubbn to BABY for display
	additionalBabyNeededInBaby := additionalBabyNeeded / float64(pkg.UbbnPerBaby)

	return &StakingAPRPublic{
		Current: apr{
			BtcStaking:  btcStakingAPR,
			BabyStaking: babyStakingAPR,
			CoStaking:   currentCoStakingAPR,
			Total:       btcStakingAPR + babyStakingAPR + currentCoStakingAPR,
		},
		AdditionalBabyNeededForBoost: additionalBabyNeededInBaby,
		Boost: apr{
			BtcStaking:  btcStakingAPR,
			BabyStaking: babyStakingAPR,
			CoStaking:   boostCoStakingAPR,
			Total:       btcStakingAPR + babyStakingAPR + boostCoStakingAPR,
		},
	}, nil
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
	// The apr is calculated based on the activeTvl of the overall stats
	btcStakingAPR, errAprCalculation := s.getBTCStakingAPR(
		ctx, overallStats.ActiveTvl,
	)
	if errAprCalculation != nil {
		log.Ctx(ctx).Error().Err(errAprCalculation).
			Msg("error while calculating BTC staking apr")
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
	btcPrice, err := s.sharedService.GetLatestBTCPrice(ctx)
	if err != nil {
		return 0, err
	}

	babyPrice, err := s.sharedService.GetLatestBABYPrice(ctx)
	if err != nil {
		return 0, err
	}

	return s.calculateBTCStakingAPR(ctx, activeTvl, btcPrice, babyPrice)
}

func (s *V2Service) calculateBTCStakingAPR(ctx context.Context, activeTvl int64, btcPrice, babyPrice float64) (float64, error) {
	// Skip calculation if activeTvl is 0
	if activeTvl <= 0 {
		return 0, nil
	}

	// Convert the activeTvl which is in satoshis to BTC as APR is calculated per BTC
	btcTvl := float64(activeTvl) / 1e8

	// Calculate the APR of the BTC staking on Babylon Genesis
	// apr = (400,000,000 * BABY Price) / (Total BTC Staked * BTC price)
	rewards, err := s.getAnnualBabyRewardsForBTCStaking(ctx)
	if err != nil {
		return 0, err
	}

	btcStakingAPR := (rewards * babyPrice) / (btcTvl * btcPrice)
	return btcStakingAPR, nil
}

func (s *V2Service) calculateBabyStakingAPR(ctx context.Context) (float64, error) {
	bbnClient := s.bbnClient

	var totalSupplyErr, stakingPoolErr error
	var totalRewardsSupply cosmostypes.Coin
	var stakingPool stakingtypes.Pool

	var wg conc.WaitGroup
	wg.Go(func() {
		totalRewardsSupply, totalSupplyErr = bbnClient.TotalSupply(ctx, "ubbn")
	})
	wg.Go(func() {
		stakingPool, stakingPoolErr = bbnClient.StakingPool(ctx)
	})
	wg.Wait()

	// if we failed to get some info - return joined error
	if totalSupplyErr != nil || stakingPoolErr != nil {
		err := errors.Join(totalSupplyErr, stakingPoolErr)
		return 0, err
	}

	// 0.02
	babyInflationRate := cosmosMath.LegacyNewDecWithPrec(2, 2)

	//nolint: gocritic
	// totalBabyRewardsSupply = totalRewardsSupply * babyInflationRate
	totalBabyRewardsSupply := totalRewardsSupply.Amount.ToLegacyDec().Mul(babyInflationRate)
	totalBabyStaked := stakingPool.BondedTokens.ToLegacyDec()
	//nolint: gocritic
	// apr = totalBabyRewardsSupply / totalBabyStaked
	apr := totalBabyRewardsSupply.Quo(totalBabyStaked)

	aprFloat, err := apr.Float64()
	return aprFloat, err
}

func (s *V2Service) getAnnualBabyRewardsForBTCStaking(ctx context.Context) (float64, error) {
	bbnClient := s.bbnClient

	var annualProvisions, stakingRewards cosmosMath.LegacyDec
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

// calculateTotalCoStakingRewardSupply calculates the total annual co-staking reward supply
func (s *V2Service) calculateTotalCoStakingRewardSupply(ctx context.Context) (float64, error) {
	var annualProvisions cosmosMath.LegacyDec
	var incentiveParams *incentiveTypes.Params
	var costakingParams costakingTypes.Params
	var err1, err2, err3 error

	var wg conc.WaitGroup
	wg.Go(func() {
		annualProvisions, err1 = s.bbnClient.AnnualProvisions(ctx)
	})
	wg.Go(func() {
		incentiveParams, err2 = s.bbnClient.IncentiveParams(ctx)
	})
	wg.Go(func() {
		costakingParams, err3 = s.bbnClient.CostakingParams(ctx)
	})
	wg.Wait()

	if err := errors.Join(err1, err2, err3); err != nil {
		return 0, fmt.Errorf("failed to fetch co-staking supply data: %w", err)
	}

	// Cascade formula:
	// total_co_staking_reward_supply = annual_provisions × (1 - btc_portion - fp_portion) × costaking_portion

	annualProvisionsFloat, err := annualProvisions.Float64()
	if err != nil {
		return 0, fmt.Errorf("failed to convert annual provisions to float64: %w", err)
	}

	btcPortion, err := incentiveParams.BtcStakingPortion.Float64()
	if err != nil {
		return 0, fmt.Errorf("failed to convert btc staking portion to float64: %w", err)
	}

	fpPortion, err := incentiveParams.FpPortion.Float64()
	if err != nil {
		return 0, fmt.Errorf("failed to convert fp portion to float64: %w", err)
	}

	costakingPortion, err := costakingParams.CostakingPortion.Float64()
	if err != nil {
		return 0, fmt.Errorf("failed to convert costaking portion to float64: %w", err)
	}

	// Calculate what remains after incentive module takes its share
	afterIncentives := 1.0 - btcPortion - fpPortion

	// Co-staking gets a portion of what remains
	totalCoStakingRewardSupply := annualProvisionsFloat * afterIncentives * costakingPortion

	return totalCoStakingRewardSupply, nil
}

// calculateUserCoStakingAPR calculates the user's personalized co-staking apr
func (s *V2Service) calculateUserCoStakingAPR(
	btcStaked, babyStaked, globalTotalScore, scoreRatio int64,
	totalCoStakingRewardSupply, btcPrice, babyPrice float64,
) float64 {
	// Edge cases
	if btcStaked == 0 || globalTotalScore == 0 {
		return 0
	}

	// Calculate user's total score based on eligible satoshis
	// user_total_score = min(btcStaked, babyStaked / scoreRatio)
	eligibleSats := min(btcStaked, babyStaked/scoreRatio)
	userTotalScore := eligibleSats

	// Calculate pool share
	poolShare := float64(userTotalScore) / float64(globalTotalScore)

	// Calculate user's annual rewards in BABY (ubbn)
	userAnnualRewardsInBaby := poolShare * totalCoStakingRewardSupply

	// Convert to USD (Fisher correction: measure apr relative to BTC investment)
	userAnnualRewardsUSD := userAnnualRewardsInBaby * babyPrice / float64(pkg.UbbnPerBaby)
	userActiveBTCinUSD := float64(btcStaked) / 1e8 * btcPrice

	// Calculate apr as percentage: (annual_rewards_usd / btc_investment_usd)
	if userActiveBTCinUSD == 0 {
		return 0
	}

	apr := userAnnualRewardsUSD / userActiveBTCinUSD
	return apr
}

// calculateBoostCoStakingAPR calculates the boost apr at 100% eligibility
func (s *V2Service) calculateBoostCoStakingAPR(
	btcStaked, babyStaked, globalTotalScore, scoreRatio int64,
	totalCoStakingRewardSupply, btcPrice, babyPrice float64,
) float64 {
	// Edge cases
	if btcStaked == 0 || globalTotalScore == 0 {
		return 0
	}

	// Calculate current user score
	eligibleSats := min(btcStaked, babyStaked/scoreRatio)
	currentUserScore := eligibleSats

	// At 100% eligibility, user's score equals their BTC staked
	maxUserTotalScore := btcStaked

	// Calculate the increase in score
	scoreIncrease := maxUserTotalScore - currentUserScore

	// Adjust global score
	adjustedGlobalScore := globalTotalScore + scoreIncrease

	// Calculate boost pool share
	boostPoolShare := float64(maxUserTotalScore) / float64(adjustedGlobalScore)

	// Calculate boost annual rewards in BABY (ubbn)
	boostAnnualRewardsInBaby := boostPoolShare * totalCoStakingRewardSupply

	// Convert to USD (Fisher formula)
	boostAnnualRewardsUSD := boostAnnualRewardsInBaby * babyPrice / float64(pkg.UbbnPerBaby)
	userActiveBTCinUSD := float64(btcStaked) / 1e8 * btcPrice

	// Calculate apr as percentage
	if userActiveBTCinUSD == 0 {
		return 0
	}

	apr := boostAnnualRewardsUSD / userActiveBTCinUSD
	return apr
}
