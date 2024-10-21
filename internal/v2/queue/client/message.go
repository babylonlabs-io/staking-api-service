package v2queueclient

import (
	"github.com/rs/zerolog/log"
)

func (q *V2QueueClient) StartReceivingMessages() {
	log.Printf("Starting to receive messages from v2 queues")
}

// Turn off all message processing
func (q *V2QueueClient) StopReceivingMessages() {
}
