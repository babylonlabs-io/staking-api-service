package tests

import (
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/api/handlers"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db"
	v1model "github.com/babylonlabs-io/staking-api-service/internal/db/model/v1"
	"github.com/babylonlabs-io/staking-api-service/internal/services"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
	testmock "github.com/babylonlabs-io/staking-api-service/tests/mocks"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	finalityProvidersPath = "/v1/finality-providers"
)

func shouldGetFinalityProvidersSuccessfully(t *testing.T, testServer *TestServer) {
	url := testServer.Server.URL + finalityProvidersPath
	defer testServer.Close()

	responseBody := fetchSuccessfulResponse[[]services.FpDetailsPublic](t, url)
	result := responseBody.Data
	assert.Equal(t, "Babylon Foundation 2", result[2].Description.Moniker)
	assert.Equal(t, "0.060000000000000000", result[1].Commission)
	assert.Equal(t, "0d2f9728abc45c0cdeefdd73f52a0e0102470e35fb689fc5bc681959a61b021f", result[3].BtcPk)
	assert.Equal(t, "094f5861be4128861d69ea4b66a5f974943f100f55400bf26f5cce124b4c9af7", result[2].BtcPk)

	assert.Equal(t, 4, len(result))

	assert.Equal(t, int64(0), result[0].ActiveTvl)
	assert.Equal(t, int64(0), result[0].TotalTvl)
	assert.Equal(t, int64(0), result[0].ActiveDelegations)
	assert.Equal(t, int64(0), result[0].TotalDelegations)
}

func TestGetFinalityProvidersSuccessfully(t *testing.T) {
	testServer := setupTestServer(t, nil)
	shouldGetFinalityProvidersSuccessfully(t, testServer)
}

func TestGetFinalityProviderShouldNotFailInCaseOfDbFailure(t *testing.T) {
	mockDB := new(testmock.V1DBClient)
	mockDB.On("FindFinalityProviderStats", mock.Anything, mock.Anything).Return(nil, errors.New("just an error"))

	testServer := setupTestServer(t, &TestServerDependency{MockDbClients: services.DbClients{
		V1DBClient: mockDB,
	}})
	shouldGetFinalityProvidersSuccessfully(t, testServer)
}

func TestGetFinalityProviderShouldReturnFallbackToGlobalParams(t *testing.T) {
	mockedResultMap := &db.DbResultMap[*v1model.FinalityProviderStatsDocument]{
		Data:            []*v1model.FinalityProviderStatsDocument{},
		PaginationToken: "",
	}
	mockDB := new(testmock.V1DBClient)
	mockDB.On("FindFinalityProviderStats", mock.Anything, mock.Anything).Return(mockedResultMap, nil)

	testServer := setupTestServer(t, &TestServerDependency{MockDbClients: services.DbClients{
		V1DBClient: mockDB,
	}})
	shouldGetFinalityProvidersSuccessfully(t, testServer)
}

