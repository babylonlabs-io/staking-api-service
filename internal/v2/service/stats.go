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
	Id                      string `json:"_id"`
	ActiveTvl               int64  `json:"active_tvl"`
	TotalTvl                int64  `json:"total_tvl"`
	ActiveDelegations       int64  `json:"active_delegations"`
	TotalDelegations        int64  `json:"total_delegations"`
	ActiveStakers           uint64 `json:"active_stakers"`
	TotalStakers            uint64 `json:"total_stakers"`
	ActiveFinalityProviders uint64 `json:"active_finality_providers"`
	TotalFinalityProviders  uint64 `json:"total_finality_providers"`
}

type StakerStatsPublic struct {
	StakerPkHex             string `json:"_id"`
	ActiveTvl               int64  `json:"active_tvl"`
	WithdrawableTvl         int64  `json:"withdrawable_tvl"`
	SlashedTvl              int64  `json:"slashed_tvl"`
	ActiveDelegations       uint32 `json:"active_delegations"`
	WithdrawableDelegations uint32 `json:"withdrawable_delegations"`
	SlashedDelegations      uint32 `json:"slashed_delegations"`
}

func (s *V2Service) GetStakerStats(ctx context.Context, stakerPKHex string) (*StakerStatsPublic, *types.Error) {
	stakerStats, err := s.DbClients.V2DBClient.GetStakerStats(ctx, stakerPKHex)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching staker stats")
		return nil, types.NewInternalServiceError(err)
	}

	return &StakerStatsPublic{
		StakerPkHex:             stakerStats.StakerPkHex,
		ActiveTvl:               stakerStats.ActiveTvl,
		WithdrawableTvl:         stakerStats.WithdrawableTvl,
		SlashedTvl:              stakerStats.SlashedTvl,
		ActiveDelegations:       stakerStats.ActiveDelegations,
		WithdrawableDelegations: stakerStats.WithdrawableDelegations,
		SlashedDelegations:      stakerStats.SlashedDelegations,
	}, nil
}

func (s *V2Service) GetOverallStats(ctx context.Context) (*OverallStatsPublic, *types.Error) {
	overallStats, err := s.DbClients.V2DBClient.GetOverallStats(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching overall stats")
		return nil, types.NewInternalServiceError(err)
	}

	return &OverallStatsPublic{
		Id:                      overallStats.Id,
		ActiveTvl:               overallStats.ActiveTvl,
		TotalTvl:                overallStats.TotalTvl,
		ActiveDelegations:       overallStats.ActiveDelegations,
		TotalDelegations:        overallStats.TotalDelegations,
		ActiveStakers:           overallStats.ActiveStakers,
		TotalStakers:            overallStats.TotalStakers,
		ActiveFinalityProviders: overallStats.ActiveFinalityProviders,
		TotalFinalityProviders:  overallStats.TotalFinalityProviders,
	}, nil
}

// ProcessStakingStatsCalculation calculates the staking stats and updates the database.
// This method tolerates duplicated calls, only the first call will be processed.
func (s *V2Service) ProcessStakingStatsCalculation(
	ctx context.Context, stakingTxHashHex, stakerPkHex, fpPkHex string,
	state types.DelegationState, amount uint64,
) *types.Error {
	// Fetch existing or initialize the stats lock document if not exist
	statsLockDocument, err := s.Service.DbClients.V2DBClient.GetOrCreateStatsLock(
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

		// TODO: Add staker stats calculation

		// Add to the overall stats
		// The overall stats should be the last to be updated as it has dependency
		// on staker stats.
		if !statsLockDocument.OverallStats {
			err = s.Service.DbClients.V2DBClient.IncrementOverallStats(
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
		// TODO: Add finality provider stats calculation

		// TODO: Add staker stats calculation

		// Subtract from the overall stats.
		// The overall stats should be the last to be updated as it has dependency
		// on staker stats.
		if !statsLockDocument.OverallStats {
			err = s.Service.DbClients.V1DBClient.SubtractOverallStats(
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
