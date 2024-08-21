package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/api"
	"github.com/babylonlabs-io/staking-api-service/internal/api/handlers"
	"github.com/babylonlabs-io/staking-api-service/internal/services"
	"github.com/babylonlabs-io/staking-api-service/internal/utils"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
	"github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/stretchr/testify/assert"
)

const (
	checkStakerDelegationUrl = "/v1/staker/delegation/check"
)

func FuzzTestStakerDelegationsWithPaginationResponse(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 3)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		activeStakingEventsByStaker1 := testutils.GenerateRandomActiveStakingEvents(
			r,
			&testutils.TestActiveEventGeneratorOpts{
				NumOfEvents:       11,
				FinalityProviders: testutils.GeneratePks(11),
				Stakers:           testutils.GeneratePks(1),
			},
		)
		activeStakingEventsByStaker2 := testutils.GenerateRandomActiveStakingEvents(
			r,
			&testutils.TestActiveEventGeneratorOpts{
				NumOfEvents:       11,
				FinalityProviders: testutils.GeneratePks(11),
				Stakers:           testutils.GeneratePks(1),
			},
		)
		testServer := setupTestServer(t, nil)
		defer testServer.Close()
		sendTestMessage(
			testServer.Queues.ActiveStakingQueueClient,
			append(activeStakingEventsByStaker1, activeStakingEventsByStaker2...),
		)
		time.Sleep(5 * time.Second)

		// Test the API
		stakerPk := activeStakingEventsByStaker1[0].StakerPkHex
		url := testServer.Server.URL + stakerDelegations + "?staker_btc_pk=" + stakerPk
		var paginationKey string
		var allDataCollected []services.DelegationPublic
		var atLeastOnePage bool
		for {
			resp, err := http.Get(url + "&pagination_key=" + paginationKey)
			assert.NoError(t, err, "making GET request to delegations by staker pk should not fail")
			assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")
			bodyBytes, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "reading response body should not fail")
			var response handlers.PublicResponse[[]services.DelegationPublic]
			err = json.Unmarshal(bodyBytes, &response)
			assert.NoError(t, err, "unmarshalling response body should not fail")

			// Check that the response body is as expected
			assert.NotEmptyf(t, response.Data, "expected response body to have data")
			for _, d := range response.Data {
				assert.Equal(t, stakerPk, d.StakerPkHex, "expected response body to match")
			}
			allDataCollected = append(allDataCollected, response.Data...)
			if response.Pagination.NextKey != "" {
				paginationKey = response.Pagination.NextKey
				atLeastOnePage = true
			} else {
				break
			}
		}

		assert.True(t, atLeastOnePage, "expected at least one page of data")
		assert.Equal(t, 11, len(allDataCollected), "expected 11 items in total")
		for _, events := range activeStakingEventsByStaker1 {
			found := false
			for _, d := range allDataCollected {
				if d.StakingTxHashHex == events.StakingTxHashHex {
					found = true
					break
				}
			}
			assert.True(t, found, "expected to find the staking tx in the response")
		}
		for i := 0; i < len(allDataCollected)-1; i++ {
			assert.True(t, allDataCollected[i].StakingTx.StartHeight >= allDataCollected[i+1].StakingTx.StartHeight, "expected collected data to be sorted by start height")
		}
	})
}

func TestActiveStakingFetchedByStakerPkWithInvalidPaginationKey(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	activeStakingEvent := testutils.GenerateRandomActiveStakingEvents(r, &testutils.TestActiveEventGeneratorOpts{
		NumOfEvents:       11,
		FinalityProviders: testutils.GeneratePks(11),
		Stakers:           testutils.GeneratePks(1),
	})
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	sendTestMessage(testServer.Queues.ActiveStakingQueueClient, activeStakingEvent)
	// Wait for 2 seconds to make sure the message is processed
	time.Sleep(2 * time.Second)

	// Test the API with an invalid pagination key
	url := fmt.Sprintf("%s%s?staker_btc_pk=%s&pagination_key=%s", testServer.Server.URL, stakerDelegations, activeStakingEvent[0].StakerPkHex, "btc_to_one_milly")

	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to delegations by staker pk should not fail")
	defer resp.Body.Close()

	// Check that the status code is 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected HTTP 400 Bad Request status")

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var response api.ErrorResponse
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	assert.Equal(t, "invalid pagination key format", response.Message)
}

func TestCheckStakerDelegationAllowOptionRequest(t *testing.T) {
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	url := testServer.Server.URL + checkStakerDelegationUrl
	client := &http.Client{}
	req, err := http.NewRequest("OPTIONS", url, nil)
	assert.NoError(t, err)
	req.Header.Add("Access-Control-Request-Method", "GET")
	req.Header.Add("Origin", "https://app.galxe.com")
	req.Header.Add("Access-Control-Request-Headers", "Content-Type")

	// Send the request
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Check that the status code is HTTP 204
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "expected HTTP 204")
	assert.Equal(t, "https://app.galxe.com", resp.Header.Get("Access-Control-Allow-Origin"), "expected Access-Control-Allow-Origin to be https://app.galxe.com")
	assert.Equal(t, "GET, OPTIONS, POST", resp.Header.Get("Access-Control-Allow-Methods"), "expected Access-Control-Allow-Methods to be GET and OPTIONS")
}

