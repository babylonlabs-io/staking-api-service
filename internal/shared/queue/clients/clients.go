package queueclients

import (
	"context"

	queueclient "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/client"
	queuehandler "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/handler"
	queuehandlers "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/handlers"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services"
	v1queueclient "github.com/babylonlabs-io/staking-api-service/internal/v1/queue/client"
	v2queueclient "github.com/babylonlabs-io/staking-api-service/internal/v2/queue/client"
	queueConfig "github.com/babylonlabs-io/staking-queue-client/config"
	"github.com/rs/zerolog/log"
)

type QueueClients struct {
	V1QueueClient *v1queueclient.V1QueueClient
	V2QueueClient *v2queueclient.V2QueueClient
}

func New(ctx context.Context, cfg *queueConfig.QueueConfig, services *services.Services) *QueueClients {
	queueClient := queueclient.New(ctx, cfg, services)
	queueHandler := queuehandler.New(queueClient.StatsQueueClient.SendMessage)
	queueHandlers, err := queuehandlers.New(services, queueHandler)
	if err != nil {
		log.Fatal().Err(err).Msg("error while setting up queue handlers")
	}
	v1QueueClient := v1queueclient.New(cfg, queueHandlers.V1QueueHandler, queueClient)
	return &QueueClients{
		V1QueueClient: v1QueueClient,
	}
}

func (q *QueueClients) StartReceivingMessages() {
	log.Printf("Starting to receive messages from queue clients")
	q.V1QueueClient.StartReceivingMessages()
	q.V2QueueClient.StartReceivingMessages()
}
