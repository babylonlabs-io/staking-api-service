package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/api/handlers"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	v1model "github.com/babylonlabs-io/staking-api-service/internal/db/model/v1"
	"github.com/babylonlabs-io/staking-api-service/internal/services"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
	"github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	overallStatsEndpoint = "/v1/stats"
	topStakerStatsPath   = "/v1/stats/staker"
)

func TestStatsShouldBeShardedInDb(t *testing.T) {
	activeStakingEvent := buildActiveStakingEvent(t, 10)
	// build the unbonding event based on the active staking event
	var unbondingEvents []client.UnbondingStakingEvent
	for _, event := range activeStakingEvent {
		unbondingEvents = append(unbondingEvents, client.NewUnbondingStakingEvent(
			event.StakingTxHashHex,
			event.StakingStartHeight+100,
			time.Now().Unix(),
			10,
			1,
			event.StakingTxHex,     // mocked data, it doesn't matter in stats calculation
			event.StakingTxHashHex, // mocked data, it doesn't matter in stats calculation
		))
	}
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	sendTestMessage(testServer.Queues.ActiveStakingQueueClient, activeStakingEvent)
	time.Sleep(2 * time.Second)
	sendTestMessage(testServer.Queues.UnbondingStakingQueueClient, unbondingEvents)
	time.Sleep(2 * time.Second)

	// directly read from the db to check that we have more than 2 records in the overall stats collection
	results, err := testutils.InspectDbDocuments[v1model.OverallStatsDocument](
		testServer.Config, model.V1OverallStatsCollection,
	)
	if err != nil {
		t.Fatalf("Failed to inspect DB documents: %v", err)
	}
	assert.Equal(t, 2, len(results), "expected 2 logical shards in the overall stats collection")

	// Sum it up, we shall get 0 active tvl and 0 active delegations. the total should remain positive number
	var totalActiveTvl int64
	var totalActiveDelegations int64
	var totalTvl int64
	var totalDelegations int64
	for _, r := range results {
		totalActiveTvl += r.ActiveTvl
		totalActiveDelegations += r.ActiveDelegations
		totalTvl += r.TotalTvl
		totalDelegations += r.TotalDelegations
	}
	assert.Equal(t, int64(0), totalActiveTvl, "total acvtive tvl shall be 0 as all staking tx are now unbonded")
	assert.Equal(t, int64(0), totalActiveDelegations)
	assert.NotZero(t, totalTvl)
	assert.Equal(t, int64(10), totalDelegations)
}

func TestShouldSkipStatsCalculationForOverflowedStakingEvent(t *testing.T) {
	activeStakingEvent := getTestActiveStakingEvent()
	// Set the overflow flag to true
	activeStakingEvent.IsOverflow = true
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	err := sendTestMessage(testServer.Queues.ActiveStakingQueueClient, []client.ActiveStakingEvent{*activeStakingEvent})
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
	results, err := testutils.InspectDbDocuments[v1model.UnbondingDocument](testServer.Config, model.V1UnbondingCollection)
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

	// directly read from the db to check that we only have 1 shard in the overall stats collection
	stats, err := testutils.InspectDbDocuments[v1model.OverallStatsDocument](
		testServer.Config, model.V1OverallStatsCollection,
	)
	if err != nil {
		t.Fatalf("Failed to inspect DB documents: %v", err)
	}
	assert.Equal(t, 0, len(stats))
}