func TestGetFinalityProviderReturn4xxErrorIfPageTokenInvalid(t *testing.T) {
	mockDB := new(testmock.V1DBClient)
	mockDB.On("FindFinalityProviderStats", mock.Anything, mock.Anything).Return(nil, &db.InvalidPaginationTokenError{})

	testServer := setupTestServer(t, &TestServerDependency{MockDbClients: services.DbClients{
		V1DBClient: mockDB,
	}})
	url := testServer.Server.URL + finalityProvidersPath
	defer testServer.Close()
	// Make a GET request to the finality providers endpoint
	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to finality providers endpoint should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetFinalityProviderReturn4xxErrorIfPkInvalid(t *testing.T) {
	testServer := setupTestServer(t, nil)
	url := testServer.Server.URL + finalityProvidersPath + "?fp_btc_pk=invalid"
	defer testServer.Close()
	// Make a GET request to the finality providers endpoint
	resp, err := http.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func FuzzGetFinalityProviderShouldReturnAllRegisteredFps(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		fpParams, registeredFpsStats, notRegisteredFpsStats := setUpFinalityProvidersStatsDataSet(t, r, nil)

		mockDB := new(testmock.V1DBClient)
		mockDB.On("FindFinalityProviderStatsByFinalityProviderPkHex",
			mock.Anything, mock.Anything,
		).Return(registeredFpsStats, nil)

		mockedFinalityProviderStats := &db.DbResultMap[*v1model.FinalityProviderStatsDocument]{
			Data:            append(registeredFpsStats, notRegisteredFpsStats...),
			PaginationToken: "",
		}
		mockDB.On("FindFinalityProviderStats", mock.Anything, mock.Anything).Return(mockedFinalityProviderStats, nil)

		testServer := setupTestServer(t, &TestServerDependency{MockDbClients: services.DbClients{
			V1DBClient: mockDB,
		}, MockedFinalityProviders: fpParams})

		url := testServer.Server.URL + finalityProvidersPath
		defer testServer.Close()
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

		result := responseBody.Data
		// Check that the response body is as expected

		assert.NotEmptyf(t, result, "expected response body to be non-empty")
		// We expect all registered finality providers to be returned, plus the one that is not registered
		var fpParamsWithStakingMap = make(map[string]bool)
		for _, fp := range fpParams {
			found := false
			for _, fpStat := range registeredFpsStats {
				if fp.BtcPk == fpStat.FinalityProviderPkHex {
					found = true
					break
				}
			}
			fpParamsWithStakingMap[fp.BtcPk] = found
		}
		assert.Equal(t, len(fpParams)+len(notRegisteredFpsStats), len(result))

		resultMap := make(map[string]services.FpDetailsPublic)
		for _, fp := range result {
			resultMap[fp.BtcPk] = fp
		}

		// Check all the registered finality providers should apprear in the response
		for _, f := range fpParams {
			assert.Equal(t, f.Description.Moniker, resultMap[f.BtcPk].Description.Moniker)
			assert.Equal(t, f.Commission, resultMap[f.BtcPk].Commission)
			// Check that the stats are correct for the registered finality providers without any delegations
			if fpParamsWithStakingMap[f.BtcPk] == false {
				assert.Equal(t, int64(0), resultMap[f.BtcPk].ActiveTvl)
				assert.Equal(t, int64(0), resultMap[f.BtcPk].TotalTvl)
				assert.Equal(t, int64(0), resultMap[f.BtcPk].ActiveDelegations)
				assert.Equal(t, int64(0), resultMap[f.BtcPk].TotalDelegations)
			} else {
				assert.NotZero(t, resultMap[f.BtcPk].ActiveTvl)
				assert.NotZero(t, resultMap[f.BtcPk].TotalTvl)
				assert.NotZero(t, resultMap[f.BtcPk].ActiveDelegations)
				assert.NotZero(t, resultMap[f.BtcPk].TotalDelegations)
			}
		}
		for _, f := range notRegisteredFpsStats {
			assert.Equal(t, "", resultMap[f.FinalityProviderPkHex].Description.Moniker)
		}
	})
}

func FuzzTestGetFinalityProviderWithPaginationResponse(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 3)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		opts := &testutils.TestActiveEventGeneratorOpts{
			NumOfEvents:       20,
			FinalityProviders: testutils.GeneratePks(20),
			Stakers:           testutils.GeneratePks(20),
		}

		activeStakingEvents := testutils.GenerateRandomActiveStakingEvents(r, opts)
		cfg, err := config.New("../config/config-test.yml")
		if err != nil {
			t.Fatalf("Failed to load test config: %v", err)
		}
		cfg.Db.MaxPaginationLimit = 2

		testServer := setupTestServer(t, &TestServerDependency{ConfigOverrides: cfg})
		defer testServer.Close()
		sendTestMessage(testServer.Queues.ActiveStakingQueueClient, activeStakingEvents)
		time.Sleep(10 * time.Second)

		var paginationKey string
		var allDataCollected []services.FpDetailsPublic
		var atLeastOnePage bool
		// Test the API
		for {
			url := testServer.Server.URL + finalityProvidersPath + "?pagination_key=" + paginationKey
			resp, err := http.Get(url)
			assert.NoError(t, err, "making GET request to finality providers endpoint should not fail")
			assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")
			bodyBytes, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "reading response body should not fail")
			var response handlers.PublicResponse[[]services.FpDetailsPublic]
			err = json.Unmarshal(bodyBytes, &response)
			assert.NoError(t, err, "unmarshalling response body should not fail")

			// Check that the response body is as expected
			assert.NotEmptyf(t, response.Data, "expected response body to have data")
			allDataCollected = append(allDataCollected, response.Data...)
			if response.Pagination.NextKey != "" {
				atLeastOnePage = true
				paginationKey = response.Pagination.NextKey
			} else {
				break
			}
		}

		assert.True(t, atLeastOnePage, "expected at least one page")
		for i := 0; i < len(allDataCollected)-1; i++ {
			assert.True(t, allDataCollected[i].ActiveTvl >= allDataCollected[i+1].ActiveTvl)
		}
	})
}

