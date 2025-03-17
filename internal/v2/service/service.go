package v2service

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
)

type V2Service struct {
	DbClients *dbclients.DbClients
	Clients   *clients.Clients
	Cfg       *config.Config
	service   *service.Service
}

func New(
	cfg *config.Config,
	clients *clients.Clients,
	dbClients *dbclients.DbClients,
	service *service.Service,
) (*V2Service, error) {
	return &V2Service{
		DbClients: dbClients,
		Clients:   clients,
		Cfg:       cfg,
		service:   service,
	}, nil
}
