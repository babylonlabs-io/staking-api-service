package v1queueclient

import (
	queueclient "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/client"
	"github.com/rs/zerolog/log"
)

func (q *V1QueueClient) StartReceivingMessages() {
	log.Printf("Starting to receive messages from v1 queues")
	// start processing messages from the active staking queue
	queueclient.StartQueueMessageProcessing(
		q.ActiveStakingQueueClient,
		q.Handler.ActiveStakingHandler, q.Handler.HandleUnprocessedMessage,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)
	log.Printf("Starting to receive messages from expired staking queue")
	queueclient.StartQueueMessageProcessing(
		q.ExpiredStakingQueueClient,
		q.Handler.ExpiredStakingHandler, q.Handler.HandleUnprocessedMessage,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)
	log.Printf("Starting to receive messages from unbonding staking queue")
	queueclient.StartQueueMessageProcessing(
		q.UnbondingStakingQueueClient,
		q.Handler.UnbondingStakingHandler, q.Handler.HandleUnprocessedMessage,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)
	log.Printf("Starting to receive messages from withdraw staking queue")
	queueclient.StartQueueMessageProcessing(
		q.WithdrawStakingQueueClient,
		q.Handler.WithdrawStakingHandler, q.Handler.HandleUnprocessedMessage,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)
	log.Printf("Starting to receive messages from stats queue")
	queueclient.StartQueueMessageProcessing(
		q.StatsQueueClient,
		q.Handler.StatsHandler, q.Handler.HandleUnprocessedMessage,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)
	log.Printf("Starting to receive messages from btc info queue")
	queueclient.StartQueueMessageProcessing(
		q.BtcInfoQueueClient,
		q.Handler.BtcInfoHandler, q.Handler.HandleUnprocessedMessage,
		q.MaxRetryAttempts, q.ProcessingTimeout,
	)
	// ...add more queues here
}

// Turn off all message processing
func (q *V1QueueClient) StopReceivingMessages() {
	activeQueueErr := q.ActiveStakingQueueClient.Stop()
	if activeQueueErr != nil {
		log.Error().Err(activeQueueErr).
			Str("queueName", q.ActiveStakingQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}
	expiredQueueErr := q.ExpiredStakingQueueClient.Stop()
	if expiredQueueErr != nil {
		log.Error().Err(expiredQueueErr).
			Str("queueName", q.ExpiredStakingQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}
	unbondingQueueErr := q.UnbondingStakingQueueClient.Stop()
	if unbondingQueueErr != nil {
		log.Error().Err(unbondingQueueErr).
			Str("queueName", q.UnbondingStakingQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}
	withdrawnQueueErr := q.WithdrawStakingQueueClient.Stop()
	if withdrawnQueueErr != nil {
		log.Error().Err(withdrawnQueueErr).
			Str("queueName", q.WithdrawStakingQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}
	statsQueueErr := q.StatsQueueClient.Stop()
	if statsQueueErr != nil {
		log.Error().Err(statsQueueErr).
			Str("queueName", q.StatsQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}
	btcInfoQueueErr := q.BtcInfoQueueClient.Stop()
	if btcInfoQueueErr != nil {
		log.Error().Err(btcInfoQueueErr).
			Str("queueName", q.BtcInfoQueueClient.GetQueueName()).
			Msg("error while stopping queue")
	}
	// ...add more queues here
}
