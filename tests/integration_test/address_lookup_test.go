package tests

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	v1dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
	"github.com/stretchr/testify/assert"
)

const (
	stakerPkLookupPath = "/v1/staker/pubkey-lookup"
)

func FuzzTestPkAddressesMapping(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 3)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		opts := &testutils.TestActiveEventGeneratorOpts{
			NumOfEvents: testutils.RandomPositiveInt(r, 5),
			Stakers:     testutils.GeneratePks(5),
		}
		activeStakingEvents := testutils.GenerateRandomActiveStakingEvents(r, opts)
		var stakerPks []string
		for _, event := range activeStakingEvents {
			stakerPks = append(stakerPks, event.StakerPkHex)
		}
		testServer := setupTestServer(t, nil)
		defer testServer.Close()
		sendTestMessage(
			testServer.Queues.V1QueueClient.ActiveStakingQueueClient, activeStakingEvents,
		)
		time.Sleep(5 * time.Second)

		// Test the API
		pks := []string{}
		for _, event := range activeStakingEvents {
			pks = append(pks, event.StakerPkHex)
		}
		pks = uniqueStrings(pks)
		// randomly convert that into addresses with different types
		addresses := []string{}
		for _, pk := range pks {
			addr, err := utils.DeriveAddressesFromNoCoordPk(
				pk, testServer.Config.Server.BTCNetParam,
			)
			assert.NoError(t, err, "deriving addresses from public key should not fail")
			// Pick a random address type
			addresses = append(addresses, pickRandomAddress(r, addr))
		}
		result := performLookupRequest(t, testServer, addresses)
		assert.Equal(t, len(pks), len(result), "expected the same number of results")
		for _, addr := range addresses {
			resultPk, ok := result[addr]
			assert.True(t, ok, "expected the result to contain the address")
			assert.Contains(t, pks, resultPk, "expected the result to contain the public key")
		}

		// fetch with non-existent addresses
		nonExistPks := testutils.GeneratePks(5)
		nonExistentAddresses := []string{}
		for _, pk := range nonExistPks {
			addr, err := utils.DeriveAddressesFromNoCoordPk(
				pk, testServer.Config.Server.BTCNetParam,
			)
			assert.NoError(t, err, "deriving addresses from public key should not fail")
			// Pick a random address type
			nonExistentAddresses = append(nonExistentAddresses, pickRandomAddress(r, addr))
		}
		nonExistentResult := performLookupRequest(t, testServer, nonExistentAddresses)
		assert.Equal(t, 0, len(nonExistentResult))

		// fetch with a mix of existent and non-existent addresses
		mixedAddresses := append(addresses, nonExistentAddresses...)
		mixedResult := performLookupRequest(t, testServer, mixedAddresses)
		assert.Equal(t, len(addresses), len(mixedResult))
	})
}

