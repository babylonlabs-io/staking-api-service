package v2queuehandler

import (
	queuehandler "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/handler"
	v2service "github.com/babylonlabs-io/staking-api-service/internal/v2/service"
)

type V2QueueHandler struct {
	*queuehandler.QueueHandler
	Service v2service.V2ServiceProvider
}

func New(queueHandler *queuehandler.QueueHandler, service v2service.V2ServiceProvider) *V2QueueHandler {
	return &V2QueueHandler{
		QueueHandler: queueHandler,
		Service:      service,
	}
}
