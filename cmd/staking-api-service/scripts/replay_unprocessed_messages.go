package scripts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	v2queue "github.com/babylonlabs-io/staking-api-service/internal/v2/queue"
	queueClient "github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/rs/zerolog/log"
)

type GenericEvent struct {
	EventType queueClient.EventType `json:"event_type"`
}

func ReplayUnprocessableMessages(ctx context.Context, cfg *config.Config, queues *v2queue.Queues, db dbclient.DBClient) (err error) {
	// Fetch unprocessable messages
	unprocessableMessages, err := db.FindUnprocessableMessages(ctx)
	if err != nil {
		return errors.New("failed to retrieve unprocessable messages")
	}

	// Get the message count
	messageCount := len(unprocessableMessages)

	// Inform the user of the number of unprocessable messages
	if messageCount == 0 {
		return errors.New("no unprocessable messages to replay")
	}

	// Process each unprocessable message
	for _, msg := range unprocessableMessages {
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
	}

	log.Info().Msg("Reprocessing of unprocessable messages completed.")
	return
}

// processEventMessage processes the event message based on its EventType.
func processEventMessage(ctx context.Context, queues *v2queue.Queues, event GenericEvent, messageBody string) error {
	switch event.EventType {
	case queueClient.ActiveStakingEventType:
		return queues.ActiveStakingQueueClient.SendMessage(ctx, messageBody)
	case queueClient.UnbondingStakingEventType:
		return queues.UnbondingStakingQueueClient.SendMessage(ctx, messageBody)
	case queueClient.WithdrawableStakingEventType:
		return queues.WithdrawableStakingQueueClient.SendMessage(ctx, messageBody)
	case queueClient.WithdrawnStakingEventType:
		return queues.WithdrawnStakingQueueClient.SendMessage(ctx, messageBody)
	default:
		return fmt.Errorf("unknown event type: %v", event.EventType)
	}
}
