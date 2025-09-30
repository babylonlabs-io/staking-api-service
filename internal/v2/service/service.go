package v2service

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/bbnclient"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/integrations/keybase"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
)

type V2Service struct {
	dbClients     *dbclients.DbClients
	clients       *clients.Clients
	cfg           *config.Config
	sharedService *service.Service
	keybaseClient *keybase.Client
	bbnClient     *bbnclient.BBNClient
}

func New(sharedService *service.Service, keybaseClient *keybase.Client, bbnClient *bbnclient.BBNClient) (*V2Service, error) {
	return &V2Service{
		dbClients:     sharedService.DbClients,
		clients:       sharedService.Clients,
		cfg:           sharedService.Cfg,
		sharedService: sharedService,
		keybaseClient: keybaseClient,
		bbnClient:     bbnClient,
	}, nil
}
