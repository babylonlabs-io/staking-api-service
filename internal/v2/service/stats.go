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

type StakingTxPublic struct {
	TxHashHex string `json:"tx_hash_hex"`
}

type StakerStatsPublic struct {
	StakerPkHex                  string            `json:"_id"`
	StakingTxs                   []StakingTxPublic `json:"staking_txs"`
	ActiveTvl                    int64             `json:"active_tvl"`
	WithdrawableTvl              int64             `json:"withdrawable_tvl"`
	SlashedTvl                   int64             `json:"slashed_tvl"`
	TotalActiveDelegations       int64             `json:"total_active_delegations"`
	TotalWithdrawableDelegations int64             `json:"total_withdrawable_delegations"`
	TotalSlashedDelegations      int64             `json:"total_slashed_delegations"`
}

func (s *V2Service) ProcessStakingStatsCalculation(ctx context.Context, stakingTxHashHex, stakerPkHex, fpPkHex string, state types.DelegationState, amount uint64) *types.Error {
	// TODO: Implement this
	return nil
}

func (s *V2Service) GetStakerStats(ctx context.Context, stakerPKHex string) (*StakerStatsPublic, *types.Error) {
	stakerStats, err := s.Service.DbClients.V2DBClient.GetStakerStats(ctx, stakerPKHex)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching staker stats")
		return nil, types.NewInternalServiceError(err)
	}

	stakingTxs := make([]StakingTxPublic, len(stakerStats.StakingTxs))
	for i, tx := range stakerStats.StakingTxs {
		stakingTxs[i] = StakingTxPublic{
			TxHashHex: tx.TxHashHex,
		}
	}

	return &StakerStatsPublic{
		StakerPkHex:                  stakerStats.StakerPkHex,
		StakingTxs:                   stakingTxs,
		ActiveTvl:                    stakerStats.ActiveTvl,
		WithdrawableTvl:              stakerStats.WithdrawableTvl,
		SlashedTvl:                   stakerStats.SlashedTvl,
		TotalActiveDelegations:       stakerStats.TotalActiveDelegations,
		TotalWithdrawableDelegations: stakerStats.TotalWithdrawableDelegations,
		TotalSlashedDelegations:      stakerStats.TotalSlashedDelegations,
	}, nil
}

func (s *V2Service) GetOverallStats(ctx context.Context) (*OverallStatsPublic, *types.Error) {
	overallStats, err := s.Service.DbClients.V2DBClient.GetOverallStats(ctx)
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
		ActiveStakers:          overallStats.ActiveStakers,
		TotalStakers:           overallStats.TotalStakers,
		ActiveFinalityProviders: overallStats.ActiveFinalityProviders,
		TotalFinalityProviders:  overallStats.TotalFinalityProviders,
	}, nil
}
