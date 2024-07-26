package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/babylonchain/staking-api-service/internal/api"
	"github.com/babylonchain/staking-api-service/internal/api/handlers"
	"github.com/babylonchain/staking-api-service/internal/clients"
	"github.com/babylonchain/staking-api-service/internal/clients/ordinals"
	"github.com/babylonchain/staking-api-service/internal/clients/unisat"
	"github.com/babylonchain/staking-api-service/internal/config"
	"github.com/babylonchain/staking-api-service/internal/services"
	"github.com/babylonchain/staking-api-service/internal/types"
	"github.com/babylonchain/staking-api-service/internal/utils"
	"github.com/babylonchain/staking-api-service/tests/mocks"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const verifyUTXOsPath = "/v1/ordinals/verify-utxos"

func createPayload(t *testing.T, r *rand.Rand, netParam *chaincfg.Params, size int, hasInvalid bool) handlers.VerifyUTXOsRequestPayload {
	var utxos []types.UTXOIdentifier

	numInvalidUTXOs := 0
	if hasInvalid {
		numInvalidUTXOs = r.Intn(size) + 1
	}

	for i := 0; i < size; i++ {
		var txid string
		if hasInvalid && i < numInvalidUTXOs {
			randomStr := randomString(r, 100)
			txid = randomStr
		} else {
			tx, _, err := generateRandomTx(r)
			if err != nil {
				t.Fatalf("Failed to generate random tx: %v", err)
			}
			txid = tx.TxHash().String()
		}
		utxos = append(utxos, types.UTXOIdentifier{
			Txid: txid,
			Vout: uint32(r.Intn(10)),
		})
	}
	pk, err := randomPk()
	if err != nil {
		t.Fatalf("Failed to generate random pk: %v", err)
	}
	address, err := utils.GetTaprootAddressFromPk(pk, netParam)
	if err != nil {
		t.Fatalf("Failed to generate taproot address from pk: %v", err)
	}
	return handlers.VerifyUTXOsRequestPayload{
		UTXOs:   utxos,
		Address: address,
	}
}

