package v2service

import (
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/bbnclient"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/integrations/keybase"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/patrickmn/go-cache"
)

const aprCacheTTL = time.Minute

type V2Service struct {
	dbClients     *dbclients.DbClients
	clients       *clients.Clients
	cfg           *config.Config
	sharedService *service.Service
	keybaseClient *keybase.Client
	bbnClient     *bbnclient.BBNClient
	aprCache      *cache.Cache
}

func New(sharedService *service.Service, keybaseClient *keybase.Client, bbnClient *bbnclient.BBNClient) (*V2Service, error) {
	// because there are no dynamic keys cleanupInterval=-1 which means there is no goroutine that removes expired items
	aprCache := cache.New(aprCacheTTL, -1)

	return &V2Service{
		dbClients:     sharedService.DbClients,
		clients:       sharedService.Clients,
		cfg:           sharedService.Cfg,
		sharedService: sharedService,
		keybaseClient: keybaseClient,
		bbnClient:     bbnClient,
		aprCache:      aprCache,
	}, nil
}
