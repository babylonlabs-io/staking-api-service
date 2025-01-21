package clients

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients/ordinals"
)

type Clients struct {
	Ordinals      ordinals.OrdinalsClient
	CoinMarketCap coinmarketcap.CoinMarketCapClientInterface // todo for review: move to another location?
}

func New(cfg *config.Config) *Clients {
	var ordinalsClient ordinals.OrdinalsClient
	// If the assets config is set, create the ordinal related clients
	if cfg.Assets != nil {
		ordinalsClient = ordinals.New(cfg.Assets.Ordinals)
	}

	var coinMarketCapClient *coinmarketcap.CoinMarketCapClient
	if cfg.ExternalAPIs != nil && cfg.ExternalAPIs.CoinMarketCap != nil {
		coinMarketCapClient = coinmarketcap.NewCoinMarketCapClient(cfg.ExternalAPIs.CoinMarketCap)
	}

	return &Clients{
		Ordinals:      ordinalsClient,
		CoinMarketCap: coinMarketCapClient,
	}
}
