package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/babylonlabs-io/staking-api-service/internal/api"
	"github.com/babylonlabs-io/staking-api-service/internal/api/handlers"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	v1model "github.com/babylonlabs-io/staking-api-service/internal/db/model/v1"
	"github.com/babylonlabs-io/staking-api-service/internal/services"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
	testmock "github.com/babylonlabs-io/staking-api-service/tests/mocks"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
)

const (
	unbondingEligibilityPath = "/v1/unbonding/eligibility"
	unbondingPath            = "/v1/unbonding"
)

func TestUnbondingRequest(t *testing.T) {
	activeStakingEvent := getTestActiveStakingEvent()
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	err := sendTestMessage(testServer.Queues.ActiveStakingQueueClient, []client.ActiveStakingEvent{*activeStakingEvent})
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	eligibilityUrl := testServer.Server.URL + unbondingEligibilityPath + "?staking_tx_hash_hex=" + activeStakingEvent.StakingTxHashHex

	// Make a GET request to the unbonding eligibility check endpoint again
	resp, err := http.Get(eligibilityUrl)
	assert.NoError(t, err, "making GET request to unbonding eligibility check endpoint should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Let's make a POST request to the unbonding endpoint
	unbondingUrl := testServer.Server.URL + unbondingPath
	requestBody := getTestUnbondDelegationRequestPayload(activeStakingEvent.StakingTxHashHex)
	requestBodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err, "marshalling request body should not fail")

	resp, err = http.Post(unbondingUrl, "application/json", bytes.NewReader(requestBodyBytes))
	assert.NoError(t, err, "making POST request to unbonding endpoint should not fail")
	defer resp.Body.Close()

	// Make a GET request to the unbonding eligibility check endpoint again
	resp, err = http.Get(eligibilityUrl)
	assert.NoError(t, err, "making GET request to unbonding eligibility check endpoint should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 403 Forbidden
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "expected HTTP 403 Forbidden status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var response api.ErrorResponse
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err, "unmarshalling response body should not fail")
	assert.Equal(t, "FORBIDDEN", response.ErrorCode, "expected error code to be FORBIDDEN")
	assert.Equal(t, "delegation state is not active", response.Message, "expected error message to be 'delegation state is not active'")

	// Let's make a POST request to the unbonding endpoint again
	resp, err = http.Post(unbondingUrl, "application/json", bytes.NewReader(requestBodyBytes))
	assert.NoError(t, err, "making POST request to unbonding endpoint should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 403 Forbidden
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "expected HTTP 403 Forbidden status")

	// Read the response body
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err, "unmarshalling response body should not fail")
	assert.Equal(t, "FORBIDDEN", response.ErrorCode, "expected error code to be FORBIDDEN")
	assert.Equal(t, "delegation state is not active", response.Message, "expected error message to be 'no active delegation found for unbonding request'")

	// The state should be updated to UnbondingRequested
	getStakerDelegationUrl := testServer.Server.URL + stakerDelegations + "?staker_btc_pk=" + activeStakingEvent.StakerPkHex
	resp, err = http.Get(getStakerDelegationUrl)
	assert.NoError(t, err, "making GET request to delegations by staker pk should not fail")
	defer resp.Body.Close()
	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var getStakerDelegationResponse handlers.PublicResponse[[]services.DelegationPublic]
	err = json.Unmarshal(bodyBytes, &getStakerDelegationResponse)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	// Check that the response body is as expected
	assert.Equal(t, activeStakingEvent.StakerPkHex, getStakerDelegationResponse.Data[0].StakerPkHex, "expected response body to match")
	assert.Equal(t, types.UnbondingRequested.ToString(), getStakerDelegationResponse.Data[0].State, "state should be unbonding requested")

	// Let's inspect what's stored in the database
	results, err := testutils.InspectDbDocuments[v1model.UnbondingDocument](
		testServer.Config, model.V1UnbondingCollection,
	)
	assert.NoError(t, err, "failed to inspect DB documents")

	assert.Equal(t, 1, len(results), "expected 1 document in the DB")
	assert.Equal(t, activeStakingEvent.StakerPkHex, results[0].StakerPkHex)
	assert.Equal(t, activeStakingEvent.FinalityProviderPkHex, results[0].FinalityPkHex)
	assert.Equal(t, requestBody.StakerSignedSignatureHex, results[0].UnbondingTxSigHex)
	assert.Equal(t, "INSERTED", results[0].State)
	assert.Equal(t, activeStakingEvent.StakingTxHashHex, results[0].StakingTxHashHex)
	assert.Equal(t, requestBody.UnbondingTxHashHex, results[0].UnbondingTxHashHex)
	assert.Equal(t, requestBody.UnbondingTxHex, results[0].UnbondingTxHex)
	assert.Equal(t, activeStakingEvent.StakingTxHex, results[0].StakingTxHex)
	assert.Equal(t, activeStakingEvent.StakingOutputIndex, results[0].StakingOutputIndex)
	assert.Equal(t, activeStakingEvent.StakingTimeLock, results[0].StakingTimelock)
	assert.Equal(t, activeStakingEvent.StakingValue, results[0].StakingAmount)
}

