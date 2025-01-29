package v2service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"golang.org/x/sync/singleflight"
)

type V2Service struct {
	DbClients *dbclients.DbClients
	Clients   *clients.Clients
	Cfg       *config.Config

	singleFlightGroup *singleflight.Group
}

func New(
	ctx context.Context,
	cfg *config.Config,
	clients *clients.Clients,
	dbClients *dbclients.DbClients,
) (*V2Service, error) {
	return &V2Service{
		DbClients:         dbClients,
		Clients:           clients,
		Cfg:               cfg,
		singleFlightGroup: &singleflight.Group{},
	}, nil
}
