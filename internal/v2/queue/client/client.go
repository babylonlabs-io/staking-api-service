package v2queueclient

import (
	queueclient "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/client"
	v2queuehandler "github.com/babylonlabs-io/staking-api-service/internal/v2/queue/handler"
	v2queueschema "github.com/babylonlabs-io/staking-api-service/internal/v2/queue/schema"
	client "github.com/babylonlabs-io/staking-queue-client/client"
	queueConfig "github.com/babylonlabs-io/staking-queue-client/config"
	"github.com/rs/zerolog/log"
)

type V2QueueClient struct {
	*queueclient.Queue
	Handler                       *v2queuehandler.V2QueueHandler
	ActiveStakingEventQueueClient client.QueueClient
	UnbondingEventQueueClient     client.QueueClient
}

func New(cfg *queueConfig.QueueConfig, handler *v2queuehandler.V2QueueHandler, queueClient *queueclient.Queue) *V2QueueClient {
	activeStakingEventQueueClient, err := client.NewQueueClient(cfg, v2queueschema.ActiveStakingQueueName)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating ActiveStakingEventQueue")
	}

	unbondingEventQueueClient, err := client.NewQueueClient(cfg, v2queueschema.UnbondingStakingQueueName)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating UnbondingEventQueue")
	}

	return &V2QueueClient{
		Queue:                         queueClient,
		Handler:                       handler,
		ActiveStakingEventQueueClient: activeStakingEventQueueClient,
		UnbondingEventQueueClient:     unbondingEventQueueClient,
	}
}
