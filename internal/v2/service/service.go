package v2service

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/jellydator/ttlcache/v3"
)

type V2Service struct {
	dbClients     *dbclients.DbClients
	clients       *clients.Clients
	cfg           *config.Config
	sharedService *service.Service
	cache         *ttlcache.Cache[string, int64]
}

func New(sharedService *service.Service) (*V2Service, error) {
	return &V2Service{
		dbClients:     sharedService.DbClients,
		clients:       sharedService.Clients,
		cfg:           sharedService.Cfg,
		sharedService: sharedService,
		// for now cache is used only for one use case which is why value type V is int64 instead of "any"
		// also mind that cleanup process (.Start() method) is not called because of that
		cache: ttlcache.New[string, int64](),
	}, nil
}
