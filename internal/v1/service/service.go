package v1service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
)

type V1Service struct {
	*service.Service
}

func New(
	ctx context.Context,
	sharedService *service.Service,
) (*V1Service, error) {
	return &V1Service{
		Service: sharedService,
	}, nil
}