func TestUnbondingRequestEligibilityWhenNoMatchingDelegation(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	activeStakingEvent := testutils.GenerateRandomActiveStakingEvents(r, &testutils.TestActiveEventGeneratorOpts{
		NumOfEvents:       1,
		FinalityProviders: testutils.GeneratePks(1),
		Stakers:           testutils.GeneratePks(1),
	})
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	eligibilityUrl := testServer.Server.URL + unbondingEligibilityPath + "?staking_tx_hash_hex=" + activeStakingEvent[0].StakingTxHashHex

	// Make a GET request to the unbonding eligibility check endpoint
	resp, err := http.Get(eligibilityUrl)
	assert.NoError(t, err, "making GET request to unbonding eligibility check endpoint should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 403 Forbidden
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "expected HTTP 403 Forbidden status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var response api.ErrorResponse
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err, "unmarshalling response body should not fail")
	assert.Equal(t, "NOT_FOUND", response.ErrorCode, "expected error code to be NOT_FOUND")
}

func getTestActiveStakingEvent() *client.ActiveStakingEvent {
	return &client.ActiveStakingEvent{
		EventType:             client.ActiveStakingEventType,
		StakingTxHashHex:      "06cb98777a9cc5d3556498f6e5607947759e60f01e1441b6ce375174b8588037",
		StakerPkHex:           "91fef91f0f00d010d6c918acb9197db6cd33d99ee617090fa8cb61fa2a31405a",
		FinalityProviderPkHex: "03d5a0bb72d71993e435d6c5a70e2aa4db500a62cfaae33c56050deefee64ec0",
		StakingValue:          100000,
		StakingStartTimestamp: 1714035159,
		StakingStartHeight:    102,
		StakingTimeLock:       100,
		StakingOutputIndex:    1,
		StakingTxHex:          "0100000000010103f23a34f828a8335be0ce0400b0f1b8a5a26fd04b94abcff2ce59443b373fff0000000000ffffffff030000000000000000496a47010203040091fef91f0f00d010d6c918acb9197db6cd33d99ee617090fa8cb61fa2a31405a03d5a0bb72d71993e435d6c5a70e2aa4db500a62cfaae33c56050deefee64ec00064a08601000000000022512072a6ef79c17676fb1b54a157a4921fa57f4295dae778a423523a378171b09f3e756a042a01000000160014257cf22a8f4502076820609a30d7370856c342c40247304402203fd7b14cd32f7640c8575fe7516b2b3233c9eb3b956fe9d26bb67c156287787d0220093a1a2cd9276b53f60435fbbc55e9c007f607987c1d62f259f14ddd14a5cab501210291fef91f0f00d010d6c918acb9197db6cd33d99ee617090fa8cb61fa2a31405a00000000",
		IsOverflow:            false,
	}
}

