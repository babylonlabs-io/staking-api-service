package tests

import (
	"testing"
	"time"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	v1model "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
	"github.com/stretchr/testify/assert"
)

func TestSaveTimelock(t *testing.T) {
	// Inject random data
	activeStakingEvent := buildActiveStakingEvent(t, 1)
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	sendTestMessage(testServer.Queues.V1QueueClient.ActiveStakingQueueClient, activeStakingEvent)

	// Wait for 2 seconds to make sure the message is processed
	time.Sleep(2 * time.Second)
	// Check from DB if the data is saved
	results, err := testutils.InspectDbDocuments[v1model.TimeLockDocument](
		testServer.Config, dbmodel.V1TimeLockCollection,
	)
	if err != nil {
		t.Fatalf("Failed to inspect DB documents: %v", err)
	}
	assert.Equal(t, 1, len(results), "expected 1 document in the DB")

	// Check the data
	assert.Equal(t, activeStakingEvent[0].StakingTxHashHex, results[0].StakingTxHashHex, "expected address to be the same")

	expectedExpireHeight := activeStakingEvent[0].StakingStartHeight + activeStakingEvent[0].StakingTimeLock
	assert.Equal(t, expectedExpireHeight, results[0].ExpireHeight, "expected address to be the same")
}

func TestNotSaveExpireCheckIfAlreadyProcessed(t *testing.T) {
	// Inject random data
	activeStakingEvent := buildActiveStakingEvent(t, 1)
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	sendTestMessage(testServer.Queues.V1QueueClient.ActiveStakingQueueClient, activeStakingEvent)
	// Send again
	sendTestMessage(testServer.Queues.V1QueueClient.ActiveStakingQueueClient, activeStakingEvent)

	// Wait for 2 seconds to make sure the message is processed
	time.Sleep(5 * time.Second)
	// Check from DB if the data is saved
	results, err := testutils.InspectDbDocuments[v1model.TimeLockDocument](
		testServer.Config, dbmodel.V1TimeLockCollection,
	)
	if err != nil {
		t.Fatalf("Failed to inspect DB documents: %v", err)
	}
	assert.Equal(t, 1, len(results), "expected 1 document in the DB")

	// Check the data
	assert.Equal(t, activeStakingEvent[0].StakingTxHashHex, results[0].StakingTxHashHex, "expected address to be the same")
	expectedExpireHeight := activeStakingEvent[0].StakingStartHeight + activeStakingEvent[0].StakingTimeLock
	assert.Equal(t, expectedExpireHeight, results[0].ExpireHeight, "expected address to be the same")

	// Now, let's inject the same data but with different expireHeight
	eventWithDifferentExpireHeight := buildActiveStakingEvent(t, 1)

	sendTestMessage(testServer.Queues.V1QueueClient.ActiveStakingQueueClient, eventWithDifferentExpireHeight)
	time.Sleep(2 * time.Second)

	results, err = testutils.InspectDbDocuments[v1model.TimeLockDocument](
		testServer.Config, dbmodel.V1TimeLockCollection,
	)
	if err != nil {
		t.Fatalf("Failed to inspect DB documents: %v", err)
	}
	assert.Equal(t, 1, len(results), "expected 1 document in the DB")

	// Check the data, it shall never be updated again
	assert.Equal(t, activeStakingEvent[0].StakingTxHashHex, results[0].StakingTxHashHex, "expected address to be the same")
	assert.Equal(t, expectedExpireHeight, results[0].ExpireHeight, "expected address to be the same")
}
