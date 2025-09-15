package service

import (
	"context"
	"errors"
	"fmt"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients/coinmarketcap"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Service) GetLatestBTCPrice(ctx context.Context) (float64, error) {
	// Try to get price from MongoDB first
	db := s.DbClients.SharedDBClient
	btcPrice, err := db.GetLatestPrice(ctx, dbmodel.SymbolBTC)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Document not found, fetch from CoinMarketCap

			// singleflight prevents sending multiple requests for btc quote from multiple goroutine
			// here we will make just 1 request, other goroutines will wait and receive whatever first one get
			// note that key in .Do call below is not very important unless we use the same string
			value, err, _ := s.singleFlightGroup.Do("fetch_btc", func() (any, error) {
				return s.doGetLatestBTCPrice(ctx)
			})
			if err != nil {
				return 0, fmt.Errorf("failed to fetch BTC price from CoinMarketCap: %w", err)
			}
			btcPrice := value.(float64)
			// Store in MongoDB with TTL
			if err := db.SetLatestPrice(ctx, dbmodel.SymbolBTC, btcPrice); err != nil {
				return 0, fmt.Errorf("failed to cache btc price: %w", err)
			}
			return btcPrice, nil
		}
		// Handle other database errors
		return 0, fmt.Errorf("database error: %w", err)
	}
	return btcPrice, nil
}

func (s *Service) GetLatestBABYPrice(ctx context.Context) (float64, error) {
	// Try to get price from MongoDB first
	db := s.DbClients.SharedDBClient
	babyPrice, err := db.GetLatestPrice(ctx, dbmodel.SymbolBABY)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			value, err, _ := s.singleFlightGroup.Do("fetch_baby", func() (any, error) {
				return s.doGetLatestBABYPrice(ctx)
			})
			if err != nil {
				return 0, fmt.Errorf("failed to fetch BABY price from CoinMarketCap: %w", err)
			}
			babyPrice := value.(float64)
			// Store in MongoDB with TTL
			if err := db.SetLatestPrice(ctx, dbmodel.SymbolBABY, babyPrice); err != nil {
				return 0, fmt.Errorf("failed to cache BABY price: %w", err)
			}
			return babyPrice, nil
		}
		// Handle other database errors
		return 0, fmt.Errorf("database error: %w", err)
	}
	return babyPrice, nil
}

func (s *Service) doGetLatestBTCPrice(ctx context.Context) (float64, error) {
	targetQuote, err := s.Clients.CoinMarketCap.LatestQuote(ctx, coinmarketcap.BtcID)
	if err != nil {
		return 0, err
	}

	btcToUsdQuote := targetQuote.Quote["USD"]
	if btcToUsdQuote == nil {
		return 0, fmt.Errorf("USD quote not found in coinmarketcap response for BTC")
	}

	return btcToUsdQuote.Price, nil
}

func (s *Service) doGetLatestBABYPrice(ctx context.Context) (float64, error) {
	targetQuote, err := s.Clients.CoinMarketCap.LatestQuote(ctx, coinmarketcap.BabyID)
	if err != nil {
		return 0, err
	}

	babyToUsdQuote := targetQuote.Quote["USD"]
	if babyToUsdQuote == nil {
		return 0, fmt.Errorf("USD quote not found in coinmarketcap response for BABY")
	}

	return babyToUsdQuote.Price, nil
}
