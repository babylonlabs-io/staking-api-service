package v2handlers

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
	v2service "github.com/babylonlabs-io/staking-api-service/internal/v2/api/service"
)

type V2Handler struct {
	handler handler.Handler
	Service v2service.V2ServiceInterface
}

func New(
	ctx context.Context, handler handler.Handler, v2Service v2service.V2ServiceInterface,
) (*V2Handler, error) {
	return &V2Handler{
		handler: handler,
		Service: v2Service,
	}, nil
}