func getTestUnbondDelegationRequestPayload(stakingTxHashHex string) handlers.UnbondDelegationRequestPayload {
	return handlers.UnbondDelegationRequestPayload{
		StakingTxHashHex:         stakingTxHashHex,
		UnbondingTxHashHex:       "17aad3b01a9fa134f0e6374b9ad1b049376a9bdbd889afc369dcb8074bbdf1b3",
		UnbondingTxHex:           "02000000000101378058b8745137ceb641141ef0609e75477960e5f6986455d3c59c7a7798cb060100000000ffffffff01905f010000000000225120ba296d88e83fd2faf5864ddda74ac8d0df6a7d76b39d5ef4c0751117a26cfd3104403c3d32c844ff751de59190ffc57427794450ba97b0d6ead43d53d865b34021abb52a6370382cfc46416f42f28d32704e504f84720d8458b5d51fee69d28cf94940728ab06ac8ab2f14b4c60cacb0e8932760f1559df93dd9053327e72cec53fc9749d5abaa4a96758b7f3588b3ad80e4c157233f1e689a37bea3ccaa9e50ba153ece2091fef91f0f00d010d6c918acb9197db6cd33d99ee617090fa8cb61fa2a31405aad2057349e985e742d5131e1e2b227b5170f6350ac2e2feb72254fcc25b3cee21a18ac2059d3532148a597a2d05c0395bf5f7176044b1cd312f37701a9b4d0aad70bc5a4ba20a5c60c2188e833d39d0fa798ab3f69aa12ed3dd2f3bad659effa252782de3c31ba20c8ccb03c379e452f10c81232b41a1ca8b63d0baf8387e57d302c987e5abb8527ba20ffeaec52a9b407b355ef6967a7ffc15fd6c3fe07de2844d61550475e7a5233e5ba539c61c150929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0957294d847c7353393803c6a2c03d6b777bb2cbb69aa4c85291d2d9bb981bb59cc2f4c8bba33b96629a6e942aad0c9815a4e5e7322f63e521e762371b71e3c1300000000",
		StakerSignedSignatureHex: "728ab06ac8ab2f14b4c60cacb0e8932760f1559df93dd9053327e72cec53fc9749d5abaa4a96758b7f3588b3ad80e4c157233f1e689a37bea3ccaa9e50ba153e",
	}
}

