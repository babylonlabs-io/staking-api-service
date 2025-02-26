package clients

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients/ordinals"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/integrations/chainanalysis"
	cmc "github.com/miguelmota/go-coinmarketcap/pro/v1"
)

type Clients struct {
	Ordinals      ordinals.OrdinalsClient
	CoinMarketCap *cmc.Client
	ChainAnalysis *chainanalysis.Client
}

func New(cfg *config.Config) *Clients {
	var ordinalsClient ordinals.OrdinalsClient
	// If the assets config is set, create the ordinal related clients
	if cfg.Assets != nil {
		ordinalsClient = ordinals.New(cfg.Assets.Ordinals)
	}

	var cmcClient *cmc.Client
	if cfg.ExternalAPIs != nil && cfg.ExternalAPIs.CoinMarketCap != nil {
		cmcClient = cmc.NewClient(&cmc.Config{
			ProAPIKey: cfg.ExternalAPIs.CoinMarketCap.APIKey,
		})
	}

	var chainAnalysisClient *chainanalysis.Client
	if cfg.ExternalAPIs != nil && cfg.ExternalAPIs.ChainAnalysis != nil {
		chainAnalysisClient = chainanalysis.NewClient(
			cfg.ExternalAPIs.ChainAnalysis.APIKey,
			cfg.ExternalAPIs.ChainAnalysis.BaseURL,
		)
	}

	return &Clients{
		Ordinals:      ordinalsClient,
		CoinMarketCap: cmcClient,
		ChainAnalysis: chainAnalysisClient,
	}
}
