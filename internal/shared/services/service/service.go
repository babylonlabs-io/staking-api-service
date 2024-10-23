package service

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// Services layer contains the business logic and is used to interact with
// the database and other external clients (if any).
type Service struct {
	DbClients         *dbclients.DbClients
	Clients           *clients.Clients
	Cfg               *config.Config
	Params            *types.GlobalParams
	FinalityProviders []types.FinalityProviderDetails
}

func New(
	ctx context.Context,
	cfg *config.Config,
	globalParams *types.GlobalParams,
	finalityProviders []types.FinalityProviderDetails,
	clients *clients.Clients,
	dbClients *dbclients.DbClients,
) (*Service, error) {
	return &Service{
		DbClients:         dbClients,
		Clients:           clients,
		Cfg:               cfg,
		Params:            globalParams,
		FinalityProviders: finalityProviders,
	}, nil
}

// DoHealthCheck checks the health of the services by ping the database.
func (s *Service) DoHealthCheck(ctx context.Context) error {
	if err := s.DbClients.SharedDBClient.Ping(ctx); err != nil {
		return err
	}
	return s.DbClients.IndexerDBClient.Ping(ctx)
}

func (s *Service) SaveUnprocessableMessages(ctx context.Context, messageBody, receipt string) *types.Error {
	err := s.DbClients.V1DBClient.SaveUnprocessableMessage(ctx, messageBody, receipt)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while saving unprocessable message")
		return types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "error while saving unprocessable message")
	}
	return nil
}
