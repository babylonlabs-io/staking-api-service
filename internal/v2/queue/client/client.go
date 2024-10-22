package v2queueclient

import (
	queueclient "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/client"
	v2queuehandler "github.com/babylonlabs-io/staking-api-service/internal/v2/queue/handler"
	queueConfig "github.com/babylonlabs-io/staking-queue-client/config"
)

type V2QueueClient struct {
	*queueclient.Queue
	Handler *v2queuehandler.V2QueueHandler
}

func New(cfg *queueConfig.QueueConfig, handler *v2queuehandler.V2QueueHandler, queueClient *queueclient.Queue) *V2QueueClient {
	return &V2QueueClient{
		Queue:   queueClient,
		Handler: handler,
	}
}
