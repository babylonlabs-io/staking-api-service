package v2queueclient

import (
	queueclient "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/client"
	"github.com/rs/zerolog/log"
)

// commented the already initialized queue clients in v1
func (q *V2QueueClient) StartReceivingMessages() {
	log.Printf("Starting to receive messages from v2 queues")
	// queueclient.StartQueueMessageProcessing(
	// 	q.ActiveStakingEventQueueClient,
	// 	q.Handler.ActiveStakingHandler, q.Handler.HandleUnprocessedMessage,
	// 	q.MaxRetryAttempts, q.ProcessingTimeout,
	// )

	log.Printf("Starting to receive messages from verified staking queue")
	queueclient.StartQueueMessageProcessing(
		q.VerifiedStakingEventQueueClient,
		q.Handler.VerifiedStakingHandler, q.Handler.HandleUnprocessedMessage,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)

	log.Printf("Starting to receive messages from pending staking queue")
	queueclient.StartQueueMessageProcessing(
		q.PendingStakingEventQueueClient,
		q.Handler.PendingStakingHandler, q.Handler.HandleUnprocessedMessage,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)

	// log.Printf("Starting to receive messages from unbonding staking queue")
	// queueclient.StartQueueMessageProcessing(
	// 	q.UnbondingEventQueueClient,
	// 	q.Handler.UnbondingStakingHandler, q.Handler.HandleUnprocessedMessage,
	// 	q.MaxRetryAttempts, q.ProcessingTimeout,
	// )

	// log.Printf("Starting to receive messages from expired staking queue")
	// queueclient.StartQueueMessageProcessing(
	// 	q.StakingExpiredEventQueueClient,
	// 	q.Handler.ExpiredStakingHandler, q.Handler.HandleUnprocessedMessage,
	// 	q.MaxRetryAttempts, q.ProcessingTimeout,
	// )

}

// Turn off all message processing
func (q *V2QueueClient) StopReceivingMessages() {
	log.Printf("Stopping to receive messages from v2 queues")

	// log.Printf("Stopping to receive messages from active staking queue")
	// activeQueueErr := q.ActiveStakingEventQueueClient.Stop()
	// if activeQueueErr != nil {
	// 	log.Error().Err(activeQueueErr).
	// 		Str("queueName", q.ActiveStakingEventQueueClient.GetQueueName()).
	// 		Msg("error while stopping queue")
	// }

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

	// log.Printf("Stopping to receive messages from unbonding staking queue")
	// unbondingQueueErr := q.UnbondingEventQueueClient.Stop()
	// if unbondingQueueErr != nil {
	// 	log.Error().Err(unbondingQueueErr).
	// 		Str("queueName", q.UnbondingEventQueueClient.GetQueueName()).
	// 		Msg("error while stopping queue")
	// }

	// log.Printf("Stopping to receive messages from expired staking queue")
	// expiredQueueErr := q.StakingExpiredEventQueueClient.Stop()
	// if expiredQueueErr != nil {
	// 	log.Error().Err(expiredQueueErr).
	// 		Str("queueName", q.StakingExpiredEventQueueClient.GetQueueName()).
	// 		Msg("error while stopping queue")
	// }
}
