package v1queueclient

import (
	queueclient "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/client"
	v1queuehandler "github.com/babylonlabs-io/staking-api-service/internal/v1/queue/handler"
	client "github.com/babylonlabs-io/staking-queue-client/client"
	queueConfig "github.com/babylonlabs-io/staking-queue-client/config"
	"github.com/rs/zerolog/log"
)

type V1QueueClient struct {
	*queueclient.QueueClient
	Handler                     *v1queuehandler.V1QueueHandler
	ActiveStakingQueueClient    client.QueueClient
	ExpiredStakingQueueClient   client.QueueClient
	UnbondingStakingQueueClient client.QueueClient
	WithdrawStakingQueueClient  client.QueueClient
	BtcInfoQueueClient          client.QueueClient
}

func New(cfg *queueConfig.QueueConfig, handler *v1queuehandler.V1QueueHandler, queueClient *queueclient.QueueClient) *V1QueueClient {
	activeStakingQueueClient, err := client.NewQueueClient(
		cfg, client.ActiveStakingQueueName,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating ActiveStakingQueueClient")
	}

	expiredStakingQueueClient, err := client.NewQueueClient(
		cfg, client.ExpiredStakingQueueName,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating ExpiredStakingQueueClient")
	}

	unbondingStakingQueueClient, err := client.NewQueueClient(
		cfg, client.UnbondingStakingQueueName,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating UnbondingStakingQueueClient")
	}
	withdrawStakingQueueClient, err := client.NewQueueClient(
		cfg, client.WithdrawStakingQueueName,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating WithdrawStakingQueueClient")
	}
	btcInfoQueueClient, err := client.NewQueueClient(
		cfg, client.BtcInfoQueueName,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating BtcInfoQueueClient")
	}
	return &V1QueueClient{
		QueueClient:                 queueClient,
		Handler:                     handler,
		ActiveStakingQueueClient:    activeStakingQueueClient,
		ExpiredStakingQueueClient:   expiredStakingQueueClient,
		UnbondingStakingQueueClient: unbondingStakingQueueClient,
		WithdrawStakingQueueClient:  withdrawStakingQueueClient,
		BtcInfoQueueClient:          btcInfoQueueClient,
	}
}
