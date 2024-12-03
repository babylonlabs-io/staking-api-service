package services

import (
	"context"

	"github.com/rs/zerolog/log"
)

func (s *Services) GetLatestBtcPriceUsd(ctx context.Context) (float64, error) {
	// Try to get price from MongoDB first
	btcPrice, err := s.DbClient.GetLatestBtcPrice(ctx)
	if err == nil {
		return btcPrice.Price, nil
	}

	// If not found or expired, fetch from CoinMarketCap
	price, err := s.fetchPriceFromCoinMarketCap(ctx)
	if err != nil {
		return 0, err
	}

	// Store in MongoDB with TTL
	err = s.DbClient.SetBtcPrice(ctx, price)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to cache btc price")
		// Don't return error here, we can still return the price
	}

	return price, nil
}

func (s *Services) fetchPriceFromCoinMarketCap(ctx context.Context) (float64, error) {
	// Implement CoinMarketCap API call here
	// Remember to use proper error handling and API key management
	return 0, nil // placeholder
}