func TestVerifyUtxosEndpointNotAvailableIfAssetsConfigNotSet(t *testing.T) {
	cfg, err := config.New("./config/config-test.yml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}
	cfg.Assets = nil

	testServer := setupTestServer(t, &TestServerDependency{ConfigOverrides: cfg})
	defer testServer.Close()

	url := testServer.Server.URL + verifyUTXOsPath
	resp, err := http.Post(url, "application/json", bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatalf("Failed to make POST request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// Test case 1: Fetching UTXOs return via Ordinal Service
func FuzzSuccessfullyVerifyUTXOsAssetsViaOrdinalService(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		numOfUTXOs := randomPositiveInt(r, 100)
		payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs, false)
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("Failed to marshal payload: %v", err)
		}

		// create some ordinal responses that contains inscriptions
		numOfUTXOsWithAsset := r.Intn(numOfUTXOs)

		var txidsWithAsset []string
		for i := 0; i < numOfUTXOsWithAsset; i++ {
			txidsWithAsset = append(txidsWithAsset, payload.UTXOs[i].Txid)
		}

		mockedOrdinalResponse := createOrdinalServiceResponse(t, r, payload.UTXOs, txidsWithAsset)

		mockOrdinal := new(mocks.OrdinalsClientInterface)
		mockOrdinal.On("FetchUTXOInfos", mock.Anything, mock.Anything).Return(mockedOrdinalResponse, nil)
		mockedClients := &clients.Clients{
			Ordinals: mockOrdinal,
		}
		testServer := setupTestServer(t, &TestServerDependency{MockedClients: mockedClients})
		defer testServer.Close()

		url := testServer.Server.URL + verifyUTXOsPath
		resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
		if err != nil {
			t.Fatalf("Failed to make POST request to %s: %v", url, err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		// decode the response body
		var response handlers.PublicResponse[[]services.SafeUTXOPublic]
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		// check the response
		assert.Equal(t, len(payload.UTXOs), len(response.Data))
		// check if the inscriptions are correctly returned and order is preserved
		for i, u := range response.Data {
			// Make sure the UTXO identifiers are correct
			assert.Equal(t, payload.UTXOs[i].Txid, u.TxId)
			assert.Equal(t, payload.UTXOs[i].Vout, u.Vout)
			var isWithAsset bool
			for _, txid := range txidsWithAsset {
				if txid == u.TxId {
					assert.True(t, u.Inscription)
					isWithAsset = true
					break
				}
			}
			if !isWithAsset {
				assert.False(t, u.Inscription)
			}
		}
	})
}

// Test case 2: Fetching more than 100 UTXOs should return an error
func TestVerifyUtxosEndpointExceedMaxAllowedLength(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	numOfUTXOs := 101 // Create 101 UTXOs
	payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs, false)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// create some ordinal responses that contains inscriptions
	numOfUTXOsWithAsset := r.Intn(numOfUTXOs)

	var txidsWithAsset []string
	for i := 0; i < numOfUTXOsWithAsset; i++ {
		txidsWithAsset = append(txidsWithAsset, payload.UTXOs[i].Txid)
	}

	mockedOrdinalResponse := createOrdinalServiceResponse(t, r, payload.UTXOs, txidsWithAsset)

	mockOrdinal := new(mocks.OrdinalsClientInterface)
	mockOrdinal.On("FetchUTXOInfos", mock.Anything, mock.Anything).Return(mockedOrdinalResponse, nil)
	mockedClients := &clients.Clients{
		Ordinals: mockOrdinal,
	}

	testServer := setupTestServer(t, &TestServerDependency{MockedClients: mockedClients})
	defer testServer.Close()

	url := testServer.Server.URL + verifyUTXOsPath
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to make POST request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// decode the response body
	var response api.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	assert.Equal(t, types.BadRequest.String(), response.ErrorCode, "expected error code to be BAD_REQUEST")
	assert.Equal(t, "too many UTXOs in the request", response.Message, "expected error message to be 'too many UTXOs in the request'")
}

// Test case 3: Invalid UTXO txid should return an error
func TestVerifyUtxosEndpointWithMixedUTXOs(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	numOfUTXOs := 10

	payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs, true)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// create some ordinal responses that contains inscriptions
	numOfUTXOsWithAsset := r.Intn(numOfUTXOs)

	var txidsWithAsset []string
	for i := 0; i < numOfUTXOsWithAsset; i++ {
		txidsWithAsset = append(txidsWithAsset, payload.UTXOs[i].Txid)
	}

	mockedOrdinalResponse := createOrdinalServiceResponse(t, r, payload.UTXOs, txidsWithAsset)

	mockOrdinal := new(mocks.OrdinalsClientInterface)
	mockOrdinal.On("FetchUTXOInfos", mock.Anything, mock.Anything).Return(mockedOrdinalResponse, nil)
	mockedClients := &clients.Clients{
		Ordinals: mockOrdinal,
	}

	testServer := setupTestServer(t, &TestServerDependency{MockedClients: mockedClients})
	defer testServer.Close()

	url := testServer.Server.URL + verifyUTXOsPath
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to make POST request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response api.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	assert.Equal(t, types.BadRequest.String(), response.ErrorCode, "expected error code to be BAD_REQUEST")
	assert.Contains(t, response.Message, "invalid UTXO txid", "expected error message to contain 'invalid UTXO txid'")
}

