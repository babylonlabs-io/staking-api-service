package v1service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type V1Service struct {
	*service.Service
}

func New(
	ctx context.Context,
	cfg *config.Config,
	globalParams *types.GlobalParams,
	finalityProviders []types.FinalityProviderDetails,
	clients *clients.Clients,
	dbClients *dbclients.DbClients,
) (*V1Service, error) {
	service, err := service.New(ctx, cfg, globalParams, finalityProviders, clients, dbClients)
	if err != nil {
		return nil, err
	}

	return &V1Service{
		service,
	}, nil
}
