package clients

import (
	"github.com/babylonlabs-io/staking-api-service/internal/clients/ordinals"
	"github.com/babylonlabs-io/staking-api-service/internal/clients/unisat"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
)

type Clients struct {
	Ordinals ordinals.OrdinalsClientInterface
	Unisat   unisat.UnisatClientInterface
}

func New(cfg *config.Config) *Clients {
	var ordinalsClient *ordinals.OrdinalsClient
	var unisatClient *unisat.UnisatClient
	// If the assets config is set, create the ordinal related clients
	if cfg.Assets != nil {
		ordinalsClient = ordinals.NewOrdinalsClient(cfg.Assets.Ordinals)
		unisatClient = unisat.NewUnisatClient(cfg.Assets.Unisat)
	}

	return &Clients{
		Ordinals: ordinalsClient,
		Unisat:   unisatClient,
	}
}