// Test case 4: Ordinal service return error, fallback to unisat service and return the result
func TestVerifyUtxosEndpointOrdinalServiceErrorFallbackToUnisat(t *testing.T) {
	cfg, err := config.New("./config/config-test.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	numOfUTXOs := randomPositiveInt(r, 100)
	payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs, false)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	numOfUTXOsWithAsset := r.Intn(numOfUTXOs)
	var txidsWithAsset []string
	for i := 0; i < numOfUTXOsWithAsset; i++ {
		txidsWithAsset = append(txidsWithAsset, payload.UTXOs[i].Txid)
	}

	mockUnisatResponse := createUnisatServiceResponse(t, r, payload.UTXOs, txidsWithAsset)

	mockOrdinal := new(mocks.OrdinalsClientInterface)
	mockOrdinal.On("FetchUTXOInfos", mock.Anything, mock.Anything).Return(nil, types.NewErrorWithMsg(
		http.StatusInternalServerError,
		types.InternalServiceError,
		"failed to verify ordinals via ordinals service",
	))

	mockUnisat := new(mocks.UnisatClientInterface)
	mockUnisat.On("FetchInscriptionsUtxosByAddress", mock.Anything, mock.Anything, mock.Anything).
		Return(mockUnisatResponse, nil).Once()
	mockUnisat.On("FetchInscriptionsUtxosByAddress", mock.Anything, mock.Anything, mock.Anything).
		Return([]*unisat.UnisatUTXO{}, nil)

	mockedClients := &clients.Clients{
		Ordinals: mockOrdinal,
		Unisat:   mockUnisat,
	}

	testServer := setupTestServer(t, &TestServerDependency{
		MockedClients:   mockedClients,
		ConfigOverrides: cfg,
	})
	defer testServer.Close()

	url := testServer.Server.URL + verifyUTXOsPath
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to make POST request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response handlers.PublicResponse[[]services.SafeUTXOPublic]
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	assert.Equal(t, len(payload.UTXOs), len(response.Data))

	unisatUTXOMap := make(map[string]*unisat.UnisatUTXO)
	for _, u := range mockUnisatResponse {
		unisatUTXOMap[u.TxId] = u
	}

	for _, u := range response.Data {
		unisatUTXO, exists := unisatUTXOMap[u.TxId]
		assert.True(t, exists, "UTXO should exist in Unisat response")

		if len(unisatUTXO.Inscriptions) > 0 {
			assert.True(t, u.Inscription, "UTXO should be marked as having an inscription")
		} else {
			assert.False(t, u.Inscription, "UTXO should not be marked as having an inscription")
		}
	}
}

// Test case 5: Unisat service return error, return error
func TestVerifyUtxosEndpointUnisatServiceErrorReturnError(t *testing.T) {
	cfg, err := config.New("./config/config-test.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	r := rand.New(rand.NewSource(time.Now().Unix()))
	numOfUTXOs := randomPositiveInt(r, 100)
	payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs, false)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// create some ordinal responses that contains inscriptions
	numOfUTXOsWithAsset := r.Intn(numOfUTXOs)

	var txidsWithAsset []string
	for i := 0; i < numOfUTXOsWithAsset; i++ {
		txidsWithAsset = append(txidsWithAsset, payload.UTXOs[i].Txid)
	}

	mockOrdinal := new(mocks.OrdinalsClientInterface)
	mockOrdinal.On("FetchUTXOInfos", mock.Anything, mock.Anything).Return(nil, types.NewErrorWithMsg(
		http.StatusInternalServerError,
		types.InternalServiceError,
		"failed to verify ordinals via ordinals service",
	))

	mockUnisat := new(mocks.UnisatClientInterface)
	mockUnisat.On("FetchInscriptionsUtxosByAddress", mock.Anything, mock.Anything, mock.Anything).Return(nil, types.NewErrorWithMsg(
		http.StatusInternalServerError,
		types.InternalServiceError,
		"failed to verify ordinals via unisat service",
	))

	mockedClients := &clients.Clients{
		Ordinals: mockOrdinal,
		Unisat:   mockUnisat,
	}
	testServer := setupTestServer(t, &TestServerDependency{
		MockedClients:   mockedClients,
		ConfigOverrides: cfg,
	})
	defer testServer.Close()

	url := testServer.Server.URL + verifyUTXOsPath
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to make POST request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	// decode the response body
	var response api.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	assert.Equal(t, types.InternalServiceError.String(), response.ErrorCode, "expected error code to be INTERNAL_SERVICE_ERROR")
	assert.Contains(t, response.Message, "Internal service error", "expected error message to contain 'Internal service error'")
}

