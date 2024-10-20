package tests

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-api-service/cmd/staking-api-service/scripts"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
	"github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/stretchr/testify/assert"
)

func TestReplayUnprocessableMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	activeStakingEvents := buildActiveStakingEvent(t, 1)
	activeStakingEvent := activeStakingEvents[0]

	data, err := json.Marshal(activeStakingEvent)
	assert.NoError(t, err, "marshal events should not return an error")

	doc := string(data)

	testutils.InjectDbDocument(
		testServer.Config, dbmodel.V1UnprocessableMsgCollection,
		dbmodel.NewUnprocessableMessageDocument(doc, "receipt"),
	)
	dbClients, _ := testutils.DirectDbConnection(testServer.Config)
	defer dbClients.MongoClient.Disconnect(ctx)
	dbclient, err := dbclient.New(ctx, dbClients.MongoClient, testServer.Config.Db)
	assert.NoError(t, err, "creating db client should not return an error")

	scripts.ReplayUnprocessableMessages(ctx, testServer.Config, testServer.Queues, dbclient)

	time.Sleep(3 * time.Second)

	url := testServer.Server.URL + stakerDelegations + "?staker_btc_pk=" + activeStakingEvent.StakerPkHex
	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to delegations by staker pk should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP response status should be OK")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var responseJSON handler.PublicResponse[[]client.ActiveStakingEvent]
	err = json.Unmarshal(body, &responseJSON)
	assert.NoError(t, err, "unmarshal response JSON should not return an error")

	// Verify the response contains expected fields
	expectedFields := []string{
		"StakingTxHashHex",
		"IsOverflow",
		"StakingTxHex",
		"StakingTimeLock",
		"StakingOutputIndex",
		"StakingStartTimestamp",
		"StakingStartHeight",
		"StakingValue",
		"FinalityProviderPkHex",
		"StakerPkHex",
	}

	assert.Greater(t, len(responseJSON.Data), 0, "'data' array should not be empty")

	for _, item := range responseJSON.Data {
		itemMap := map[string]interface{}{
			"StakingTxHashHex":      item.StakingTxHashHex,
			"IsOverflow":            item.IsOverflow,
			"StakingTxHex":          item.StakingTxHex,
			"StakingTimeLock":       item.StakingTimeLock,
			"StakingOutputIndex":    item.StakingOutputIndex,
			"StakingStartTimestamp": item.StakingStartTimestamp,
			"StakingStartHeight":    item.StakingStartHeight,
			"StakingValue":          item.StakingValue,
			"FinalityProviderPkHex": item.FinalityProviderPkHex,
			"StakerPkHex":           item.StakerPkHex,
		}

		for _, field := range expectedFields {
			_, exists := itemMap[field]
			assert.True(t, exists, "response should contain field %s", field)
		}
	}
}
