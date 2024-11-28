package v2service

import (
	"context"

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