// Test case 6: Ordinal service took too long to respond, fallback to unisat service and return the result
func TestVerifyUtxosEndpointOrdinalServiceTimeoutFallbackToUnisat(t *testing.T) {
	cfg, err := config.New("./config/config-test.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	r := rand.New(rand.NewSource(time.Now().Unix()))
	numOfUTXOs := randomPositiveInt(r, 100)
	payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs, false)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create some UTXOs with assets
	numOfUTXOsWithAsset := r.Intn(numOfUTXOs)
	var txidsWithAsset []string
	for i := 0; i < numOfUTXOsWithAsset; i++ {
		txidsWithAsset = append(txidsWithAsset, payload.UTXOs[i].Txid)
	}

	mockUnisatResponse := createUnisatServiceResponse(t, r, payload.UTXOs, txidsWithAsset)

	mockOrdinal := new(mocks.OrdinalsClientInterface)
	mockOrdinal.On("FetchUTXOInfos", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(2 * time.Second)
	}).Return(nil, types.NewErrorWithMsg(http.StatusRequestTimeout,
		types.RequestTimeout,
		"request timeout after"))

	mockUnisat := new(mocks.UnisatClientInterface)

	mockUnisat.On("FetchInscriptionsUtxosByAddress", mock.Anything, mock.Anything, mock.Anything).
		Return(mockUnisatResponse, nil).Once()

	mockUnisat.On("FetchInscriptionsUtxosByAddress", mock.Anything, mock.Anything, mock.Anything).
		Return([]*unisat.UnisatUTXO{}, nil)

	mockedClients := &clients.Clients{
		Ordinals: mockOrdinal,
		Unisat:   mockUnisat,
	}

	testServer := setupTestServer(t, &TestServerDependency{
		MockedClients:   mockedClients,
		ConfigOverrides: cfg,
	})
	defer testServer.Close()

	url := testServer.Server.URL + verifyUTXOsPath
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to make POST request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Decode the response body
	var response handlers.PublicResponse[[]services.SafeUTXOPublic]
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// check the response
	assert.Equal(t, len(payload.UTXOs), len(response.Data))

	unisatUTXOMap := make(map[string]*unisat.UnisatUTXO)
	for _, u := range mockUnisatResponse {
		unisatUTXOMap[u.TxId] = u
	}

	for i, u := range response.Data {
		assert.Equal(t, payload.UTXOs[i].Txid, u.TxId)
		assert.Equal(t, payload.UTXOs[i].Vout, u.Vout)

		unisatUTXO, exists := unisatUTXOMap[u.TxId]
		assert.True(t, exists, "UTXO should exist in Unisat response")

		if len(unisatUTXO.Inscriptions) > 0 {
			assert.True(t, u.Inscription, "UTXO should be marked as having an inscription")
		} else {
			assert.False(t, u.Inscription, "UTXO should not be marked as having an inscription")
		}
	}
}

// Test case 7: Unisat service took too long to respond, return error within the timeout window
func TestVerifyUtxosEndpointUnisatServiceTimeoutReturnError(t *testing.T) {
	cfg, err := config.New("./config/config-test.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	r := rand.New(rand.NewSource(time.Now().Unix()))
	numOfUTXOs := randomPositiveInt(r, 100)
	payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs, false)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create some UTXOs with assets
	numOfUTXOsWithAsset := r.Intn(numOfUTXOs)
	var txidsWithAsset []string
	for i := 0; i < numOfUTXOsWithAsset; i++ {
		txidsWithAsset = append(txidsWithAsset, payload.UTXOs[i].Txid)
	}

	mockOrdinal := new(mocks.OrdinalsClientInterface)
	mockOrdinal.On("FetchUTXOInfos", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(2 * time.Second)
	}).Return(nil, types.NewErrorWithMsg(http.StatusRequestTimeout,
		types.RequestTimeout,
		"request timeout after"))

	// Mock Unisat service
	mockUnisat := new(mocks.UnisatClientInterface)
	mockUnisat.On("FetchInscriptionsUtxosByAddress", mock.Anything, mock.Anything, mock.Anything).Return(nil, types.NewErrorWithMsg(http.StatusRequestTimeout,
		types.RequestTimeout,
		"request timeout after"))

	mockedClients := &clients.Clients{
		Ordinals: mockOrdinal,
		Unisat:   mockUnisat,
	}

	testServer := setupTestServer(t, &TestServerDependency{
		MockedClients:   mockedClients,
		ConfigOverrides: cfg,
	})

	defer testServer.Close()

	url := testServer.Server.URL + verifyUTXOsPath
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to make POST request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	// decode the response body
	var response api.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	assert.Equal(t, types.RequestTimeout.String(), response.ErrorCode, "expected error code to be REQUEST_TIMEOUT")
	assert.Contains(t, response.Message, "request timeout after", "expected error message to contain 'request timeout after'")
}

