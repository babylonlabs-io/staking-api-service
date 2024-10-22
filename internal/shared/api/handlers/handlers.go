package handlers

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services"
	v1handler "github.com/babylonlabs-io/staking-api-service/internal/v1/api/handlers"
	v2handler "github.com/babylonlabs-io/staking-api-service/internal/v2/api/handlers"
)

type Handlers struct {
	SharedHandler *handler.Handler
	V1Handler     *v1handler.V1Handler
	V2Handler     *v2handler.V2Handler
}

func New(ctx context.Context, config *config.Config, services *services.Services) (*Handlers, error) {
	sharedHandler, err := handler.New(ctx, config, services.SharedService)
	if err != nil {
		return nil, err
	}
	v1Handler, err := v1handler.New(ctx, sharedHandler, services.V1Service)
	if err != nil {
		return nil, err
	}
	v2Handler, err := v2handler.New(ctx, sharedHandler, services.V2Service)
	if err != nil {
		return nil, err
	}

	return &Handlers{
		SharedHandler: sharedHandler,
		V1Handler:     v1Handler,
		V2Handler:     v2Handler,
	}, nil
}
