package clients

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients/coinmarketcap"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients/ordinals"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/integrations/chainalysis"
)

type Clients struct {
	Ordinals      ordinals.OrdinalsClient
	CoinMarketCap *coinmarketcap.Client
	Chainalysis   *chainalysis.Client
}

func New(cfg *config.Config) *Clients {
	var ordinalsClient ordinals.OrdinalsClient
	// If the assets config is set, create the ordinal related clients
	if cfg.Assets != nil {
		ordinalsClient = ordinals.New(cfg.Assets.Ordinals)
	}

	cmcConfig := cfg.ExternalAPIs.CoinMarketCap
	cmcClient := coinmarketcap.NewClient(cmcConfig.APIKey, int(cmcConfig.Timeout))

	var chainalysisClient *chainalysis.Client
	if cfg.ExternalAPIs.Chainalysis != nil {
		chainalysisClient = chainalysis.NewClient(
			cfg.ExternalAPIs.Chainalysis.APIKey,
			cfg.ExternalAPIs.Chainalysis.BaseURL,
		)
	}

	return &Clients{
		Ordinals:      ordinalsClient,
		CoinMarketCap: cmcClient,
		Chainalysis:   chainalysisClient,
	}
}
