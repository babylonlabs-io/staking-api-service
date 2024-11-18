package queuehandlers

import (
	queuehandler "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services"
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
		V2QueueHandler: v2queuehandler.New(queueHandler, services.V2Service),
	}, nil
}
