package v1handlers

import (
	"context"

	handler "github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	v1service "github.com/babylonlabs-io/staking-api-service/internal/v1/service"
)

type V1Handler struct {
	*handler.Handler
	Service v1service.V1ServiceProvider
}

func New(
	ctx context.Context, handler *handler.Handler, v1Service v1service.V1ServiceProvider,
) (*V1Handler, error) {
	return &V1Handler{
		Handler: handler,
		Service: v1Service,
	}, nil
}
