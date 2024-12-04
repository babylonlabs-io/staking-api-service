package coinmarketcap

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

type CoinMarketCapClientInterface interface {
	GetLatestBtcPrice(ctx context.Context) (float64, *types.Error)
}
