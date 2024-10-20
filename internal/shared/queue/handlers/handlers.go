package queuehandlers

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/services"
	queuehandler "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/handler"
	v1queuehandler "github.com/babylonlabs-io/staking-api-service/internal/v1/queue/handler"
	v2queuehandler "github.com/babylonlabs-io/staking-api-service/internal/v2/queue/handler"
)

type QueueHandlers struct {
	V1QueueHandler *v1queuehandler.V1QueueHandler
	V2QueueHandler *v2queuehandler.V2QueueHandler
}

func New(services *services.Services, queueHandler *queuehandler.QueueHandler) (*QueueHandlers, error) {
	return &QueueHandlers{
		V1QueueHandler: v1queuehandler.New(queueHandler, services.V1Service),
	}, nil
}
