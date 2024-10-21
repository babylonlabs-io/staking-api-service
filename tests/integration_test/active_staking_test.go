package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handler "github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1service "github.com/babylonlabs-io/staking-api-service/internal/v1/service"
)

const (
	stakerDelegations = "/v1/staker/delegations"
)

func TestUnbondActiveStaking(t *testing.T) {
	activeStakingEvent := buildActiveStakingEvent(t, 1)
	expiredStakingEvent := client.NewExpiredStakingEvent(activeStakingEvent[0].StakingTxHashHex, types.ActiveTxType.ToString())
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	sendTestMessage(testServer.Queues.V1QueueClient.ActiveStakingQueueClient, activeStakingEvent)
	time.Sleep(2 * time.Second)
	sendTestMessage(testServer.Queues.V1QueueClient.ExpiredStakingQueueClient, []client.ExpiredStakingEvent{expiredStakingEvent})
	time.Sleep(2 * time.Second)

	// Test the API
	url := testServer.Server.URL + stakerDelegations + "?staker_btc_pk=" + activeStakingEvent[0].StakerPkHex
	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to delegations by staker pk should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var response handler.PublicResponse[[]v1service.DelegationPublic]
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	// Check that the response body is as expected
	assert.Equal(t, 1, len(response.Data), "expected contain 1 item in response")
	assert.Equal(t, "unbonded", response.Data[0].State)
}

func TestUnbondActiveStakingShouldTolerateOutOfOrder(t *testing.T) {
	activeStakingEvent := buildActiveStakingEvent(t, 1)
	expiredStakingEvent := client.NewExpiredStakingEvent(activeStakingEvent[0].StakingTxHashHex, types.ActiveTxType.ToString())
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	sendTestMessage(testServer.Queues.V1QueueClient.ExpiredStakingQueueClient, []client.ExpiredStakingEvent{expiredStakingEvent})
	time.Sleep(2 * time.Second)
	sendTestMessage(testServer.Queues.V1QueueClient.ActiveStakingQueueClient, activeStakingEvent)
	time.Sleep(10 * time.Second)

	// Test the API
	url := testServer.Server.URL + stakerDelegations + "?staker_btc_pk=" + activeStakingEvent[0].StakerPkHex
	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to delegations by staker pk should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var response handler.PublicResponse[[]v1service.DelegationPublic]
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	// Check that the response body is as expected
	assert.Equal(t, 1, len(response.Data), "expected contain 1 item in response")
	assert.Equal(t, "unbonded", response.Data[0].State)
}

func TestShouldNotUnbondIfNotActiveState(t *testing.T) {
	activeStakingEvent := getTestActiveStakingEvent()
	expiredStakingEvent := client.NewExpiredStakingEvent(activeStakingEvent.StakingTxHashHex, types.ActiveTxType.ToString())
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	err := sendTestMessage(testServer.Queues.V1QueueClient.ActiveStakingQueueClient, []client.ActiveStakingEvent{*activeStakingEvent})
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	// Let's make a POST request to the unbonding endpoint to change the state to unbonding_requested
	unbondingUrl := testServer.Server.URL + unbondingPath
	requestBody := getTestUnbondDelegationRequestPayload(activeStakingEvent.StakingTxHashHex)
	requestBodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err, "marshalling request body should not fail")

	resp, err := http.Post(unbondingUrl, "application/json", bytes.NewReader(requestBodyBytes))
	assert.NoError(t, err, "making POST request to unbonding endpoint should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 202
	assert.Equal(t, http.StatusAccepted, resp.StatusCode, "expected HTTP 202 Accepted status")

	sendTestMessage(testServer.Queues.V1QueueClient.ExpiredStakingQueueClient, []client.ExpiredStakingEvent{expiredStakingEvent})
	time.Sleep(2 * time.Second)

	// Test the API
	url := testServer.Server.URL + stakerDelegations + "?staker_btc_pk=" + activeStakingEvent.StakerPkHex
	resp, err = http.Get(url)
	assert.NoError(t, err, "making GET request to delegations by staker pk should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var response handler.PublicResponse[[]v1service.DelegationPublic]
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	// Check that the response body is as expected
	assert.Equal(t, 1, len(response.Data), "expected contain 1 item in response")
	// The state should not be updated to unbonded
	assert.Equal(t, "unbonding_requested", response.Data[0].State)
}
