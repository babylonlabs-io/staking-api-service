package v1handlers

import (
	"context"

	handler "github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
	v1service "github.com/babylonlabs-io/staking-api-service/internal/v1/api/service"
)

type V1Handler struct {
	handler handler.Handler
	Service v1service.V1ServiceInterface
}

func New(
	ctx context.Context, handler handler.Handler, v1Service v1service.V1ServiceInterface,
) (*V1Handler, error) {
	return &V1Handler{
		handler: handler,
		Service: v1Service,
	}, nil
}
