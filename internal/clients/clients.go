package clients

import (
	"github.com/babylonlabs-io/staking-api-service/internal/clients/coinmarketcap"
	"github.com/babylonlabs-io/staking-api-service/internal/clients/ordinals"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
)

type Clients struct {
	Ordinals      ordinals.OrdinalsClientInterface
	CoinMarketCap coinmarketcap.CoinMarketCapClientInterface
}

func New(cfg *config.Config) *Clients {
	var ordinalsClient *ordinals.OrdinalsClient
	// If the assets config is set, create the ordinal related clients
	if cfg.Assets != nil {
		ordinalsClient = ordinals.NewOrdinalsClient(cfg.Assets.Ordinals)
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