func TestShouldNotPerformStatsCalculationForUnbondingTxWhenDelegationIsOverflowed(t *testing.T) {
	activeStakingEvent := buildActiveStakingEvent(t, 10)
	// Let's pick a random staking event and set the overflow flag to true
	event := activeStakingEvent[6]
	event.IsOverflow = true
	// build the unbonding event based on the active staking event
	var unbondingEvents []client.UnbondingStakingEvent
	unbondingEvents = append(unbondingEvents, client.NewUnbondingStakingEvent(
		event.StakingTxHashHex,
		event.StakingStartHeight+100,
		time.Now().Unix(),
		10,
		1,
		event.StakingTxHex,     // mocked data, it doesn't matter in stats calculation
		event.StakingTxHashHex, // mocked data, it doesn't matter in stats calculation
	))
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	sendTestMessage(testServer.Queues.ActiveStakingQueueClient, activeStakingEvent)
	time.Sleep(2 * time.Second)
	sendTestMessage(testServer.Queues.UnbondingStakingQueueClient, unbondingEvents)
	time.Sleep(2 * time.Second)

	// directly read from the db to check that we have more than 2 records in the
	// overall stats collection
	results, err := testutils.InspectDbDocuments[v1model.OverallStatsDocument](
		testServer.Config, model.V1OverallStatsCollection,
	)
	if err != nil {
		t.Fatalf("Failed to inspect DB documents: %v", err)
	}
	assert.Equal(t, 2, len(results), "expected 2 logical shards in the overall stats collection")

	// Sum it up, we shall get 0 active tvl and 0 active delegations. the total should remain positive number
	var totalActiveTvl int64
	var totalActiveDelegations int64
	var totalTvl int64
	var totalDelegations int64
	for _, r := range results {
		totalActiveTvl += r.ActiveTvl
		totalActiveDelegations += r.ActiveDelegations
		totalTvl += r.TotalTvl
		totalDelegations += r.TotalDelegations
	}

	// calculate the total expect tvl from the active staking events
	var expectedTotalTvl int64
	for _, e := range activeStakingEvent {
		if !e.IsOverflow {
			expectedTotalTvl += int64(e.StakingValue)
		}
	}
	assert.Equal(t, expectedTotalTvl, totalActiveTvl)
	assert.Equal(t, int64(9), totalActiveDelegations)
	assert.Equal(t, expectedTotalTvl, totalTvl)
	assert.Equal(t, int64(9), totalDelegations)
}

