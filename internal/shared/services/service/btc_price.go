package service

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Service) GetLatestBtcPriceUsd(ctx context.Context) (float64, error) {
	// Try to get price from MongoDB first
	db := s.DbClients.SharedDBClient
	btcPrice, err := db.GetLatestBtcPrice(ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Document not found, fetch from CoinMarketCap
			price, err := s.Clients.CoinMarketCap.GetLatestBtcPrice(ctx)
			if err != nil {
				return 0, fmt.Errorf("failed to fetch price from CoinMarketCap: %w", err)
			}
			// Store in MongoDB with TTL
			if err := db.SetBtcPrice(ctx, price); err != nil {
				return 0, fmt.Errorf("failed to cache btc price: %w", err)
			}
			return price, nil
		}
		// Handle other database errors
		return 0, fmt.Errorf("database error: %w", err)
	}
	return btcPrice.Price, nil
}