func TestProcessUnbondingStakingEvent(t *testing.T) {
	activeStakingEvent := getTestActiveStakingEvent()
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	err := sendTestMessage(testServer.Queues.ActiveStakingQueueClient, []*client.ActiveStakingEvent{activeStakingEvent})
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// Let's make a POST request to the unbonding endpoint
	unbondingUrl := testServer.Server.URL + unbondingPath
	requestBody := getTestUnbondDelegationRequestPayload(activeStakingEvent.StakingTxHashHex)
	requestBodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err, "marshalling request body should not fail")

	resp, err := http.Post(unbondingUrl, "application/json", bytes.NewReader(requestBodyBytes))
	assert.NoError(t, err, "making POST request to unbonding endpoint should not fail")
	defer resp.Body.Close()

	// Let's inspect what's stored in the database
	results, err := testutils.InspectDbDocuments[v1model.UnbondingDocument](
		testServer.Config, model.V1UnbondingCollection,
	)
	assert.NoError(t, err, "failed to inspect DB documents")

	assert.Equal(t, 1, len(results), "expected 1 document in the DB")
	assert.Equal(t, "INSERTED", results[0].State)
	assert.Equal(t, activeStakingEvent.StakingTxHex, results[0].StakingTxHex)

	// Let's send an unbonding event
	unbondingEvent := client.UnbondingStakingEvent{
		EventType:               client.UnbondingStakingEventType,
		StakingTxHashHex:        requestBody.StakingTxHashHex,
		UnbondingTxHashHex:      requestBody.UnbondingTxHashHex,
		UnbondingTxHex:          requestBody.UnbondingTxHex,
		UnbondingTimeLock:       10,
		UnbondingStartTimestamp: time.Now().Unix(),
		UnbondingStartHeight:    activeStakingEvent.StakingStartHeight + 100,
		UnbondingOutputIndex:    1,
	}

	sendTestMessage(testServer.Queues.UnbondingStakingQueueClient, []client.UnbondingStakingEvent{unbondingEvent})
	time.Sleep(2 * time.Second)

	// Let's GET the delegation from API
	getStakerDelegationUrl := testServer.Server.URL + stakerDelegations + "?staker_btc_pk=" + activeStakingEvent.StakerPkHex
	resp, err = http.Get(getStakerDelegationUrl)
	assert.NoError(t, err, "making GET request to delegations by staker pk should not fail")
	defer resp.Body.Close()
	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var getStakerDelegationResponse handlers.PublicResponse[[]services.DelegationPublic]
	err = json.Unmarshal(bodyBytes, &getStakerDelegationResponse)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	// Check that the response body is as expected
	assert.Equal(t, 1, len(getStakerDelegationResponse.Data), "expected 1 delegation in the response")
	assert.Equal(t, activeStakingEvent.StakerPkHex, getStakerDelegationResponse.Data[0].StakerPkHex, "expected response body to match")
	assert.Equal(t, types.Unbonding.ToString(), getStakerDelegationResponse.Data[0].State, "state should be unbonding")
	// Make sure the unbonding tx exist in the response body
	assert.NotNil(t, getStakerDelegationResponse.Data[0].UnbondingTx, "expected unbonding tx to be present in the response body")
	assert.Equal(t, unbondingEvent.UnbondingTxHex, getStakerDelegationResponse.Data[0].UnbondingTx.TxHex, "expected unbonding tx to match")

	_, err = time.Parse(time.RFC3339, getStakerDelegationResponse.Data[0].UnbondingTx.StartTimestamp)
	assert.NoError(t, err, "expected timestamp to be in RFC3339 format")

	// Let's also fetch the DB to make sure the expired check is processed
	timeLockResults, err := testutils.InspectDbDocuments[v1model.TimeLockDocument](
		testServer.Config, model.V1TimeLockCollection,
	)
	assert.NoError(t, err, "failed to inspect DB documents")

	assert.Equal(t, 2, len(timeLockResults), "expected 2 document in the DB")
	// The first one is from the
	assert.Equal(t, activeStakingEvent.StakingTxHashHex, timeLockResults[0].StakingTxHashHex)
	assert.Equal(t, types.ActiveTxType.ToString(), timeLockResults[0].TxType)
	// Point to the same staking tx hash
	assert.Equal(t, activeStakingEvent.StakingTxHashHex, timeLockResults[1].StakingTxHashHex)
	assert.Equal(t, requestBody.StakingTxHashHex, timeLockResults[1].StakingTxHashHex)
	assert.Equal(t, types.UnbondingTxType.ToString(), timeLockResults[1].TxType)
}

