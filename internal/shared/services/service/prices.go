package service

import (
	"context"
	"errors"
	"fmt"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	coinmarketcap "github.com/miguelmota/go-coinmarketcap/pro/v1"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Service) GetLatestBTCPrice(ctx context.Context) (float64, error) {
	// Try to get price from MongoDB first
	db := s.DbClients.SharedDBClient
	price, err := db.GetLatestPrice(ctx, dbmodel.SymbolBTC)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Document not found, fetch from CoinMarketCap

			// singleflight prevents sending multiple requests for btc quote from multiple goroutine
			// here we will make just 1 request, other goroutines will wait and receive whatever first one get
			// note that key in .Do call below is not very important unless we use the same string
			value, err, _ := s.singleFlightGroup.Do("fetch_btc", func() (any, error) {
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

func (s *Service) doGetLatestBTCPrice() (float64, error) {
	quotes, err := s.Clients.CoinMarketCap.Cryptocurrency.LatestQuotes(&coinmarketcap.QuoteOptions{
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
