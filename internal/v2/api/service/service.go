package v2service

import (
	"context"

	service "github.com/babylonlabs-io/staking-api-service/internal/shared/api/service"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type V2Service struct {
	*service.Service
}

func New(
	ctx context.Context,
	cfg *config.Config,
	globalParams *types.GlobalParams,
	finalityProviders []types.FinalityProviderDetails,
	clients *clients.Clients,
	dbClients *dbclients.DbClients,
) (*V2Service, error) {
	service, err := service.New(ctx, cfg, globalParams, finalityProviders, clients, dbClients)
	if err != nil {
		return nil, err
	}

	return &V2Service{
		service,
	}, nil
}
