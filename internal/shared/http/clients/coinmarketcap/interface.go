package coinmarketcap

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

//go:generate mockery --name=CoinMarketCapClientInterface --output=../../../../../tests/mocks --outpkg=mocks --filename=mock_coinmarketcap_client.go
type CoinMarketCapClientInterface interface {
	GetLatestBtcPrice(ctx context.Context) (float64, *types.Error)
}
