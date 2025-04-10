package v1service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	coinmarketcap "github.com/miguelmota/go-coinmarketcap/pro/v1"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/singleflight"
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

// Add a singleflight group to the V1Service struct to prevent multiple concurrent requests
var singleFlightGroup singleflight.Group

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

	overallStatsV2, err := s.Service.DbClients.V2DBClient.GetOverallStats(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching v2 overall stats")
		return nil, types.NewInternalServiceError(err)
	}

	return &OverallStatsPublic{
		ActiveTvl:         stats.ActiveTvl + overallStatsV2.ActiveTvl,
		TotalTvl:          stats.TotalTvl,
		ActiveDelegations: stats.ActiveDelegations,
		TotalDelegations:  stats.TotalDelegations,
		TotalStakers:      stats.TotalStakers,
		UnconfirmedTvl:    0, // No longer relevant in phase-2
		PendingTvl:        0, // No longer relevant in phase-2
		BtcPriceUsd:       btcPrice,
	}, nil
}

// getLatestBTCPrice fetches the latest BTC price, first trying from MongoDB cache
// and falling back to CoinMarketCap if needed
func (s *V1Service) getLatestBTCPrice(ctx context.Context) (float64, error) {
	// Try to get price from MongoDB first
	db := s.Service.DbClients.SharedDBClient
	price, err := db.GetLatestPrice(ctx, dbmodel.SymbolBTC)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Document not found, fetch from CoinMarketCap

			// singleflight prevents sending multiple requests for btc quote from multiple goroutines
			// here we will make just 1 request, other goroutines will wait and receive whatever first one gets
			value, err, _ := singleFlightGroup.Do("fetch_btc", func() (interface{}, error) {
				return s.doGetLatestBTCPrice()
			})
			if err != nil {
				return 0, fmt.Errorf("failed to fetch price from CoinMarketCap: %w", err)
			}
			price := value.(float64)
			// Store in MongoDB with TTL
			if err := db.SetLatestPrice(ctx, dbmodel.SymbolBTC, price); err != nil {
				return 0, fmt.Errorf("failed to cache btc price: %w", err)
			}
			return price, nil
		}
		// Handle other database errors
		return 0, fmt.Errorf("database error: %w", err)
	}
	return price, nil
}

// doGetLatestBTCPrice fetches the latest BTC price directly from CoinMarketCap
func (s *V1Service) doGetLatestBTCPrice() (float64, error) {
	quotes, err := s.Service.Clients.CoinMarketCap.Cryptocurrency.LatestQuotes(&coinmarketcap.QuoteOptions{
		Symbol: "BTC",
	})
	if err != nil {
		return 0, err
	}

	if len(quotes) != 1 {
		return 0, fmt.Errorf("number of quotes from coinmarketcap != 1")
	}
	btcLatestQuote := quotes[0]

	btcToUsdQuote := btcLatestQuote.Quote["USD"]
	if btcToUsdQuote == nil {
		return 0, fmt.Errorf("USD quote not found in coinmarketcap response")
	}

	return btcToUsdQuote.Price, nil
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