func TestStatsEndpoints(t *testing.T) {
	activeStakingEvent := getTestActiveStakingEvent()
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	sendTestMessage(testServer.Queues.ActiveStakingQueueClient, []client.ActiveStakingEvent{*activeStakingEvent})
	time.Sleep(2 * time.Second)

	// Test the finality endpoint first
	result := fetchFinalityEndpoint(t, testServer)
	assert.Equal(t, 4, len(result))
	for _, r := range result {
		if r.BtcPk == activeStakingEvent.FinalityProviderPkHex {
			assert.Equal(t, int64(activeStakingEvent.StakingValue), r.ActiveTvl)
			assert.Equal(t, int64(activeStakingEvent.StakingValue), r.TotalTvl)
			assert.Equal(t, int64(1), r.ActiveDelegations)
			assert.Equal(t, int64(1), r.TotalDelegations)
		} else {
			assert.Equal(t, int64(0), r.ActiveTvl)
			assert.Equal(t, int64(0), r.TotalTvl)
			assert.Equal(t, int64(0), r.ActiveDelegations)
			assert.Equal(t, int64(0), r.TotalDelegations)
		}
	}

	// Test the overall stats endpoint
	overallStats := fetchOverallStatsEndpoint(t, testServer)
	assert.Equal(t, int64(activeStakingEvent.StakingValue), overallStats.TotalTvl)
	assert.Equal(t, int64(1), overallStats.ActiveDelegations)
	assert.Equal(t, int64(1), overallStats.TotalDelegations)
	assert.Equal(t, uint64(1), overallStats.TotalStakers)
	// We have not yet sent any ConfirmedInfoEvent and UnconfirmedInfoEvent, hence no recrod in db
	assert.Equal(t, int64(0), overallStats.ActiveTvl)
	assert.Equal(t, uint64(0), overallStats.UnconfirmedTvl)
	assert.Equal(t, uint64(0), overallStats.PendingTvl)

	// Test the top staker stats endpoint
	stakerStats := fetchStakerStatsEndpoint(t, testServer, "")
	assert.Equal(t, 1, len(stakerStats))
	assert.Equal(t, activeStakingEvent.StakerPkHex, stakerStats[0].StakerPkHex)
	assert.Equal(t, int64(activeStakingEvent.StakingValue), stakerStats[0].ActiveTvl)
	assert.Equal(t, int64(activeStakingEvent.StakingValue), stakerStats[0].TotalTvl)
	assert.Equal(t, int64(1), stakerStats[0].ActiveDelegations)
	assert.Equal(t, int64(1), stakerStats[0].TotalDelegations)

	// Now let's send an unbonding event
	unbondingEvent := client.NewUnbondingStakingEvent(
		activeStakingEvent.StakingTxHashHex,
		activeStakingEvent.StakingStartHeight+100,
		time.Now().Unix(),
		10,
		1,
		activeStakingEvent.StakingTxHex,     // mocked data, it doesn't matter in stats calculation
		activeStakingEvent.StakingTxHashHex, // mocked data, it doesn't matter in stats calculation
	)
	sendTestMessage(testServer.Queues.UnbondingStakingQueueClient, []client.UnbondingStakingEvent{unbondingEvent})
	time.Sleep(2 * time.Second)

	// Make a GET request to the finality providers endpoint
	result = fetchFinalityEndpoint(t, testServer)
	assert.Equal(t, 4, len(result))
	for _, r := range result {
		if r.BtcPk == activeStakingEvent.FinalityProviderPkHex {
			assert.Equal(t, int64(0), r.ActiveTvl)
			assert.Equal(t, int64(activeStakingEvent.StakingValue), r.TotalTvl)
			assert.Equal(t, int64(0), r.ActiveDelegations)
			assert.Equal(t, int64(1), r.TotalDelegations)
		} else {
			assert.Equal(t, int64(0), r.ActiveTvl)
			assert.Equal(t, int64(0), r.TotalTvl)
			assert.Equal(t, int64(0), r.ActiveDelegations)
			assert.Equal(t, int64(0), r.TotalDelegations)
		}
	}

	overallStats = fetchOverallStatsEndpoint(t, testServer)
	assert.Equal(t, int64(0), overallStats.ActiveTvl)
	assert.Equal(t, int64(activeStakingEvent.StakingValue), overallStats.TotalTvl)
	assert.Equal(t, int64(0), overallStats.ActiveDelegations)
	assert.Equal(t, int64(1), overallStats.TotalDelegations)
	assert.Equal(t, uint64(1), overallStats.TotalStakers)

	stakerStats = fetchStakerStatsEndpoint(t, testServer, "")
	assert.Equal(t, 1, len(stakerStats))
	assert.Equal(t, activeStakingEvent.StakerPkHex, stakerStats[0].StakerPkHex)
	assert.Equal(t, int64(0), stakerStats[0].ActiveTvl)
	assert.Equal(t, int64(activeStakingEvent.StakingValue), stakerStats[0].TotalTvl)
	assert.Equal(t, int64(0), stakerStats[0].ActiveDelegations)
	assert.Equal(t, int64(1), stakerStats[0].TotalDelegations)

	// Send two new active events, it will increment the stats
	activeEvents := buildActiveStakingEvent(t, 2)
	sendTestMessage(testServer.Queues.ActiveStakingQueueClient, activeEvents)
	time.Sleep(2 * time.Second)

	// Make a GET request to the finality providers endpoint
	finalityProviderStats := fetchFinalityEndpoint(t, testServer)
	assert.Equal(t, 6, len(finalityProviderStats))
	// Make sure sorted by active TVL
	for i := 0; i < len(finalityProviderStats)-1; i++ {
		assert.True(t, finalityProviderStats[i].ActiveTvl >= finalityProviderStats[i+1].ActiveTvl, "expected response body to be sorted")
	}

	overallStats = fetchOverallStatsEndpoint(t, testServer)

	expectedTvl := int64(activeEvents[0].StakingValue + activeEvents[1].StakingValue)
	expectedTotalTvl := int64(expectedTvl) + int64(activeStakingEvent.StakingValue)
	assert.Equal(t, expectedTotalTvl, overallStats.TotalTvl)
	assert.Equal(t, int64(2), overallStats.ActiveDelegations)
	assert.Equal(t, int64(3), overallStats.TotalDelegations)
	assert.Equal(t, uint64(2), overallStats.TotalStakers, "expected 2 stakers as the last 2 belong to same staker")

	stakerStats = fetchStakerStatsEndpoint(t, testServer, "")
	assert.Equal(t, 2, len(stakerStats))

	// Also make sure the returned data is sorted by active TVL
	for i := 0; i < len(stakerStats)-1; i++ {
		assert.True(t, stakerStats[i].ActiveTvl >= stakerStats[i+1].ActiveTvl, "expected response body to be sorted")
	}

	// send an BtcInfoEvent which shall update the unconfirmed active TVL
	btcInfoEvent := &client.BtcInfoEvent{
		EventType:      client.BtcInfoEventType,
		Height:         100,
		ConfirmedTvl:   90,
		UnconfirmedTvl: 100,
	}
	sendTestMessage(testServer.Queues.BtcInfoQueueClient, []*client.BtcInfoEvent{btcInfoEvent})

	time.Sleep(2 * time.Second)

	overallStats = fetchOverallStatsEndpoint(t, testServer)
	assert.Equal(t, uint64(100), overallStats.UnconfirmedTvl)
	assert.Equal(t, int64(90), overallStats.ActiveTvl)
}

