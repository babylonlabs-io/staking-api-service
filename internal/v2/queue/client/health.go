package v2queueclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/babylonlabs-io/staking-queue-client/client"
)

func (q *V2QueueClient) IsConnectionHealthy() error {
	var errorMessages []string

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checkQueue := func(name string, client client.QueueClient) {
		if err := client.Ping(ctx); err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("%s is not healthy: %v", name, err))
		}
	}

	checkQueue("ActiveStakingEventQueueClient", q.ActiveStakingEventQueueClient)
	checkQueue("UnbondingEventQueueClient", q.UnbondingEventQueueClient)

	if len(errorMessages) > 0 {
		return fmt.Errorf(strings.Join(errorMessages, "; "))
	}
	return nil
}
