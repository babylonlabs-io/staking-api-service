package clients

import (
	"github.com/babylonlabs-io/staking-api-service/internal/clients/ordinals"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
)

type Clients struct {
	Ordinals ordinals.OrdinalsClientInterface
}

func New(cfg *config.Config) *Clients {
	var ordinalsClient *ordinals.OrdinalsClient
	// If the assets config is set, create the ordinal related clients
	if cfg.Assets != nil {
		ordinalsClient = ordinals.NewOrdinalsClient(cfg.Assets.Ordinals)
	}

	return &Clients{
		Ordinals: ordinalsClient,
	}
}
