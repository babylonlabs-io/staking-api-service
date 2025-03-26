package v2service

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"time"
)

// for how long we keep the overall stats in memory
const overallStatsTTL = 5 * time.Minute

type V2Service struct {
	dbClients           *dbclients.DbClients
	clients             *clients.Clients
	cfg                 *config.Config
	sharedService       *service.Service
	overallStatsService *overallStatsService
}

func New(sharedService *service.Service) (*V2Service, error) {
	return &V2Service{
		dbClients:           sharedService.DbClients,
		clients:             sharedService.Clients,
		cfg:                 sharedService.Cfg,
		sharedService:       sharedService,
		overallStatsService: newOverallStatsService(sharedService.DbClients.V2DBClient, overallStatsTTL),
	}, nil
}