func FuzzCheckStakerActiveDelegations(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 3)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		opts := &testutils.TestActiveEventGeneratorOpts{
			NumOfEvents:        testutils.RandomPositiveInt(r, 10),
			Stakers:            testutils.GeneratePks(1),
			EnforceNotOverflow: true,
		}
		activeStakingEvents := testutils.GenerateRandomActiveStakingEvents(
			r, opts,
		)
		testServer := setupTestServer(t, nil)
		defer testServer.Close()
		sendTestMessage(
			testServer.Queues.ActiveStakingQueueClient, activeStakingEvents,
		)
		time.Sleep(5 * time.Second)

		// Test the API
		stakerPk := activeStakingEvents[0].StakerPkHex
		addresses, err := utils.DeriveAddressesFromNoCoordPk(
			stakerPk, testServer.Config.Server.BTCNetParam,
		)
		assert.NoError(t, err, "failed to get taproot address from staker pk")
		isExist := fetchCheckStakerActiveDelegations(t, testServer, addresses.Taproot, "")

		assert.True(t, isExist, "expected staker to have active delegation")

		// Test the API with a staker PK that never had any active delegation
		stakerPkWithoutDelegation, err := testutils.RandomPk()
		if err != nil {
			t.Fatalf("failed to generate random public key for staker: %v", err)
		}
		addressWithNoDelegation, err := utils.DeriveAddressesFromNoCoordPk(
			stakerPkWithoutDelegation, testServer.Config.Server.BTCNetParam,
		)
		assert.NoError(t, err, "failed to get taproot address from staker pk")
		isExist = fetchCheckStakerActiveDelegations(
			t, testServer, addressWithNoDelegation.Taproot, "",
		)
		assert.False(t, isExist, "expected staker to not have active delegation")

		// Update the staker to have its delegations in a different state
		var unbondingEvents []client.UnbondingStakingEvent
		for _, activeStakingEvent := range activeStakingEvents {
			unbondingEvent := client.NewUnbondingStakingEvent(
				activeStakingEvent.StakingTxHashHex,
				activeStakingEvent.StakingStartHeight+100,
				time.Now().Unix(),
				10,
				1,
				activeStakingEvent.StakingTxHex,     // mocked data, it doesn't matter in stats calculation
				activeStakingEvent.StakingTxHashHex, // mocked data, it doesn't matter in stats calculation
			)
			unbondingEvents = append(unbondingEvents, unbondingEvent)
		}
		sendTestMessage(testServer.Queues.UnbondingStakingQueueClient, unbondingEvents)
		time.Sleep(5 * time.Second)

		isExist = fetchCheckStakerActiveDelegations(t, testServer, addresses.Taproot, "")
		assert.False(t, isExist, "expected staker to not have active delegation")
	})
}

func FuzzCheckStakerActiveDelegationsForToday(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 3)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		stakerPk := testutils.GeneratePks(1)
		opts := &testutils.TestActiveEventGeneratorOpts{
			NumOfEvents:        testutils.RandomPositiveInt(r, 3),
			Stakers:            stakerPk,
			EnforceNotOverflow: true,
			BeforeTimestamp:    utils.GetTodayStartTimestampInSeconds() - 1, // To make it yesterday
		}
		activeStakingEvents := testutils.GenerateRandomActiveStakingEvents(r, opts)
		testServer := setupTestServer(t, nil)
		defer testServer.Close()
		sendTestMessage(
			testServer.Queues.ActiveStakingQueueClient, activeStakingEvents,
		)
		time.Sleep(3 * time.Second)

		// Test the API
		addresses, err := utils.DeriveAddressesFromNoCoordPk(
			stakerPk[0], testServer.Config.Server.BTCNetParam,
		)
		assert.NoError(t, err, "failed to get taproot address from staker pk")
		isExist := fetchCheckStakerActiveDelegations(t, testServer, addresses.Taproot, "")

		assert.True(t, isExist, "expected staker to have active delegation")

		// Test with the is_active_today query parameter
		isExist = fetchCheckStakerActiveDelegations(t, testServer, addresses.Taproot, "today")
		assert.False(t, isExist, "expected staker to not have active delegation")

		opts = &testutils.TestActiveEventGeneratorOpts{
			NumOfEvents:        testutils.RandomPositiveInt(r, 3),
			Stakers:            stakerPk,
			EnforceNotOverflow: true,
			AfterTimestamp:     utils.GetTodayStartTimestampInSeconds(), // To make it today
		}
		activeStakingEvents = testutils.GenerateRandomActiveStakingEvents(r, opts)
		sendTestMessage(
			testServer.Queues.ActiveStakingQueueClient, activeStakingEvents,
		)
		time.Sleep(3 * time.Second)

		isExist = fetchCheckStakerActiveDelegations(t, testServer, addresses.Taproot, "today")
		assert.True(t, isExist, "expected staker to have active delegation")
	})
}

func TestGetDelegationReturnEmptySliceWhenNoDelegation(t *testing.T) {
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	stakerPk, err := testutils.RandomPk()
	assert.NoError(t, err)
	url := testServer.Server.URL + stakerDelegations + "?staker_btc_pk=" + stakerPk
	resp, err := http.Get(url)
	assert.NoError(t, err)

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var response handlers.PublicResponse[[]services.DelegationPublic]
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	assert.NotNil(t, response.Data, "expected response body to have data")
	assert.Equal(t, 0, len(response.Data), "expected response body to have no data")
}

func fetchCheckStakerActiveDelegations(
	t *testing.T, testServer *TestServer, btcAddress string, timeframe string,
) bool {
	url := testServer.Server.URL + checkStakerDelegationUrl + "?address=" + btcAddress
	if timeframe != "" {
		url += "&timeframe=" + timeframe
	}
	resp, err := http.Get(url)
	assert.NoError(t, err)

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var response handlers.PublicResponse[bool]
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	return response.Data
}
