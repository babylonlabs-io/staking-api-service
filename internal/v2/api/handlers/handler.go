package v2handlers

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	v2service "github.com/babylonlabs-io/staking-api-service/internal/v2/service"
)

type V2Handler struct {
	*handler.Handler
	Service v2service.V2ServiceProvider
}

func New(
	ctx context.Context, handler *handler.Handler, v2Service v2service.V2ServiceProvider,
) (*V2Handler, error) {
	return &V2Handler{
		Handler: handler,
		Service: v2Service,
	}, nil
}
