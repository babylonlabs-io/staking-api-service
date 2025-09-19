package v1service

import (
	"context"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

type OverallStatsPublic struct {
	ActiveTvl         int64    `json:"active_tvl"`
	TotalTvl          int64    `json:"total_tvl"`
	ActiveDelegations int64    `json:"active_delegations"`
	TotalDelegations  int64    `json:"total_delegations"`
	TotalStakers      uint64   `json:"total_stakers"`
	UnconfirmedTvl    uint64   `json:"unconfirmed_tvl"`
	PendingTvl        uint64   `json:"pending_tvl"`
	BtcPriceUsd       *float64 `json:"btc_price_usd,omitempty"` // Optional field
}

type StakerStatsPublic struct {
	StakerPkHex       string `json:"staker_pk_hex"`
	ActiveTvl         int64  `json:"active_tvl"`
	TotalTvl          int64  `json:"total_tvl"`
	ActiveDelegations int64  `json:"active_delegations"`
	TotalDelegations  int64  `json:"total_delegations"`
}

// V1OverallStatsPublic represents the simplified overall stats for Phase-1
// that are maintained by the expiry-checker cron job recalculation.
type V1OverallStatsPublic struct {
	ActiveTvl         int64 `json:"active_tvl"`
	ActiveDelegations int64 `json:"active_delegations"`
}

// Deprecated: GetOverallStats uses incremental stats approach which has audit concerns.
// Use GetV1OverallStats for cron job recalculated stats instead.
func (s *V1Service) GetOverallStats(
	ctx context.Context,
) (*OverallStatsPublic, *types.Error) {
	stats, err := s.Service.DbClients.V1DBClient.GetOverallStats(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching overall stats")
		return nil, types.NewInternalServiceError(err)
	}

	// Fetch BTC price for backward compatibility with phase-1 API
	var btcPrice *float64
	if s.Service.Clients.CoinMarketCap != nil {
		price, err := s.Service.GetLatestBTCPrice(ctx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("error while fetching latest btc price")
		} else {
			btcPrice = &price
		}
	}

	return &OverallStatsPublic{
		ActiveTvl:         stats.ActiveTvl,
		TotalTvl:          stats.TotalTvl,
		ActiveDelegations: stats.ActiveDelegations,
		TotalDelegations:  stats.TotalDelegations,
		TotalStakers:      stats.TotalStakers,
		UnconfirmedTvl:    0, // No longer relevant in phase-2
		PendingTvl:        0, // No longer relevant in phase-2
		BtcPriceUsd:       btcPrice,
	}, nil
}

func (s *V1Service) GetStakerStats(
	ctx context.Context, stakerPkHex string,
) (*StakerStatsPublic, *types.Error) {
	stats, err := s.Service.DbClients.V1DBClient.GetStakerStats(ctx, stakerPkHex)
	if err != nil {
		// Not found error is not an error, return nil
		if db.IsNotFoundError(err) {
			return nil, nil
		}
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching staker stats")
		return nil, types.NewInternalServiceError(err)
	}

	return &StakerStatsPublic{
		StakerPkHex:       stakerPkHex,
		ActiveTvl:         stats.ActiveTvl,
		TotalTvl:          stats.TotalTvl,
		ActiveDelegations: stats.ActiveDelegations,
		TotalDelegations:  stats.TotalDelegations,
	}, nil
}

func (s *V1Service) GetTopStakersByActiveTvl(
	ctx context.Context, pageToken string,
) ([]StakerStatsPublic, string, *types.Error) {
	resultMap, err := s.Service.DbClients.V1DBClient.FindTopStakersByTvl(ctx, pageToken)
	if err != nil {
		if db.IsInvalidPaginationTokenError(err) {
			log.Ctx(ctx).Warn().Err(err).
				Msg("invalid pagination token while fetching top stakers by active tvl")
			return nil, "", types.NewError(http.StatusBadRequest, types.BadRequest, err)
		}
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching top stakers by active tvl")
		return nil, "", types.NewInternalServiceError(err)
	}
	var topStakersStats []StakerStatsPublic
	for _, d := range resultMap.Data {
		topStakersStats = append(topStakersStats, StakerStatsPublic{
			StakerPkHex:       d.StakerPkHex,
			ActiveTvl:         d.ActiveTvl,
			TotalTvl:          d.TotalTvl,
			ActiveDelegations: d.ActiveDelegations,
			TotalDelegations:  d.TotalDelegations,
		})
	}

	return topStakersStats, resultMap.PaginationToken, nil
}

func (s *V1Service) ProcessBtcInfoStats(
	ctx context.Context, btcHeight uint64, confirmedTvl uint64, unconfirmedTvl uint64,
) *types.Error {
	err := s.Service.DbClients.V1DBClient.UpsertLatestBtcInfo(ctx, btcHeight, confirmedTvl, unconfirmedTvl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while upserting latest btc info")
		return types.NewInternalServiceError(err)
	}
	return nil
}

// GetV1OverallStats retrieves the simplified overall stats from the
// new v1_overall_stats collection that is maintained by the expiry-checker cron job.
// This replaces the deprecated incremental stats approach with regular recalculation.
func (s *V1Service) GetV1OverallStats(
	ctx context.Context,
) (*V1OverallStatsPublic, *types.Error) {
	stats, err := s.Service.DbClients.V1DBClient.GetV1OverallStats(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching simplified V1 overall stats")
		return nil, types.NewInternalServiceError(err)
	}

	return &V1OverallStatsPublic{
		ActiveTvl:         stats.ActiveTvl,
		ActiveDelegations: stats.ActiveDelegations,
	}, nil
}