func FuzzGetFinalityProviderShouldNotReturnRegisteredFpWithoutStakingForPaginatedDbResponse(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		fpParams, registeredFpsStats, notRegisteredFpsStats := setUpFinalityProvidersStatsDataSet(t, r, nil)

		mockDB := new(testmock.V1DBClient)
		mockDB.On("FindFinalityProviderStatsByFinalityProviderPkHex",
			mock.Anything, mock.Anything,
		).Return(registeredFpsStats, nil)

		registeredWithoutStakeFpsStats := registeredFpsStats[:len(registeredFpsStats)-testutils.RandomPositiveInt(r, len(registeredFpsStats))]

		mockedFinalityProviderStats := &db.DbResultMap[*v1model.FinalityProviderStatsDocument]{
			Data:            append(registeredWithoutStakeFpsStats, notRegisteredFpsStats...),
			PaginationToken: "abcd",
		}
		mockDB.On("FindFinalityProviderStats", mock.Anything, mock.Anything).Return(mockedFinalityProviderStats, nil)

		testServer := setupTestServer(t, &TestServerDependency{MockDbClients: services.DbClients{
			V1DBClient: mockDB,
		}, MockedFinalityProviders: fpParams})

		url := testServer.Server.URL + finalityProvidersPath
		defer testServer.Close()
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
		result := responseBody.Data

		var registeredFpsWithoutStaking []string
		for _, fp := range fpParams {
			for _, fpStat := range registeredWithoutStakeFpsStats {
				if fp.BtcPk == fpStat.FinalityProviderPkHex {
					registeredFpsWithoutStaking = append(registeredFpsWithoutStaking, fp.BtcPk)
					break
				}
			}
		}

		assert.Equal(t, len(registeredWithoutStakeFpsStats)+len(notRegisteredFpsStats), len(result))
		assert.Less(t, len(registeredFpsWithoutStaking), len(fpParams))
	})
}

func FuzzShouldNotReturnDefaultFpFromParamsWhenPageTokenIsPresent(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		opts := &SetupFpStatsDataSetOpts{
			NumOfRegisterFps:      testutils.RandomPositiveInt(r, 10),
			NumOfNotRegisteredFps: testutils.RandomPositiveInt(r, 10),
		}
		fpParams, registeredFpsStats, _ := setUpFinalityProvidersStatsDataSet(t, r, opts)

		mockDB := new(testmock.V1DBClient)
		// Mock the response for the registered finality providers
		numOfFpNotHaveStats := testutils.RandomPositiveInt(r, int(opts.NumOfRegisterFps))
		mockDB.On("FindFinalityProviderStatsByFinalityProviderPkHex",
			mock.Anything, mock.Anything,
		).Return(registeredFpsStats[:len(registeredFpsStats)-numOfFpNotHaveStats], nil)

		// We are mocking the last page of the response where there is no more data to fetch
		mockedFinalityProviderStats := &db.DbResultMap[*v1model.FinalityProviderStatsDocument]{
			Data:            []*v1model.FinalityProviderStatsDocument{},
			PaginationToken: "",
		}
		mockDB.On("FindFinalityProviderStats", mock.Anything, mock.Anything).Return(mockedFinalityProviderStats, nil)

		testServer := setupTestServer(t, &TestServerDependency{MockDbClients: services.DbClients{
			V1DBClient: mockDB,
		}, MockedFinalityProviders: fpParams})

		url := testServer.Server.URL + finalityProvidersPath + "?pagination_key=abcd"
		defer testServer.Close()
		// Make a GET request to the finality providers endpoint
		resp, err := http.Get(url)
		assert.NoError(t, err, "making GET request to finality providers endpoint should not fail")
		bodyBytes, err := io.ReadAll(resp.Body)
		assert.NoError(t, err, "reading response body should not fail")
		var response handlers.PublicResponse[[]services.FpDetailsPublic]
		err = json.Unmarshal(bodyBytes, &response)
		assert.NoError(t, err, "unmarshalling response body should not fail")
		assert.Equal(t, numOfFpNotHaveStats, len(response.Data))
	})
}

