package clients

import (
	"testing"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	coinmarketcap "github.com/miguelmota/go-coinmarketcap/pro/v1"
	"github.com/stretchr/testify/require"
	"github.com/davecgh/go-spew/spew"
	"os"
)

func TestCMC(t *testing.T) {
	t.Skip("test for manual testing")

	const envVarName = "COINMARKETCAP_API_KEY"
	apiKey, found := os.LookupEnv(envVarName)
	require.True(t, found, "%s env var is required", envVarName)

	clients := New(&config.Config{
		ExternalAPIs: &config.ExternalAPIsConfig{
			CoinMarketCap: &config.CoinMarketCapConfig{
				APIKey:  apiKey,
				BaseURL: "https://pro-api.coinmarketcap.com/v1",
			},
		},
	})
	quotes, err := clients.CoinMarketCap.Cryptocurrency.LatestQuotes(&coinmarketcap.QuoteOptions{
		Symbol: "BTC",
	})
	require.NoError(t, err)

	spew.Dump(quotes)
}
