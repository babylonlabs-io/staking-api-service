package handlers

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/services"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	v1handler "github.com/babylonlabs-io/staking-api-service/internal/v1/api/handlers"
	v2handler "github.com/babylonlabs-io/staking-api-service/internal/v2/api/handlers"
)

type Handlers struct {
	V1 *v1handler.V1Handler
	V2 *v2handler.V2Handler
}

func New(ctx context.Context, config *config.Config, services *services.Services) (*Handlers, error) {
	v1Handler, err := v1handler.New(ctx, handler.Handler{Config: config}, services.V1Service)
	if err != nil {
		return nil, err
	}
	v2Handler, err := v2handler.New(ctx, handler.Handler{Config: config}, services.V2Service)
	if err != nil {
		return nil, err
	}

	return &Handlers{
		V1: v1Handler,
		V2: v2Handler,
	}, nil
}
