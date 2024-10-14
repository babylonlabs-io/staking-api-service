package scripts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db"
	"github.com/babylonlabs-io/staking-api-service/internal/queue"
	queueClient "github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/rs/zerolog/log"
)

type GenericEvent struct {
	EventType queueClient.EventType `json:"event_type"`
}

func ReplayUnprocessableMessages(ctx context.Context, cfg *config.Config, queues *queue.Queues, db db.BaseDBClient) (err error) {
	fmt.Println("Starting to replay unprocessable messages...")

	// Fetch unprocessable messages
	unprocessableMessages, err := db.FindUnprocessableMessages(ctx)
	if err != nil {
		return errors.New("failed to retrieve unprocessable messages")
	}

	// Get the message count
	messageCount := len(unprocessableMessages)

	// Inform the user of the number of unprocessable messages
	fmt.Printf("There are %d unprocessable messages.\n", messageCount)
	if messageCount == 0 {
		return errors.New("no unprocessable messages to replay")
	}

	// Process each unprocessable message
	for i, msg := range unprocessableMessages {
		fmt.Printf("Processing message %d/%d: %s\n", i+1, messageCount, msg.MessageBody)

		var genericEvent GenericEvent
		if err := json.Unmarshal([]byte(msg.MessageBody), &genericEvent); err != nil {
			return errors.New("failed to unmarshal event message")
		}

		// Process the event message
		if err := processEventMessage(ctx, queues, genericEvent, msg.MessageBody); err != nil {
			return errors.New("failed to process message")
		}

		// Delete the processed message from the database
		if err := db.DeleteUnprocessableMessage(ctx, msg.Receipt); err != nil {
			return errors.New("failed to delete unprocessable message")
		}

		fmt.Printf("Message %d/%d processed and deleted successfully.\n", i+1, messageCount)
	}

	log.Info().Msg("Reprocessing of unprocessable messages completed.")
	fmt.Println("Reprocessing of unprocessable messages completed.")
	return
}

// processEventMessage processes the event message based on its EventType.
func processEventMessage(ctx context.Context, queues *queue.Queues, event GenericEvent, messageBody string) error {
	fmt.Printf("Sending message to the queue for event type: %v\n", event.EventType)

	switch event.EventType {
	case queueClient.ActiveStakingEventType:
		return queues.ActiveStakingQueueClient.SendMessage(ctx, messageBody)
	case queueClient.UnbondingStakingEventType:
		return queues.UnbondingStakingQueueClient.SendMessage(ctx, messageBody)
	case queueClient.WithdrawStakingEventType:
		return queues.WithdrawStakingQueueClient.SendMessage(ctx, messageBody)
	case queueClient.ExpiredStakingEventType:
		return queues.ExpiredStakingQueueClient.SendMessage(ctx, messageBody)
	case queueClient.StatsEventType:
		return queues.StatsQueueClient.SendMessage(ctx, messageBody)
	case queueClient.BtcInfoEventType:
		return queues.BtcInfoQueueClient.SendMessage(ctx, messageBody)
	default:
		fmt.Printf("Error: unknown event type: %v\n", event.EventType)
		return fmt.Errorf("unknown event type: %v", event.EventType)
	}
}
