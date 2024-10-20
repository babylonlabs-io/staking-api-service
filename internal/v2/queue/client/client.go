package v2queueclient

import (
	queueclient "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/client"
	v2queuehandler "github.com/babylonlabs-io/staking-api-service/internal/v2/queue/handler"
	queueConfig "github.com/babylonlabs-io/staking-queue-client/config"
)

type V2QueueClient struct {
	*queueclient.QueueClient
	Handler *v2queuehandler.V2QueueHandler
}

func New(cfg *queueConfig.QueueConfig, handler *v2queuehandler.V2QueueHandler, queueClient *queueclient.QueueClient) *V2QueueClient {
	return &V2QueueClient{
		QueueClient: queueClient,
		Handler:     handler,
	}
}
