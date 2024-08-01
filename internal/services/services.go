package services

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/babylonlabs-io/staking-api-service/internal/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

// Service layer contains the business logic and is used to interact with
// the database and other external clients (if any).
type Services struct {
	DbClient          db.DBClient
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
	dbClient, err := db.New(ctx, cfg.Db)
	if err != nil {
		log.Ctx(ctx).Fatal().Err(err).Msg("error while creating db client")
		return nil, err
	}
	return &Services{
		DbClient:          dbClient,
		Clients:           clients,
		cfg:               cfg,
		params:            globalParams,
		finalityProviders: finalityProviders,
	}, nil
}

// DoHealthCheck checks the health of the services by ping the database.
func (s *Services) DoHealthCheck(ctx context.Context) error {
	return s.DbClient.Ping(ctx)
}

func (s *Services) SaveUnprocessableMessages(ctx context.Context, messageBody, receipt string) *types.Error {
	err := s.DbClient.SaveUnprocessableMessage(ctx, messageBody, receipt)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while saving unprocessable message")
		return types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "error while saving unprocessable message")
	}
	return nil
}
