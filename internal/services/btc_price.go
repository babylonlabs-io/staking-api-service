package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Services) GetLatestBtcPriceUsd(ctx context.Context) (float64, error) {
	// Try to get price from MongoDB first
	btcPrice, err := s.DbClient.GetLatestBtcPrice(ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Document not found, fetch from CoinMarketCap
			price, err := s.fetchPriceFromCoinMarketCap(ctx)
			if err != nil {
				return 0, fmt.Errorf("failed to fetch price from CoinMarketCap: %w", err)
			}

			// Store in MongoDB with TTL
			if err := s.DbClient.SetBtcPrice(ctx, price); err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("failed to cache btc price")
				// Don't return error here, we can still return the price
			}

			return price, nil
		}
		// Handle other database errors
		return 0, fmt.Errorf("database error: %w", err)
	}

	return btcPrice.Price, nil
}

func (s *Services) fetchPriceFromCoinMarketCap(ctx context.Context) (float64, error) {
	// Implement CoinMarketCap API call here
	// Remember to use proper error handling and API key management
	return 0, nil // placeholder
}
