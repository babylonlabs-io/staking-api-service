package v2queueclient

import (
	queueclient "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/client"
	"github.com/rs/zerolog/log"
)

func (q *V2QueueClient) StartReceivingMessages() {
	log.Printf("Starting to receive messages from v2 queues")

	log.Printf("Starting to receive messages from verified staking queue")
	queueclient.StartQueueMessageProcessing(
		q.VerifiedStakingEventQueueClient,
		q.Handler.VerifiedStakingHandler, nil,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)

	log.Printf("Starting to receive messages from pending staking queue")
	queueclient.StartQueueMessageProcessing(
		q.PendingStakingEventQueueClient,
		q.Handler.PendingStakingHandler, nil,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)

}

// Turn off all message processing
func (q *V2QueueClient) StopReceivingMessages() {
	log.Printf("Stopping to receive messages from v2 queues")

	log.Printf("Stopping to receive messages from verified staking queue")
	verifiedQueueErr := q.VerifiedStakingEventQueueClient.Stop()
	if verifiedQueueErr != nil {
		log.Error().Err(verifiedQueueErr).
			Str("queueName", q.VerifiedStakingEventQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}

	log.Printf("Stopping to receive messages from pending staking queue")
	pendingQueueErr := q.PendingStakingEventQueueClient.Stop()
	if pendingQueueErr != nil {
		log.Error().Err(pendingQueueErr).
			Str("queueName", q.PendingStakingEventQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}
}
