package tests

import (
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
	"github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/stretchr/testify/assert"
)

func TestUnprocessableMessageShouldBeStoredInDB(t *testing.T) {
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	sendTestMessage[string](testServer.Queues.ActiveStakingQueueClient, []string{"a rubbish message"})
	// In test, we retry 3 times. (config is 2, but counting start from 0)
	time.Sleep(20 * time.Second)

	// Fetch from DB and check
	docs, err := testutils.InspectDbDocuments[model.UnprocessableMessageDocument](
		testServer.Config, model.UnprocessableMsgCollection,
	)
	if err != nil {
		t.Fatalf("Failed to inspect DB documents: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("Expected 1 unprocessable message, got %d", len(docs))
	}

	assert.Equal(t, "\"a rubbish message\"", docs[0].MessageBody)

	// Also make sure the message is not in the queue anymore
	count, err := inspectQueueMessageCount(t, testServer.Conn, client.ActiveStakingQueueName)
	if err != nil {
		t.Fatalf("Failed to inspect queue: %v", err)
	}
	assert.Equal(t, 0, count, "expected no message in the queue")
}