func TestProcessUnbondingStakingEventDuringBootstrap(t *testing.T) {
	activeStakingEvent := getTestActiveStakingEvent()
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	err := sendTestMessage(testServer.Queues.ActiveStakingQueueClient, []*client.ActiveStakingEvent{activeStakingEvent})
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// We generate the necessary unbonding request payload, but we not sending it to the unbonding endpoint
	// Instead, we send the unbonding event directly to simulate the case where our system is bootstrapping
	// which means we have the unbonding data but not the unbonding request
	requestBody := getTestUnbondDelegationRequestPayload(activeStakingEvent.StakingTxHashHex)
	unbondingEvent := client.UnbondingStakingEvent{
		EventType:               client.UnbondingStakingEventType,
		StakingTxHashHex:        requestBody.StakingTxHashHex,
		UnbondingTxHashHex:      requestBody.UnbondingTxHashHex,
		UnbondingTxHex:          requestBody.UnbondingTxHex,
		UnbondingTimeLock:       10,
		UnbondingStartTimestamp: time.Now().Unix(),
		UnbondingStartHeight:    activeStakingEvent.StakingStartHeight + 100,
		UnbondingOutputIndex:    1,
	}

	sendTestMessage(testServer.Queues.UnbondingStakingQueueClient, []client.UnbondingStakingEvent{unbondingEvent})
	time.Sleep(2 * time.Second)

	// Let's GET the delegation from API
	getStakerDelegationUrl := testServer.Server.URL + stakerDelegations + "?staker_btc_pk=" + activeStakingEvent.StakerPkHex
	resp, err := http.Get(getStakerDelegationUrl)
	assert.NoError(t, err, "making GET request to delegations by staker pk should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var getStakerDelegationResponse handlers.PublicResponse[[]services.DelegationPublic]
	err = json.Unmarshal(bodyBytes, &getStakerDelegationResponse)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	// Check that the response body is as expected
	assert.Equal(t, 1, len(getStakerDelegationResponse.Data), "expected 1 delegation in the response")
	assert.Equal(t, activeStakingEvent.StakerPkHex, getStakerDelegationResponse.Data[0].StakerPkHex, "expected response body to match")
	assert.Equal(t, types.Unbonding.ToString(), getStakerDelegationResponse.Data[0].State, "state should be unbonding")
	// Make sure the unbonding tx exist in the response body
	assert.NotNil(t, getStakerDelegationResponse.Data[0].UnbondingTx, "expected unbonding tx to be present in the response body")
	assert.Equal(t, unbondingEvent.UnbondingTxHex, getStakerDelegationResponse.Data[0].UnbondingTx.TxHex, "expected unbonding tx to match")

	// Let's also fetch the DB to make sure the expired check is processed
	timeLockResults, err := testutils.InspectDbDocuments[v1model.TimeLockDocument](
		testServer.Config, model.V1TimeLockCollection,
	)
	assert.NoError(t, err, "failed to inspect DB documents")

	assert.Equal(t, 2, len(timeLockResults), "expected 2 document in the DB")
	// The first one is from the
	assert.Equal(t, activeStakingEvent.StakingTxHashHex, timeLockResults[0].StakingTxHashHex)
	assert.Equal(t, types.ActiveTxType.ToString(), timeLockResults[0].TxType)
	// Point to the same staking tx hash
	assert.Equal(t, activeStakingEvent.StakingTxHashHex, timeLockResults[1].StakingTxHashHex)
	assert.Equal(t, requestBody.StakingTxHashHex, timeLockResults[1].StakingTxHashHex)
	assert.Equal(t, types.UnbondingTxType.ToString(), timeLockResults[1].TxType)
}

func TestShouldIgnoreOutdatedUnbondingEvent(t *testing.T) {
	activeStakingEvent := getTestActiveStakingEvent()
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	err := sendTestMessage(testServer.Queues.ActiveStakingQueueClient, []*client.ActiveStakingEvent{activeStakingEvent})
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	requestBody := getTestUnbondDelegationRequestPayload(activeStakingEvent.StakingTxHashHex)
	unbondingEvent := client.UnbondingStakingEvent{
		EventType:               client.UnbondingStakingEventType,
		StakingTxHashHex:        requestBody.StakingTxHashHex,
		UnbondingTxHashHex:      requestBody.UnbondingTxHashHex,
		UnbondingTxHex:          requestBody.UnbondingTxHex,
		UnbondingTimeLock:       10,
		UnbondingStartTimestamp: time.Now().Unix(),
		UnbondingStartHeight:    activeStakingEvent.StakingStartHeight + 100,
		UnbondingOutputIndex:    1,
	}

	sendTestMessage(testServer.Queues.UnbondingStakingQueueClient, []client.UnbondingStakingEvent{unbondingEvent})
	time.Sleep(2 * time.Second)

	// Let's GET the delegation from API
	getStakerDelegationUrl := testServer.Server.URL + stakerDelegations + "?staker_btc_pk=" + activeStakingEvent.StakerPkHex
	resp, err := http.Get(getStakerDelegationUrl)
	assert.NoError(t, err, "making GET request to delegations by staker pk should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var getStakerDelegationResponse handlers.PublicResponse[[]services.DelegationPublic]
	err = json.Unmarshal(bodyBytes, &getStakerDelegationResponse)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	// Check that the response body is as expected
	assert.Equal(t, 1, len(getStakerDelegationResponse.Data), "expected 1 delegation in the response")
	assert.Equal(t, activeStakingEvent.StakerPkHex, getStakerDelegationResponse.Data[0].StakerPkHex, "expected response body to match")
	assert.Equal(t, types.Unbonding.ToString(), getStakerDelegationResponse.Data[0].State, "state should be unbonding")
	// Make sure the unbonding tx exist in the response body
	assert.NotNil(t, getStakerDelegationResponse.Data[0].UnbondingTx, "expected unbonding tx to be present in the response body")
	assert.Equal(t, unbondingEvent.UnbondingTxHex, getStakerDelegationResponse.Data[0].UnbondingTx.TxHex, "expected unbonding tx to match")

	// Let's also fetch the DB to make sure the expired check is processed
	timeLockResults, err := testutils.InspectDbDocuments[v1model.TimeLockDocument](
		testServer.Config, model.V1TimeLockCollection,
	)
	assert.NoError(t, err, "failed to inspect DB documents")

	assert.Equal(t, 2, len(timeLockResults), "expected 2 document in the DB")
	// The first one is from the
	assert.Equal(t, activeStakingEvent.StakingTxHashHex, timeLockResults[0].StakingTxHashHex)
	assert.Equal(t, types.ActiveTxType.ToString(), timeLockResults[0].TxType)
	// Point to the same staking tx hash
	assert.Equal(t, activeStakingEvent.StakingTxHashHex, timeLockResults[1].StakingTxHashHex)
	assert.Equal(t, requestBody.StakingTxHashHex, timeLockResults[1].StakingTxHashHex)
	assert.Equal(t, types.UnbondingTxType.ToString(), timeLockResults[1].TxType)

	// Let's send an outdated unbonding event
	sendTestMessage(testServer.Queues.UnbondingStakingQueueClient, []client.UnbondingStakingEvent{unbondingEvent})
	time.Sleep(2 * time.Second)

	// Fetch from the expire checker to make sure we only processed the unbonding event once
	timeLockResults, err = testutils.InspectDbDocuments[v1model.TimeLockDocument](
		testServer.Config, model.V1TimeLockCollection,
	)
	assert.NoError(t, err, "failed to inspect DB documents")

	// Should still be 2
	assert.Equal(t, 2, len(timeLockResults), "expected 2 document in the DB")
}

func TestProcessUnbondingStakingEventShouldTolerateEventMsgOutOfOrder(t *testing.T) {
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	// We generate the active event, but not sending it yet. we will send it after sending the unbonding event
	activeStakingEvent := getTestActiveStakingEvent()
	requestBody := getTestUnbondDelegationRequestPayload(activeStakingEvent.StakingTxHashHex)
	unbondingEvent := client.UnbondingStakingEvent{
		EventType:               client.UnbondingStakingEventType,
		StakingTxHashHex:        requestBody.StakingTxHashHex,
		UnbondingTxHashHex:      requestBody.UnbondingTxHashHex,
		UnbondingTxHex:          requestBody.UnbondingTxHex,
		UnbondingTimeLock:       10,
		UnbondingStartTimestamp: time.Now().Unix(),
		UnbondingStartHeight:    activeStakingEvent.StakingStartHeight + 100,
		UnbondingOutputIndex:    1,
	}

	sendTestMessage(testServer.Queues.UnbondingStakingQueueClient, []client.UnbondingStakingEvent{unbondingEvent})
	time.Sleep(2 * time.Second)
	// Check DB, there should be no unbonding document
	results, err := testutils.InspectDbDocuments[v1model.UnbondingDocument](
		testServer.Config, model.V1UnbondingCollection,
	)
	assert.NoError(t, err, "failed to inspect DB documents")
	assert.Empty(t, results, "expected no unbonding document in the DB")

	// Send the active event
	err = sendTestMessage(testServer.Queues.ActiveStakingQueueClient, []*client.ActiveStakingEvent{activeStakingEvent})
	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	// Let's GET the delegation from API
	getStakerDelegationUrl := testServer.Server.URL + stakerDelegations + "?staker_btc_pk=" + activeStakingEvent.StakerPkHex
	resp, err := http.Get(getStakerDelegationUrl)
	assert.NoError(t, err, "making GET request to delegations by staker pk should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var getStakerDelegationResponse handlers.PublicResponse[[]services.DelegationPublic]
	err = json.Unmarshal(bodyBytes, &getStakerDelegationResponse)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	// Check that the response body is as expected
	assert.Equal(t, 1, len(getStakerDelegationResponse.Data), "expected 1 delegation in the response")
	assert.Equal(t, types.Unbonding.ToString(), getStakerDelegationResponse.Data[0].State, "state should be unbonding")
	// Make sure the unbonding tx exist in the response body
	assert.NotNil(t, getStakerDelegationResponse.Data[0].UnbondingTx, "expected unbonding tx to be present in the response body")
	assert.Equal(t, unbondingEvent.UnbondingTxHex, getStakerDelegationResponse.Data[0].UnbondingTx.TxHex, "expected unbonding tx to match")

	// Let's also fetch the DB to make sure the expired check is processed
	timeLockResults, err := testutils.InspectDbDocuments[v1model.TimeLockDocument](
		testServer.Config, model.V1TimeLockCollection,
	)
	assert.NoError(t, err, "failed to inspect DB documents")

	assert.Equal(t, 2, len(timeLockResults), "expected 2 document in the DB")
	// The first one is from the
	assert.Equal(t, activeStakingEvent.StakingTxHashHex, timeLockResults[0].StakingTxHashHex)
	assert.Equal(t, types.ActiveTxType.ToString(), timeLockResults[0].TxType)
	// Point to the same staking tx hash
	assert.Equal(t, activeStakingEvent.StakingTxHashHex, timeLockResults[1].StakingTxHashHex)
	assert.Equal(t, requestBody.StakingTxHashHex, timeLockResults[1].StakingTxHashHex)
	assert.Equal(t, types.UnbondingTxType.ToString(), timeLockResults[1].TxType)
}

func TestUnbondingRequestValidation(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	mockDB := new(testmock.V1DBClient)
	mockDB.On("FindDelegationByTxHashHex", mock.Anything, mock.Anything).Return(
		&v1model.DelegationDocument{
			State: "active",
			StakingTx: &v1model.TimelockTransaction{
				StartHeight: 300,
			},
		},
		nil,
	)
	testServer := setupTestServer(t, &TestServerDependency{MockDbClients: services.DbClients{
		V1DBClient: mockDB,
	}})
	defer testServer.Close()
	unbondingUrl := testServer.Server.URL + unbondingPath

	// TODO: Convert this to a table driven test, and test for all validations.
	// Test case 1: Hash from unbonding request does not match the hash on unbonding tranactions
	_, stakingTxHashHex := testutils.RandomBytes(r, 32)
	_, unbondingTxHashHex := testutils.RandomBytes(r, 32)
	_, unbondingTxHex, _ := testutils.GenerateRandomTx(
		r,
		&struct{ DisableRbf bool }{DisableRbf: true},
	)
	_, fakeSig := testutils.RandomBytes(r, 64)

	payload := handlers.UnbondDelegationRequestPayload{
		StakingTxHashHex: stakingTxHashHex,
		// unbondingTxHashHex does not match the hash of unbondingTxHex as it is
		// generated randomly
		UnbondingTxHashHex:       unbondingTxHashHex,
		UnbondingTxHex:           unbondingTxHex,
		StakerSignedSignatureHex: fakeSig,
	}

	requestBodyBytes, err := json.Marshal(payload)
	assert.NoError(t, err, "marshalling request body should not fail")

	resp, err := http.Post(unbondingUrl, "application/json", bytes.NewReader(requestBodyBytes))
	assert.NoError(t, err, "making POST request to unbonding endpoint should not fail")
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var unbondingResponse api.ErrorResponse
	err = json.Unmarshal(bodyBytes, &unbondingResponse)
	assert.NoError(t, err, "unmarshalling response body should not fail")
	assert.Equal(t, types.ValidationError.String(), unbondingResponse.ErrorCode)
	assert.Equal(t, "unbonding_tx_hash_hex must match the hash calculated from the provided unbonding tx", unbondingResponse.Message)

	// Test case 2: unbonding tx input enables rbf
	_, stakingTxHashHex = testutils.RandomBytes(r, 32)
	unbondingTx, unbondingTxHex, _ := testutils.GenerateRandomTx(r, nil)
	unbondingTxHashHex = unbondingTx.TxHash().String()
	_, fakeSig = testutils.RandomBytes(r, 64)

	payload = handlers.UnbondDelegationRequestPayload{
		StakingTxHashHex: stakingTxHashHex,
		// unbondingTxHashHex does not match the hash of unbondingTxHex as it is
		// generated randomly
		UnbondingTxHashHex:       unbondingTxHashHex,
		UnbondingTxHex:           unbondingTxHex,
		StakerSignedSignatureHex: fakeSig,
	}

	requestBodyBytes, err = json.Marshal(payload)
	assert.NoError(t, err, "marshalling request body should not fail")

	resp, err = http.Post(unbondingUrl, "application/json", bytes.NewReader(requestBodyBytes))
	assert.NoError(t, err, "making POST request to unbonding endpoint should not fail")
	defer resp.Body.Close()
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	err = json.Unmarshal(bodyBytes, &unbondingResponse)
	assert.NoError(t, err, "unmarshalling response body should not fail")
	assert.Equal(t, types.ValidationError.String(), unbondingResponse.ErrorCode)
	assert.Equal(t, "invalid unbonding tx hex: the unbonding tx is not a valid pre-signed unbonding tx: pre-signed tx must not be replaceable", unbondingResponse.Message)
}

func TestContentLength(t *testing.T) {
	// Setup test server with ContentLengthMiddleware
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	unbondingUrl := testServer.Server.URL + unbondingPath

	cfg, err := config.New("../config/config-test.yml")
	if err != nil {
		t.Fatal(err)
	}

	maxContentLength := cfg.Server.MaxContentLength

	// Create a payload that exceeds the max content length
	exceedingPayloadLen := maxContentLength + 1
	exceedingPayload := make([]byte, exceedingPayloadLen)
	for i := range exceedingPayload {
		exceedingPayload[i] = 'a'
	}

	// Make a POST request with the exceeding payload
	resp, err := http.Post(unbondingUrl, "application/json", bytes.NewReader(exceedingPayload))
	assert.NoError(t, err, "making POST request with exceeding payload should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 413 Request Entity Too Large
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode, "expected HTTP 413 Request Entity Too Large status")

	// Test payload exactly at the limit
	exactPayload := make([]byte, maxContentLength)
	for i := range exactPayload {
		exactPayload[i] = 'a'
	}

	resp, err = http.Post(unbondingUrl, "application/json", bytes.NewReader(exactPayload))
	assert.NoError(t, err, "making POST request with exact payload should not fail")
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusRequestEntityTooLarge, resp.StatusCode, "expected status other than HTTP 413 Request Entity Too Large")

	// Create a normal payload that's below the max content length
	activeStakingEvent := getTestActiveStakingEvent()
	normalPayload := getTestUnbondDelegationRequestPayload(activeStakingEvent.StakingTxHashHex)
	requestBodyBytes, err := json.Marshal(normalPayload)
	assert.NoError(t, err, "marshalling request body should not fail")

	resp, err = http.Post(unbondingUrl, "application/json", bytes.NewReader(requestBodyBytes))
	assert.NoError(t, err, "making POST request with normal payload should not fail")
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusRequestEntityTooLarge, resp.StatusCode, "expected status other than HTTP 413 Request Entity Too Large")
}
