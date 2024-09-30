package tests

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"testing"

	"github.com/babylonlabs-io/staking-api-service/internal/api"
	"github.com/babylonlabs-io/staking-api-service/internal/api/handlers"
	"github.com/babylonlabs-io/staking-api-service/internal/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/clients/ordinals"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/services"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"github.com/babylonlabs-io/staking-api-service/internal/utils"
	"github.com/babylonlabs-io/staking-api-service/tests/mocks"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const verifyUTXOsPath = "/v1/ordinals/verify-utxos"

func TestVerifyUtxosEndpointNotAvailableIfAssetsConfigNotSet(t *testing.T) {
	cfg, err := config.New("../config/config-test.yml")
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

func FuzzSuccessfullyVerifyUTXOsAssetsViaOrdinalService(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		numOfUTXOs := testutils.RandomPositiveInt(r, 100)
		payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs)
		jsonPayload, err := json.Marshal(payload)
		assert.NoError(t, err, "failed to marshal payload")

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

		mockOrdinal.AssertNumberOfCalls(
			t, "FetchUTXOInfos",
			1,
		)
	})
}

func FuzzErrorWhenExceedMaxAllowedLength(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		cfg, err := config.New("../config/config-test.yml")
		if err != nil {
			t.Fatalf("Failed to load test config: %v", err)
		}
		numOfUTXOs := testutils.RandomPositiveInt(r, 100) + int(cfg.Assets.MaxUTXOs)
		payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs)
		jsonPayload, err := json.Marshal(payload)
		assert.NoError(t, err)

		testServer := setupTestServer(t, nil)
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
	})
}

func FuzzErrorWithInvalidTxid(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		cfg, err := config.New("../config/config-test.yml")
		if err != nil {
			t.Fatalf("Failed to load test config: %v", err)
		}
		numOfUTXOs := testutils.RandomPositiveInt(r, int(cfg.Assets.MaxUTXOs))

		payload := createPayload(t, r, &chaincfg.MainNetParams, numOfUTXOs)
		// Create an invalid UTXO txid
		payload.UTXOs[r.Intn(numOfUTXOs)].Txid = testutils.RandomString(r, 64)
		jsonPayload, err := json.Marshal(payload)
		assert.NoError(t, err)

		testServer := setupTestServer(t, nil)
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
	})
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
					Inscriptions: []string{testutils.RandomString(r, r.Intn(100))},
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

func createPayload(t *testing.T, r *rand.Rand, netParam *chaincfg.Params, size int) handlers.VerifyUTXOsRequestPayload {
	var utxos []types.UTXOIdentifier

	for i := 0; i < size; i++ {
		tx, _, err := testutils.GenerateRandomTx(r, nil)
		if err != nil {
			t.Fatalf("Failed to generate random tx: %v", err)
		}
		txid := tx.TxHash().String()
		utxos = append(utxos, types.UTXOIdentifier{
			Txid: txid,
			Vout: uint32(r.Intn(10)),
		})
	}
	pk, err := testutils.RandomPk()
	if err != nil {
		t.Fatalf("Failed to generate random pk: %v", err)
	}
	addresses, err := utils.DeriveAddressesFromNoCoordPk(pk, netParam)
	if err != nil {
		t.Fatalf("Failed to generate taproot address from pk: %v", err)
	}
	return handlers.VerifyUTXOsRequestPayload{
		UTXOs:   utxos,
		Address: addresses.Taproot,
	}
}

// Chunk function to split a slice into chunks of specified size
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil // Return nil if the size is invalid
	}

	var chunks [][]T
	for size < len(slice) {
		slice, chunks = slice[size:], append(chunks, slice[0:size:size])
	}
	chunks = append(chunks, slice)
	return chunks
}
