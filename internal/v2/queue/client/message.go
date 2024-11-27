package v2queueclient

import (
	queueclient "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/client"
	"github.com/rs/zerolog/log"
)

func (q *V2QueueClient) StartReceivingMessages() {
	log.Printf("Starting to receive messages from v2 queues")
	// start processing messages from the active staking queue
	queueclient.StartQueueMessageProcessing(
		q.ActiveStakingEventQueueClient,
		q.Handler.ActiveStakingHandler, q.Handler.HandleUnprocessedMessage,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)
	log.Printf("Starting to receive messages from unbonding staking queue")
	queueclient.StartQueueMessageProcessing(
		q.UnbondingEventQueueClient,
		q.Handler.UnbondingStakingHandler, q.Handler.HandleUnprocessedMessage,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)
	// ...add more queues here
}

func (q *V2QueueClient) StopReceivingMessages() {
	activeQueueErr := q.ActiveStakingEventQueueClient.Stop()
	if activeQueueErr != nil {
		log.Error().Err(activeQueueErr).
			Str("queueName", q.ActiveStakingEventQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}
	unbondingQueueErr := q.UnbondingEventQueueClient.Stop()
	if unbondingQueueErr != nil {
		log.Error().Err(unbondingQueueErr).
			Str("queueName", q.UnbondingEventQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}
	// ...add more queues here
}
