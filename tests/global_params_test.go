package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/babylonlabs-io/staking-api-service/internal/api/handlers"
	"github.com/babylonlabs-io/staking-api-service/internal/services"
)

const (
	globalParamsPath = "/v1/global-params"
)

func TestGlobalParams(t *testing.T) {
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	url := testServer.Server.URL + globalParamsPath

	// Make a GET request to the global params endpoint
	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to global params endpoint should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var responseBody handlers.PublicResponse[services.GlobalParamsPublic]
	err = json.Unmarshal(bodyBytes, &responseBody)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	result := responseBody.Data.Versions
	// Check that the response body is as expected
	assert.NotEmptyf(t, result, "expected response body to be non-empty")
	assert.Equal(t, 4, len(result))
	versionedGlobalParam := result[0]
	assert.Equal(t, uint64(0), versionedGlobalParam.Version)
	assert.Equal(t, uint64(100), versionedGlobalParam.ActivationHeight)
	assert.Equal(t, uint64(5000000), versionedGlobalParam.StakingCap)
	assert.Equal(t, "01020304", versionedGlobalParam.Tag)
	assert.Equal(t, 5, len(versionedGlobalParam.CovenantPks))
	assert.Equal(t, uint64(3), versionedGlobalParam.CovenantQuorum)
	assert.Equal(t, uint64(1000), versionedGlobalParam.UnbondingTime)
	assert.Equal(t, uint64(10000), versionedGlobalParam.UnbondingFee)
	assert.Equal(t, uint64(300000), versionedGlobalParam.MaxStakingAmount)
	assert.Equal(t, uint64(30000), versionedGlobalParam.MinStakingAmount)
	assert.Equal(t, uint64(10000), versionedGlobalParam.MaxStakingTime)
	assert.Equal(t, uint64(100), versionedGlobalParam.MinStakingTime)
	assert.Equal(t, uint64(10), versionedGlobalParam.ConfirmationDepth)

	versionedGlobalParam2 := result[1]
	assert.Equal(t, uint64(1), versionedGlobalParam2.Version)
	assert.Equal(t, uint64(200), versionedGlobalParam2.ActivationHeight)
	assert.Equal(t, uint64(50000000), versionedGlobalParam2.StakingCap)
	assert.Equal(t, "01020304", versionedGlobalParam2.Tag)
	assert.Equal(t, 4, len(versionedGlobalParam2.CovenantPks))
	assert.Equal(t, uint64(2), versionedGlobalParam2.CovenantQuorum)
	assert.Equal(t, uint64(2000), versionedGlobalParam2.UnbondingTime)
	assert.Equal(t, uint64(20000), versionedGlobalParam2.UnbondingFee)
	assert.Equal(t, uint64(200000), versionedGlobalParam2.MaxStakingAmount)
	assert.Equal(t, uint64(30000), versionedGlobalParam2.MinStakingAmount)
	assert.Equal(t, uint64(20000), versionedGlobalParam2.MaxStakingTime)
	assert.Equal(t, uint64(200), versionedGlobalParam2.MinStakingTime)
	assert.Equal(t, uint64(10), versionedGlobalParam2.ConfirmationDepth)

	versionedGlobalParam3 := result[2]
	assert.Equal(t, uint64(2), versionedGlobalParam3.Version)
	assert.Equal(t, uint64(300), versionedGlobalParam3.ActivationHeight)
	assert.Equal(t, uint64(500), versionedGlobalParam3.CapHeight)
	assert.Equal(t, uint64(0), versionedGlobalParam3.StakingCap)

	versionedGlobalParam4 := result[3]
	assert.Equal(t, uint64(3), versionedGlobalParam4.Version)
	assert.Equal(t, uint64(400), versionedGlobalParam4.ActivationHeight)
	assert.Equal(t, uint64(1000), versionedGlobalParam4.CapHeight)
	assert.Equal(t, uint64(0), versionedGlobalParam4.StakingCap)
}