// Test case 8: Send 100 UTXOs, ordinal service fail, unisat service return 50(limit), then fetch again for the remaining 50 item (test the pagination)
func TestVerifyUtxosEndpointPaginationWithOrdinalServiceFailure(t *testing.T) {
	cfg, err := config.New("./config/config-test.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	numOfUTXOs := 100
	payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs, false)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create some UTXOs with assets
	numOfUTXOsWithAsset := r.Intn(numOfUTXOs)
	var txidsWithAsset []string
	for i := 0; i < numOfUTXOsWithAsset; i++ {
		txidsWithAsset = append(txidsWithAsset, payload.UTXOs[i].Txid)
	}

	mockOrdinal := new(mocks.OrdinalsClientInterface)
	mockOrdinal.On("FetchUTXOInfos", mock.Anything, mock.Anything).Return(nil, types.NewErrorWithMsg(
		http.StatusInternalServerError,
		types.InternalServiceError,
		"failed to verify ordinals via ordinals service",
	))

	callCount := 0
	var mockUnisatResponses [][]*unisat.UnisatUTXO
	mockUnisat := new(mocks.UnisatClientInterface)
	mockUnisat.On("FetchInscriptionsUtxosByAddress", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			callCount++
			if callCount > 10 {
				t.Fatalf("Too many calls to Unisat service (> 10), possible infinite loop")
			}
		}).
		Return(func(ctx context.Context, address string, offset uint32) []*unisat.UnisatUTXO {
			start := int(offset)
			end := start + int(cfg.Assets.Unisat.Limit)
			if end > len(payload.UTXOs) {
				end = len(payload.UTXOs)
			}
			if start >= len(payload.UTXOs) {
				return []*unisat.UnisatUTXO{}
			}
			response := createUnisatServiceResponse(t, r, payload.UTXOs[start:end], txidsWithAsset)
			mockUnisatResponses = append(mockUnisatResponses, response)
			return response
		}, nil)

	mockedClients := &clients.Clients{
		Ordinals: mockOrdinal,
		Unisat:   mockUnisat,
	}

	testServer := setupTestServer(t, &TestServerDependency{
		MockedClients:   mockedClients,
		ConfigOverrides: cfg,
	})
	defer testServer.Close()

	url := testServer.Server.URL + verifyUTXOsPath
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to make POST request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Decode the response body
	var response handlers.PublicResponse[[]services.SafeUTXOPublic]
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	assert.Equal(t, numOfUTXOs, len(response.Data), "Response should contain all UTXOs")

	assert.LessOrEqual(t, callCount, 3, "Unisat service should not be called more than 3 times")

	// Create a map of all Unisat UTXOs
	unisatUTXOMap := make(map[string]*unisat.UnisatUTXO)
	for _, resp := range mockUnisatResponses {
		for _, u := range resp {
			unisatUTXOMap[u.TxId] = u
		}
	}

	// Verify the correctness of the response
	for i, u := range response.Data {
		assert.Equal(t, payload.UTXOs[i].Txid, u.TxId)
		assert.Equal(t, payload.UTXOs[i].Vout, u.Vout)

		unisatUTXO, exists := unisatUTXOMap[u.TxId]
		assert.True(t, exists, "UTXO should exist in Unisat response")

		if len(unisatUTXO.Inscriptions) > 0 {
			assert.True(t, u.Inscription, "UTXO should be marked as having an inscription")
		} else {
			assert.False(t, u.Inscription, "UTXO should not be marked as having an inscription")
		}
	}
}

