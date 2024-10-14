package services

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/babylonlabs-io/staking-api-service/internal/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db"
	v1db "github.com/babylonlabs-io/staking-api-service/internal/db/v1"
	v2db "github.com/babylonlabs-io/staking-api-service/internal/db/v2"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

type DbClients struct {
	V1DBClient v1db.V1DBClient
	V2DBClient v2db.V2DBClient
}

// Service layer contains the business logic and is used to interact with
// the database and other external clients (if any).
type Services struct {
	DbClients         DbClients
	Clients           *clients.Clients
	cfg               *config.Config
	params            *types.GlobalParams
	finalityProviders []types.FinalityProviderDetails
}

func New(
	ctx context.Context,
	cfg *config.Config,
	globalParams *types.GlobalParams,
	finalityProviders []types.FinalityProviderDetails,
	clients *clients.Clients,
) (*Services, error) {
	client, err := db.NewMongoClient(ctx, cfg.Db)
	if err != nil {
		return nil, err
	}
	v1dbClient, err := v1db.New(ctx, client, cfg.Db)
	v2dbClient, err := v2db.New(ctx, client, cfg.Db)
	if err != nil {
		log.Ctx(ctx).Fatal().Err(err).Msg("error while creating v1 db client")
		return nil, err
	}

	dbClients := DbClients{
		V1DBClient: v1dbClient,
		V2DBClient: v2dbClient,
	}

	return &Services{
		DbClients:         dbClients,
		Clients:           clients,
		cfg:               cfg,
		params:            globalParams,
		finalityProviders: finalityProviders,
	}, nil
}

// DoHealthCheck checks the health of the services by ping the database.
func (s *Services) DoHealthCheck(ctx context.Context) error {
	return s.DbClients.V1DBClient.Ping(ctx)
}

func (s *Services) SaveUnprocessableMessages(ctx context.Context, messageBody, receipt string) *types.Error {
	err := s.DbClients.V1DBClient.SaveUnprocessableMessage(ctx, messageBody, receipt)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while saving unprocessable message")
		return types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "error while saving unprocessable message")
	}
	return nil
}
