package tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"github.com/babylonlabs-io/staking-api-service/internal/utils"
	"github.com/stretchr/testify/assert"
)

func FuzzTestPkAddressesMapping(f *testing.F) {
	attachRandomSeedsToFuzzer(f, 1)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		opts := &TestActiveEventGeneratorOpts{
			NumOfEvents: randomPositiveInt(r, 5),
			Stakers:     generatePks(t, 5),
		}
		activeStakingEvents := generateRandomActiveStakingEvents(t, r, opts)
		var stakerPks []string
		for _, event := range activeStakingEvents {
			stakerPks = append(stakerPks, event.StakerPkHex)
		}
		testServer := setupTestServer(t, nil)
		defer testServer.Close()
		sendTestMessage(
			testServer.Queues.ActiveStakingQueueClient, activeStakingEvents,
		)
		time.Sleep(5 * time.Second)

		// inspect the items in the database
		pkAddressMappings, err := inspectDbDocuments[model.PkAddressMapping](
			t, "pk_address_mappings",
		)
		assert.NoError(t, err, "failed to inspect the items in the database")
		// for each stakerPks, there should be a corresponding pkAddressMappings
		for _, pk := range stakerPks {
			found := false
			for _, mapping := range pkAddressMappings {
				if mapping.PkHex == pk {
					found = true
					assert.NotEmpty(t, mapping.Taproot)
					assert.NotEmpty(t, mapping.NativeSegwitOdd)
					assert.NotEmpty(t, mapping.NativeSegwitEven)
					break
				}
			}
			assert.True(t, found, "expected to find the staker pk in the database")
		}
	})
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
	tx, txHash, err := generateRandomTx(r)
	assert.NoError(t, err, "failed to generate random tx")
	stakerPk, err := randomPk()
	assert.NoError(t, err, "failed to generate random public key")
	fpPk, err := randomPk()
	assert.NoError(t, err, "failed to generate random public key")
	del := &model.DelegationDocument{
		StakingTxHashHex:      txHash,
		StakerPkHex:           stakerPk,
		FinalityProviderPkHex: fpPk,
		StakingValue:          uint64(randomAmount(r)),
		State:                 types.Active,
		StakingTx: &model.TimelockTransaction{
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
	injectDbDocuments(t, model.DelegationCollection, del)
	sendTestMessage(
		testServer.Queues.StatsQueueClient, []event{*oldStatsMsg},
	)
	time.Sleep(5 * time.Second)

	// inspect the items in the database
	pkAddresses, err := inspectDbDocuments[model.PkAddressMapping](
		t, "pk_address_mappings",
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
