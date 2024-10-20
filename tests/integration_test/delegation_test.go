package tests

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/stretchr/testify/assert"

	handler "github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1service "github.com/babylonlabs-io/staking-api-service/internal/v1/api/service"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
)

const (
	delegationRouter = "/v1/delegation"
)

func TestGetDelegationByTxHashHex(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	activeStakingEvent := testutils.GenerateRandomActiveStakingEvents(
		r,
		&testutils.TestActiveEventGeneratorOpts{
			NumOfEvents:       1,
			FinalityProviders: testutils.GeneratePks(1),
			Stakers:           testutils.GeneratePks(1),
		},
	)

	expiredStakingEvent := client.NewExpiredStakingEvent(activeStakingEvent[0].StakingTxHashHex, types.ActiveTxType.ToString())
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	sendTestMessage(testServer.Queues.V1QueueClient.ActiveStakingQueueClient, activeStakingEvent)
	time.Sleep(2 * time.Second)
	sendTestMessage(testServer.Queues.V1QueueClient.ExpiredStakingQueueClient, []client.ExpiredStakingEvent{expiredStakingEvent})
	time.Sleep(2 * time.Second)

	// Test the API
	url := testServer.Server.URL + delegationRouter + "?staking_tx_hash_hex=" + activeStakingEvent[0].StakingTxHashHex
	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to delegation by tx hash should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var response handler.PublicResponse[v1service.DelegationPublic]
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	// Check that the response body is as expected
	assert.Equal(t, "unbonded", response.Data.State)
	assert.Equal(t, activeStakingEvent[0].StakingTxHashHex, response.Data.StakingTxHashHex)
}
