package coinmarketcap

import (
	"context"
	"net/http"

	baseclient "github.com/babylonlabs-io/staking-api-service/internal/clients/base"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
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
	path := "/cryptocurrency/quotes/latest"

	opts := &baseclient.BaseClientOptions{
		Path:         path + "?symbol=BTC",
		TemplatePath: path,
		Headers:      c.defaultHeaders,
	}

	// Use struct{} for input (no request body)
	// Use CMCResponse for response type
	response, err := baseclient.SendRequest[struct{}, CMCResponse](
		ctx, c, http.MethodGet, opts, nil,
	)
	if err != nil {
		return 0, err
	}

	btcData, exists := response.Data["BTC"]
	if !exists {
		return 0, types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			"BTC data not found in response",
		)
	}

	usdQuote, exists := btcData.Quote["USD"]
	if !exists {
		return 0, types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			"USD quote not found in response",
		)
	}

	return usdQuote.Price, nil
}
