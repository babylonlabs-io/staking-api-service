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
	Handler                         *v2queuehandler.V2QueueHandler
	ActiveStakingEventQueueClient   client.QueueClient
	StakingExpiredEventQueueClient  client.QueueClient
	UnbondingEventQueueClient       client.QueueClient
	PendingStakingEventQueueClient  client.QueueClient
	VerifiedStakingEventQueueClient client.QueueClient
}

func New(cfg *queueConfig.QueueConfig, handler *v2queuehandler.V2QueueHandler, queueClient *queueclient.Queue) *V2QueueClient {
	activeStakingEventQueueClient, err := client.NewQueueClient(cfg, v2queueschema.ActiveStakingQueueName)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating ActiveStakingEventQueue")
	}

	stakingExpiredEventQueueClient, err := client.NewQueueClient(cfg, v2queueschema.ExpiredStakingQueueName)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating StakingExpiredEventQueue")
	}

	unbondingEventQueueClient, err := client.NewQueueClient(cfg, v2queueschema.UnbondingStakingQueueName)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating UnbondingEventQueue")
	}

	pendingStakingEventQueueClient, err := client.NewQueueClient(cfg, v2queueschema.PendingStakingQueueName)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating PendingStakingEventQueue")
	}

	verifiedStakingEventQueueClient, err := client.NewQueueClient(cfg, v2queueschema.VerifiedStakingQueueName)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating VerifiedStakingEventQueue")
	}

	return &V2QueueClient{
		Queue:                           queueClient,
		Handler:                         handler,
		ActiveStakingEventQueueClient:   activeStakingEventQueueClient,
		StakingExpiredEventQueueClient:  stakingExpiredEventQueueClient,
		UnbondingEventQueueClient:       unbondingEventQueueClient,
		PendingStakingEventQueueClient:  pendingStakingEventQueueClient,
		VerifiedStakingEventQueueClient: verifiedStakingEventQueueClient,
	}
}