func TestReturnEmptyArrayWhenNoStakerStatsFound(t *testing.T) {
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	stakerPk, err := testutils.RandomPk()
	require.NoError(t, err)
	stakerStats := fetchStakerStatsEndpoint(t, testServer, stakerPk)
	assert.Equal(t, 0, len(stakerStats))
}

func FuzzReturnStakerStatsByStakerPk(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 3)
	f.Fuzz(func(t *testing.T, seed int64) {
		testServer := setupTestServer(t, nil)
		defer testServer.Close()
		r := rand.New(rand.NewSource(seed))
		events := testutils.GenerateRandomActiveStakingEvents(r, &testutils.TestActiveEventGeneratorOpts{
			NumOfEvents:        testutils.RandomPositiveInt(r, 10),
			Stakers:            testutils.GeneratePks(10),
			EnforceNotOverflow: true,
		})
		sendTestMessage(testServer.Queues.ActiveStakingQueueClient, events)
		time.Sleep(10 * time.Second)

		// Find the unique staker pks
		var stakerPks []string
		for _, e := range events {
			// append into stakerPks if it's not already there
			if !testutils.Contains(stakerPks, e.StakerPkHex) {
				stakerPks = append(stakerPks, e.StakerPkHex)
			}
		}

		// Fetch the staker stats for each staker
		for _, stakerPk := range stakerPks {
			stakerStats := fetchStakerStatsEndpoint(t, testServer, stakerPk)
			assert.Equal(t, 1, len(stakerStats))
		}
	})
}

func FuzzStatsEndpointReturnHighestUnconfirmedTvlFromEvents(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 5)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		testServer := setupTestServer(t, nil)
		defer testServer.Close()

		overallStats := fetchOverallStatsEndpoint(t, testServer)
		assert.Equal(t, uint64(0), overallStats.UnconfirmedTvl)

		highestHeightEvent := &client.BtcInfoEvent{
			EventType:      client.BtcInfoEventType,
			Height:         0,
			ConfirmedTvl:   0,
			UnconfirmedTvl: 0,
		}
		var messages []*client.BtcInfoEvent
		for i := 0; i < 10; i++ {
			confirmedTvl := uint64(testutils.RandomAmount(r))
			btcInfoEvent := &client.BtcInfoEvent{
				EventType:      client.BtcInfoEventType,
				Height:         uint64(testutils.RandomPositiveInt(r, 1000000)),
				ConfirmedTvl:   confirmedTvl,
				UnconfirmedTvl: confirmedTvl + uint64(testutils.RandomAmount(r)),
			}
			messages = append(messages, btcInfoEvent)
			if btcInfoEvent.Height > highestHeightEvent.Height {
				highestHeightEvent = btcInfoEvent
			}
		}
		sendTestMessage(testServer.Queues.BtcInfoQueueClient, messages)
		time.Sleep(5 * time.Second)

		overallStats = fetchOverallStatsEndpoint(t, testServer)
		assert.Equal(t, &highestHeightEvent.UnconfirmedTvl, &overallStats.UnconfirmedTvl)
		pendingTvl := highestHeightEvent.UnconfirmedTvl - highestHeightEvent.ConfirmedTvl
		assert.Equal(t, pendingTvl, overallStats.PendingTvl)
	})
}

