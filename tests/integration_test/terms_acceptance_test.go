package tests

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/api/handlers"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
	"github.com/stretchr/testify/assert"
)

const (
	termsAcceptancePath = "/log-terms-acceptance"
)

func TestTermsAcceptance(t *testing.T) {
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	address, _ := testutils.RandomBtcAddress(r, testServer.Config.Server.BTCNetParam)
	publicKey, _ := testutils.RandomPk()

	// Prepare request body
	requestBody := handlers.TermsAcceptanceLoggingRequest{
		Address:   address,
		PublicKey: publicKey,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	url := testServer.Server.URL + termsAcceptancePath
	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	assert.NoError(t, err, "making POST request to terms acceptance endpoint should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	var response handlers.PublicResponse[handlers.TermsAcceptancePublic]
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err, "decoding response body should not fail")
	assert.Equal(t, true, response.Data.Status)
}

func TestTermsAcceptanceInvalidAddress(t *testing.T) {
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	// Use invalid address
	invalidAddress := "invalidaddress"
	publicKey, _ := testutils.RandomPk()

	requestBody := handlers.TermsAcceptanceLoggingRequest{}
	bodyBytes, _ := json.Marshal(requestBody)

	url := testServer.Server.URL + termsAcceptancePath + "?address=" + invalidAddress + "&public_key=" + publicKey
	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	assert.NoError(t, err, "making POST request to terms acceptance endpoint should not fail")
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected HTTP 400 Bad Request status")
}