func TestErrorForNoneTaprootOrNativeSegwitAddressLookup(t *testing.T) {
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	// Test the API with a non-taproot or native segwit address
	legacyAddress := "16o1TKSUWXy51oDpL5wbPxnezSGWC9rMPv"
	url := testServer.Server.URL + stakerPkLookupPath + "?" + "address=" + legacyAddress
	resp, err := http.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	NestedSegWitAddress := "3A2yqzgfxwwqxgse5rDTCQ2qmxZhMnfd5b"
	url = testServer.Server.URL + stakerPkLookupPath + "?" + "address=" + NestedSegWitAddress
	resp, err = http.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// Manually build an older version of the stats event which does not have the
// IsOverflow field and has a schema version less than 1
type event struct {
	EventType             int    `json:"event_type"`
	StakingTxHashHex      string `json:"staking_tx_hash_hex"`
	StakerPKHex           string `json:"staker_pk_hex"`
	FinalityProviderPKHex string `json:"finality_provider_pk_hex"`
	StakingValue          int64  `json:"staking_value"`
	State                 string `json:"state"`
}

func TestPkAddressMappingWorksForOlderStatsEventVersion(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	testServer := setupTestServer(t, nil)
	defer testServer.Close()
	tx, txHash, err := testutils.GenerateRandomTx(r, nil)
	assert.NoError(t, err, "failed to generate random tx")
	stakerPk, err := testutils.RandomPk()
	assert.NoError(t, err, "failed to generate random public key")
	fpPk, err := testutils.RandomPk()
	assert.NoError(t, err, "failed to generate random public key")
	del := &v1dbmodel.DelegationDocument{
		StakingTxHashHex:      txHash,
		StakerPkHex:           stakerPk,
		FinalityProviderPkHex: fpPk,
		StakingValue:          uint64(testutils.RandomAmount(r)),
		State:                 types.Active,
		StakingTx: &v1dbmodel.TimelockTransaction{
			TxHex:          tx.TxHash().String(),
			OutputIndex:    uint64(tx.TxOut[0].Value),
			StartTimestamp: time.Now().Unix(),
			StartHeight:    1,
			TimeLock:       100,
		},
		IsOverflow: false,
	}
	oldStatsMsg := &event{
		EventType:             5,
		StakingTxHashHex:      del.StakingTxHashHex,
		StakerPKHex:           del.StakerPkHex,
		FinalityProviderPKHex: del.FinalityProviderPkHex,
		StakingValue:          int64(del.StakingValue),
		State:                 string(del.State),
	}
	testutils.InjectDbDocument(
		testServer.Config, dbmodel.V1DelegationCollection, del,
	)
	sendTestMessage(
		testServer.Queues.V1QueueClient.StatsQueueClient, []event{*oldStatsMsg},
	)
	time.Sleep(5 * time.Second)

	// inspect the items in the database
	pkAddresses, err := testutils.InspectDbDocuments[dbmodel.PkAddressMapping](
		testServer.Config, dbmodel.V1PkAddressMappingsCollection,
	)
	assert.NoError(t, err, "failed to inspect the items in the database")
	assert.Equal(t, 1, len(pkAddresses), "expected only one item in the database")
	assert.Equal(t, del.StakerPkHex, pkAddresses[0].PkHex)
	// Check the address is correct
	addresses, err := utils.DeriveAddressesFromNoCoordPk(
		del.StakerPkHex, testServer.Config.Server.BTCNetParam,
	)
	assert.NoError(t, err, "failed to derive addresses from the public key")
	assert.Equal(t, addresses.Taproot, pkAddresses[0].Taproot)
	assert.Equal(t, addresses.NativeSegwitOdd, pkAddresses[0].NativeSegwitOdd)
	assert.Equal(t, addresses.NativeSegwitEven, pkAddresses[0].NativeSegwitEven)
}

func pickRandomAddress(r *rand.Rand, addresses *utils.SupportedAddress) string {
	choices := []string{
		addresses.Taproot, addresses.NativeSegwitEven, addresses.NativeSegwitOdd,
	}
	return choices[r.Intn(len(choices))]
}

func performLookupRequest(
	t *testing.T, testServer *TestServer, addresses []string,
) map[string]string {
	// form the addresses query as a string with format of `address=xyz&address=abc`
	query := ""
	for index, addr := range addresses {
		if index == len(addresses)-1 {
			query += "address=" + addr
		} else {
			query += "address=" + addr + "&"
		}
	}

	url := testServer.Server.URL + stakerPkLookupPath + "?" + query
	resp, err := http.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var response handler.PublicResponse[map[string]string]
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	return response.Data
}

func uniqueStrings(input []string) []string {
	// Create a map to track unique strings
	uniqueMap := make(map[string]struct{})

	// Iterate over the input slice and add each string to the map
	for _, str := range input {
		uniqueMap[str] = struct{}{}
	}

	// Create a slice to hold the unique strings
	var uniqueSlice []string
	for str := range uniqueMap {
		uniqueSlice = append(uniqueSlice, str)
	}

	return uniqueSlice
}