func FuzzTestTopStakersWithPaginationResponse(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 3)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		numOfStakers := testutils.RandomPositiveInt(r, 10)
		// Pagination size shall alway be greater than 2
		paginationSize := testutils.RandomPositiveInt(r, 10) + 1
		var events []*client.ActiveStakingEvent
		for i := 0; i < numOfStakers; i++ {
			opts := &testutils.TestActiveEventGeneratorOpts{
				NumOfEvents:        testutils.RandomPositiveInt(r, 1),
				Stakers:            testutils.GeneratePks(1),
				EnforceNotOverflow: true,
			}
			activeStakingEventsByStaker :=
				testutils.GenerateRandomActiveStakingEvents(r, opts)
			events = append(events, activeStakingEventsByStaker...)
		}
		cfg, err := config.New("../config/config-test.yml")
		if err != nil {
			t.Fatalf("Failed to load test config: %v", err)
		}
		cfg.Db.MaxPaginationLimit = int64(paginationSize)

		testServer := setupTestServer(t, &TestServerDependency{ConfigOverrides: cfg})
		defer testServer.Close()
		sendTestMessage(
			testServer.Queues.ActiveStakingQueueClient,
			events,
		)
		time.Sleep(5 * time.Second)
		// Test the API
		url := testServer.Server.URL + topStakerStatsPath
		var paginationKey string
		var allDataCollected []services.StakerStatsPublic
		var numOfRequestsToFetchAllResults int
		for {
			numOfRequestsToFetchAllResults++
			resp, err := http.Get(url + "?pagination_key=" + paginationKey)
			assert.NoError(t, err, "making GET request to staker stats endpoint should not fail")
			assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")
			bodyBytes, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "reading response body should not fail")
			var response handlers.PublicResponse[[]services.StakerStatsPublic]
			err = json.Unmarshal(bodyBytes, &response)
			assert.NoError(t, err, "unmarshalling response body should not fail")
			// Check that the response body is as expected
			allDataCollected = append(allDataCollected, response.Data...)
			if response.Pagination.NextKey != "" {
				assert.NotEmptyf(t, response.Data, "expected response body to have data")
				assert.Equal(t, paginationSize, len(response.Data))
				paginationKey = response.Pagination.NextKey
			} else {
				break
			}
		}

		assert.Equal(t, math.Ceil(float64(numOfStakers)/float64(paginationSize)), float64(numOfRequestsToFetchAllResults))
		assert.Equal(t, numOfStakers, len(allDataCollected))
		for i := 0; i < len(allDataCollected)-1; i++ {
			assert.True(t, allDataCollected[i].ActiveTvl >= allDataCollected[i+1].ActiveTvl, "expected collected data to be sorted by start height")
		}
	})
}

func fetchFinalityEndpoint(t *testing.T, testServer *TestServer) []services.FpDetailsPublic {
	url := testServer.Server.URL + finalityProvidersPath
	// Make a GET request to the finality providers endpoint
	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to finality providers endpoint should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var responseBody handlers.PublicResponse[[]services.FpDetailsPublic]
	err = json.Unmarshal(bodyBytes, &responseBody)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	return responseBody.Data
}

func fetchOverallStatsEndpoint(t *testing.T, testServer *TestServer) services.OverallStatsPublic {
	url := testServer.Server.URL + overallStatsEndpoint
	// Make a GET request to the stats endpoint
	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to stats endpoint should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var responseBody handlers.PublicResponse[services.OverallStatsPublic]
	err = json.Unmarshal(bodyBytes, &responseBody)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	return responseBody.Data
}

func fetchStakerStatsEndpoint(t *testing.T, testServer *TestServer, stakerPk string) []services.StakerStatsPublic {
	url := testServer.Server.URL + topStakerStatsPath
	if stakerPk != "" {
		url += "?staker_btc_pk=" + stakerPk
	}
	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to staker stats endpoint should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var responseBody handlers.PublicResponse[[]services.StakerStatsPublic]
	err = json.Unmarshal(bodyBytes, &responseBody)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	return responseBody.Data
}
