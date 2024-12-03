package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

// Response structures - only include what we need
type CMCResponse struct {
	Data map[string]CryptoData `json:"data"`
}

type CryptoData struct {
	Quote map[string]QuoteData `json:"quote"`
}

type QuoteData struct {
	Price float64 `json:"price"`
}

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
	logger := log.Ctx(ctx)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/cryptocurrency/quotes/latest", s.cfg.ExternalAPIs.CoinMarketCap.BaseURL),
		nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("X-CMC_PRO_API_KEY", s.cfg.ExternalAPIs.CoinMarketCap.APIKey)
	req.Header.Set("Accept", "application/json")

	// Add query parameters
	q := req.URL.Query()
	q.Add("symbol", "BTC")
	req.URL.RawQuery = q.Encode()

	logger.Debug().
		Str("url", req.URL.String()).
		Msg("making request to CoinMarketCap")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: s.cfg.ExternalAPIs.CoinMarketCap.Timeout,
	}

	// Make the actual HTTP request using the client
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var cmcResp CMCResponse
	if err := json.Unmarshal(body, &cmcResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract BTC price in USD
	btcData, exists := cmcResp.Data["BTC"]
	if !exists {
		return 0, fmt.Errorf("BTC data not found in response")
	}

	usdQuote, exists := btcData.Quote["USD"]
	if !exists {
		return 0, fmt.Errorf("USD quote not found in response")
	}

	return usdQuote.Price, nil
}
