package services

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1service "github.com/babylonlabs-io/staking-api-service/internal/v1/api/service"
	v2service "github.com/babylonlabs-io/staking-api-service/internal/v2/api/service"
)

type Services struct {
	V1Service v1service.V1ServiceInterface
	V2Service v2service.V2ServiceInterface
}

func New(
	ctx context.Context,
	cfg *config.Config,
	globalParams *types.GlobalParams,
	finalityProviders []types.FinalityProviderDetails,
	clients *clients.Clients,
	dbClients *dbclients.DbClients,
) (*Services, error) {
	v1Service, err := v1service.New(ctx, cfg, globalParams, finalityProviders, clients, dbClients)
	if err != nil {
		return nil, err
	}
	v2Service, err := v2service.New(ctx, cfg, globalParams, finalityProviders, clients, dbClients)
	if err != nil {
		return nil, err
	}

	services := Services{
		V1Service: v1Service,
		V2Service: v2Service,
	}

	return &services, nil
}
