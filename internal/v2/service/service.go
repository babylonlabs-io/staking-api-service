package v2service

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"golang.org/x/sync/singleflight"
)

type V2Service struct {
	DbClients         *dbclients.DbClients
	Clients           *clients.Clients
	Cfg               *config.Config
	singleFlightGroup *singleflight.Group
}

func New(sharedService *service.Service) (*V2Service, error) {
	return &V2Service{
		DbClients:         sharedService.DbClients,
		Clients:           sharedService.Clients,
		Cfg:               sharedService.Cfg,
		singleFlightGroup: &singleflight.Group{},
	}, nil
}
