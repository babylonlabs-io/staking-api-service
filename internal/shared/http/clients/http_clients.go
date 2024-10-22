package clients

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients/ordinals"
)

type Clients struct {
	Ordinals ordinals.OrdinalsClient
}

func New(cfg *config.Config) *Clients {
	var ordinalsClient ordinals.OrdinalsClient
	// If the assets config is set, create the ordinal related clients
	if cfg.Assets != nil {
		ordinalsClient = ordinals.New(cfg.Assets.Ordinals)
	}

	return &Clients{
		Ordinals: ordinalsClient,
	}
}
