package service

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Services) GetLatestBtcPriceUsd(ctx context.Context) (float64, error) {
	// Try to get price from MongoDB first
	btcPrice, err := s.DbClient.GetLatestBtcPrice(ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Document not found, fetch from CoinMarketCap
			price, err := s.Clients.CoinMarketCap.GetLatestBtcPrice(ctx)
			if err != nil {
				return 0, fmt.Errorf("failed to fetch price from CoinMarketCap: %w", err)
			}
			// Store in MongoDB with TTL
			if err := s.DbClient.SetBtcPrice(ctx, price); err != nil {
				return 0, fmt.Errorf("failed to cache btc price: %w", err)
			}
			return price, nil
		}
		// Handle other database errors
		return 0, fmt.Errorf("database error: %w", err)
	}
	return btcPrice.Price, nil
}
