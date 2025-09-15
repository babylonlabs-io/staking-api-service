package coinmarketcap

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/client"
)

const (
	// CoinMarketCap UCID for BTC.
	// https://coinmarketcap.com/currencies/bitcoin/
	BtcID = 1
	// CoinMarketCap UCID for BABY.
	// https://coinmarketcap.com/currencies/babylon/
	BabyID = 32198
)

const baseURL = "https://pro-api.coinmarketcap.com/v1"

type Client struct {
	apiKey         string
	requestTimeout int
	httpClient     *http.Client
}

func NewClient(apiKey string, requestTimeout int) *Client {
	return &Client{
		apiKey:         apiKey,
		requestTimeout: requestTimeout,
		httpClient:     &http.Client{},
	}
}

func (c *Client) GetBaseURL() string {
	return baseURL
}

func (c *Client) GetDefaultRequestTimeout() int {
	return c.requestTimeout
}

func (c *Client) GetHttpClient() *http.Client {
	return c.httpClient
}

// LatestQuotes returns latest quotes data for given ucid (unified cryptoasset id)
// if ucid doesn't exist this method return an error
func (c *Client) LatestQuotes(ctx context.Context, ucid int) (*QuoteLatest, error) {
	ucidStr := strconv.Itoa(ucid)

	path := "/cryptocurrency/quotes/latest"
	url := path + "?id=" + ucidStr

	opts := &client.HttpClientOptions{
		Path:         url,
		TemplatePath: path,
		Headers: map[string]string{
			"Accept":            "application/json",
			"X-CMC_PRO_API_KEY": c.apiKey,
		},
	}

	type empty struct{}
	resp, err := client.SendRequest[empty, response](
		ctx, c, http.MethodGet, opts, nil,
	)
	if err != nil {
		return nil, err
	}

	return resp.Data[ucidStr], nil
}

type QuoteLatest struct {
	ID                float64           `json:"id"`
	Name              string            `json:"name"`
	Symbol            string            `json:"symbol"`
	Slug              string            `json:"slug"`
	CirculatingSupply float64           `json:"circulating_supply"`
	TotalSupply       float64           `json:"total_supply"`
	MaxSupply         float64           `json:"max_supply"`
	DateAdded         string            `json:"date_added"`
	NumMarketPairs    float64           `json:"num_market_pairs"`
	CMCRank           float64           `json:"cmc_rank"`
	LastUpdated       string            `json:"last_updated"`
	Quote             map[string]*Quote `json:"quote"`
}

type Quote struct {
	Price            float64 `json:"price"`
	Volume24H        float64 `json:"volume_24h"`
	PercentChange1H  float64 `json:"percent_change_1h"`
	PercentChange24H float64 `json:"percent_change_24h"`
	PercentChange7D  float64 `json:"percent_change_7d"`
	MarketCap        float64 `json:"market_cap"`
	LastUpdated      string  `json:"last_updated"`
}

type response struct {
	Status struct {
		Timestamp    time.Time   `json:"timestamp"`
		ErrorCode    int         `json:"error_code"`
		ErrorMessage interface{} `json:"error_message"`
		Elapsed      int         `json:"elapsed"`
		CreditCount  int         `json:"credit_count"`
		Notice       interface{} `json:"notice"`
	} `json:"status"`
	Data map[string]*QuoteLatest `json:"data"`
}
