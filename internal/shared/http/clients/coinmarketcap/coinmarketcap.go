package coinmarketcap

import (
	"context"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type CoinMarketCapClient struct {
	config         *config.CoinMarketCapConfig
	defaultHeaders map[string]string
	httpClient     *http.Client
}

type CMCResponse struct {
	Data map[string]CryptoData `json:"data"`
}

type CryptoData struct {
	Quote map[string]QuoteData `json:"quote"`
}

type QuoteData struct {
	Price float64 `json:"price"`
}

func NewCoinMarketCapClient(config *config.CoinMarketCapConfig) *CoinMarketCapClient {
	// Client is disabled if config is nil
	if config == nil {
		return nil
	}

	httpClient := &http.Client{}
	headers := map[string]string{
		"X-CMC_PRO_API_KEY": config.APIKey,
		"Accept":            "application/json",
	}

	return &CoinMarketCapClient{
		config,
		headers,
		httpClient,
	}
}

// Necessary for the BaseClient interface
func (c *CoinMarketCapClient) GetBaseURL() string {
	return c.config.BaseURL
}

func (c *CoinMarketCapClient) GetDefaultRequestTimeout() int {
	return int(c.config.Timeout.Milliseconds())
}

func (c *CoinMarketCapClient) GetHttpClient() *http.Client {
	return c.httpClient
}

func (c *CoinMarketCapClient) GetLatestBtcPrice(ctx context.Context) (float64, *types.Error) {
	// todo implement me
	return 0, nil
}