// Test case 9: Fall back to unisat when ordinal service return data that is not in the right order of the request UTXOs
func TestVerifyUtxosEndpointFallbackToUnisatOnOrdinalServiceWrongOrder(t *testing.T) {
	cfg, err := config.New("./config/config-test.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	r := rand.New(rand.NewSource(time.Now().Unix()))
	numOfUTXOs := 10
	payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs, false)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create some UTXOs with assets
	numOfUTXOsWithAsset := r.Intn(numOfUTXOs)
	var txidsWithAsset []string
	for i := 0; i < numOfUTXOsWithAsset; i++ {
		txidsWithAsset = append(txidsWithAsset, payload.UTXOs[i].Txid)
	}

	mockedOrdinalResponse := createOrdinalServiceResponse(t, r, payload.UTXOs, txidsWithAsset)

	// Shuffle the ordinal response to simulate wrong order
	r.Shuffle(len(mockedOrdinalResponse), func(i, j int) {
		mockedOrdinalResponse[i], mockedOrdinalResponse[j] = mockedOrdinalResponse[j], mockedOrdinalResponse[i]
	})

	mockOrdinal := new(mocks.OrdinalsClientInterface)
	mockOrdinal.On("FetchUTXOInfos", mock.Anything, mock.Anything).Return(mockedOrdinalResponse, types.NewErrorWithMsg(
		http.StatusInternalServerError,
		types.InternalServiceError,
		"response does not contain all requested UTXOs or in the wrong order",
	))

	mockUnisatResponse := createUnisatServiceResponse(t, r, payload.UTXOs, txidsWithAsset)

	mockUnisat := new(mocks.UnisatClientInterface)
	mockUnisat.On("FetchInscriptionsUtxosByAddress", mock.Anything, mock.Anything, mock.Anything).
		Return(mockUnisatResponse, nil).Once()

	mockedClients := &clients.Clients{
		Ordinals: mockOrdinal,
		Unisat:   mockUnisat,
	}

	testServer := setupTestServer(t, &TestServerDependency{
		MockedClients:   mockedClients,
		ConfigOverrides: cfg,
	})
	defer testServer.Close()

	url := testServer.Server.URL + verifyUTXOsPath
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to make POST request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Decode the response body
	var response handlers.PublicResponse[[]services.SafeUTXOPublic]
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	assert.Equal(t, numOfUTXOs, len(response.Data), "Response should contain all UTXOs")

	unisatUTXOMap := make(map[string]*unisat.UnisatUTXO)
	for _, u := range append(mockUnisatResponse) {
		unisatUTXOMap[u.TxId] = u
	}

	inscriptionsCount := 0
	for i, u := range response.Data {
		assert.Equal(t, payload.UTXOs[i].Txid, u.TxId)
		assert.Equal(t, payload.UTXOs[i].Vout, u.Vout)

		unisatUTXO, exists := unisatUTXOMap[u.TxId]
		assert.True(t, exists, "UTXO should exist in Unisat response")

		if len(unisatUTXO.Inscriptions) > 0 {
			assert.True(t, u.Inscription, "UTXO should be marked as having an inscription")
			inscriptionsCount++
		} else {
			assert.False(t, u.Inscription, "UTXO should not be marked as having an inscription")
		}
	}
}

func createOrdinalServiceResponse(t *testing.T, r *rand.Rand, utxos []types.UTXOIdentifier, txidsWithAsset []string) []ordinals.OrdinalsOutputResponse {
	var responses []ordinals.OrdinalsOutputResponse

	for _, utxo := range utxos {
		withAsset := false
		for _, txid := range txidsWithAsset {
			if txid == utxo.Txid {
				withAsset = true
				break
			}
		}
		if withAsset {
			// randomly inject runes or inscriptions
			if r.Intn(2) == 0 {
				responses = append(responses, ordinals.OrdinalsOutputResponse{
					Transaction:  utxo.Txid,
					Inscriptions: []string{randomString(r, r.Intn(100))},
					Runes:        json.RawMessage(`{}`),
				})
			} else {
				responses = append(responses, ordinals.OrdinalsOutputResponse{
					Transaction:  utxo.Txid,
					Inscriptions: []string{},
					Runes:        json.RawMessage(`{"rune1": "rune1"}`),
				})
			}
		} else {
			responses = append(responses, ordinals.OrdinalsOutputResponse{
				Transaction:  utxo.Txid,
				Inscriptions: []string{},
				Runes:        json.RawMessage(`{}`),
			})
		}

	}
	return responses
}

func createUnisatServiceResponse(t *testing.T, r *rand.Rand, utxos []types.UTXOIdentifier, txidsWithAsset []string) []*unisat.UnisatUTXO {
	var responses []*unisat.UnisatUTXO

	for _, utxo := range utxos {
		withAsset := false
		for _, txid := range txidsWithAsset {
			if txid == utxo.Txid {
				withAsset = true
				break
			}
		}

		unisatUTXO := &unisat.UnisatUTXO{
			TxId: utxo.Txid,
			Vout: utxo.Vout,
		}

		if withAsset {
			if r.Intn(2) == 0 {
				numInscriptions := r.Intn(3) + 1
				for i := 0; i < numInscriptions; i++ {
					inscription := &unisat.UnisatInscriptions{
						InscriptionId: randomString(r, 64),
						Offset:        uint32(r.Intn(1000)),
					}
					unisatUTXO.Inscriptions = append(unisatUTXO.Inscriptions, inscription)
				}
			}
		}

		responses = append(responses, unisatUTXO)
	}

	return responses
}