func FuzzGetFinalityProvider(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 3)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		fpParams, registeredFpsStats, notRegisteredFpsStats := setUpFinalityProvidersStatsDataSet(t, r, nil)
		// Manually force a single value for the finality provider to be used in db mocking
		fpStats := []*v1model.FinalityProviderStatsDocument{registeredFpsStats[0]}

		mockDB := new(testmock.V1DBClient)
		mockDB.On("FindFinalityProviderStatsByFinalityProviderPkHex",
			mock.Anything, mock.Anything,
		).Return(fpStats, nil)

		testServer := setupTestServer(t, &TestServerDependency{MockDbClients: services.DbClients{
			V1DBClient: mockDB,
		}, MockedFinalityProviders: fpParams})
		url := testServer.Server.URL + finalityProvidersPath + "?fp_btc_pk=" + fpParams[0].BtcPk
		// Make a GET request to the finality providers endpoint
		respBody := fetchSuccessfulResponse[[]services.FpDetailsPublic](t, url)
		result := respBody.Data
		assert.Equal(t, 1, len(result))
		assert.Equal(t, fpParams[0].Description.Moniker, result[0].Description.Moniker)
		assert.Equal(t, fpParams[0].Commission, result[0].Commission)
		assert.Equal(t, fpParams[0].BtcPk, result[0].BtcPk)
		assert.Equal(t, registeredFpsStats[0].ActiveTvl, result[0].ActiveTvl)
		assert.Equal(t, registeredFpsStats[0].TotalTvl, result[0].TotalTvl)
		assert.Equal(t, registeredFpsStats[0].ActiveDelegations, result[0].ActiveDelegations)
		assert.Equal(t, registeredFpsStats[0].TotalDelegations, result[0].TotalDelegations)
		testServer.Close()

		// Test the API with a non-existent finality provider from notRegisteredFpsStats
		fpStats = []*v1model.FinalityProviderStatsDocument{notRegisteredFpsStats[0]}
		mockDB = new(testmock.V1DBClient)
		mockDB.On("FindFinalityProviderStatsByFinalityProviderPkHex",
			mock.Anything, mock.Anything,
		).Return(fpStats, nil)
		testServer = setupTestServer(t, &TestServerDependency{
			MockDbClients: services.DbClients{
				V1DBClient: mockDB,
			},
			MockedFinalityProviders: fpParams,
		})
		notRegisteredFp := notRegisteredFpsStats[0]
		url = testServer.Server.URL +
			finalityProvidersPath +
			"?fp_btc_pk=" + notRegisteredFp.FinalityProviderPkHex
		respBody = fetchSuccessfulResponse[[]services.FpDetailsPublic](t, url)
		result = respBody.Data
		assert.Equal(t, 1, len(result))
		assert.Equal(t, "", result[0].Description.Moniker)
		assert.Equal(t, "", result[0].Commission)
		assert.Equal(t, notRegisteredFp.FinalityProviderPkHex, result[0].BtcPk)
		assert.Equal(t, notRegisteredFp.ActiveTvl, result[0].ActiveTvl)
		testServer.Close()

		// Test the API with a non-existent finality provider PK
		randomPk, err := testutils.RandomPk()
		testServer = setupTestServer(t, &TestServerDependency{
			MockedFinalityProviders: fpParams,
		})
		defer testServer.Close()
		assert.NoError(t, err, "generating random public key should not fail")
		url = testServer.Server.URL + finalityProvidersPath + "?fp_btc_pk=" + randomPk
		respBody = fetchSuccessfulResponse[[]services.FpDetailsPublic](t, url)
		result = respBody.Data
		assert.Equal(t, 0, len(result))
	})
}

func generateFinalityProviderStatsDocument(r *rand.Rand, pk string) *v1model.FinalityProviderStatsDocument {
	return &v1model.FinalityProviderStatsDocument{
		FinalityProviderPkHex: pk,
		ActiveTvl:             testutils.RandomAmount(r),
		TotalTvl:              testutils.RandomAmount(r),
		ActiveDelegations:     r.Int63n(100) + 1,
		TotalDelegations:      r.Int63n(1000) + 1,
	}
}

type SetupFpStatsDataSetOpts struct {
	NumOfRegisterFps      int
	NumOfNotRegisteredFps int
}

func setUpFinalityProvidersStatsDataSet(t *testing.T, r *rand.Rand, opts *SetupFpStatsDataSetOpts) ([]types.FinalityProviderDetails, []*v1model.FinalityProviderStatsDocument, []*v1model.FinalityProviderStatsDocument) {
	numOfRegisterFps := testutils.RandomPositiveInt(r, 10)
	numOfNotRegisteredFps := testutils.RandomPositiveInt(r, 10)
	if opts != nil {
		numOfRegisterFps = opts.NumOfRegisterFps
		numOfNotRegisteredFps = opts.NumOfNotRegisteredFps
	}
	fpParams := testutils.GenerateRandomFinalityProviderDetail(r, uint64(numOfRegisterFps))

	// Generate a set of registered finality providers
	var registeredFpsStats []*v1model.FinalityProviderStatsDocument
	for i := 0; i < numOfRegisterFps; i++ {
		fpStats := generateFinalityProviderStatsDocument(r, fpParams[i].BtcPk)
		registeredFpsStats = append(registeredFpsStats, fpStats)
	}

	var notRegisteredFpsStats []*v1model.FinalityProviderStatsDocument
	for i := 0; i < numOfNotRegisteredFps; i++ {
		fpNotRegisteredPk, err := testutils.RandomPk()
		assert.NoError(t, err, "generating random public key should not fail")

		stats := generateFinalityProviderStatsDocument(r, fpNotRegisteredPk)
		notRegisteredFpsStats = append(notRegisteredFpsStats, stats)
	}
	assert.LessOrEqual(t, len(registeredFpsStats), len(fpParams))

	return fpParams, registeredFpsStats, notRegisteredFpsStats
}
