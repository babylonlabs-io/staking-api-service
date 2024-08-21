package scripts_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/babylonlabs-io/staking-api-service/cmd/staking-api-service/scripts"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"github.com/babylonlabs-io/staking-api-service/internal/utils"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
	"github.com/stretchr/testify/assert"
)

func createNewDelegationDocuments(cfg *config.Config, numOfDocs int) []*model.DelegationDocument {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var delegationDocuments []*model.DelegationDocument
	opts := &testutils.TestActiveEventGeneratorOpts{
		NumOfEvents: 1, // a single event to make sure it's always unique
	}
	for i := 0; i < numOfDocs; i++ {
		activeStakingEvenets := testutils.GenerateRandomActiveStakingEvents(r, opts)
		for _, event := range activeStakingEvenets {
			doc := &model.DelegationDocument{
				StakingTxHashHex:      event.StakingTxHashHex,
				StakerPkHex:           event.StakerPkHex,
				FinalityProviderPkHex: event.FinalityProviderPkHex,
				StakingValue:          event.StakingValue,
				State:                 types.Active,
				StakingTx: &model.TimelockTransaction{
					TxHex:          event.StakingTxHex,
					OutputIndex:    event.StakingOutputIndex,
					StartTimestamp: event.StakingStartTimestamp,
					StartHeight:    event.StakingStartHeight,
					TimeLock:       event.StakingStartHeight,
				},
				IsOverflow: event.IsOverflow,
			}
			delegationDocuments = append(delegationDocuments, doc)
		}
	}
	return delegationDocuments
}

func TestBackfillAddressesBasedOnPubKeys(t *testing.T) {
	cfg := testutils.LoadTestConfig()
	ctx := context.Background()
	// Clean the database
	testutils.SetupTestDB(*cfg)
	// inject some data
	docs := createNewDelegationDocuments(cfg, int(cfg.Db.MaxPaginationLimit)+1)
	for _, doc := range docs {
		testutils.InjectDbDocument(
			cfg,
			model.DelegationCollection,
			doc,
		)
	}

	// sleep for a while to let the data be inserted
	time.Sleep(5 * time.Second)
	err := scripts.BackfillPubkeyAddressesMappings(ctx, cfg)
	assert.Nil(t, err)
	// check if the data is inserted
	results, err := testutils.InspectDbDocuments[model.PkAddressMapping](
		cfg,
		model.PkAddressMappingsCollection,
	)
	assert.Nil(t, err)
	// find the num of unique staker pks from the docs
	stakerPks := make(map[string]struct{})
	for _, doc := range docs {
		stakerPks[doc.StakerPkHex] = struct{}{}
	}
	// check if the number of unique staker pks is equal to the number of results
	assert.Equal(t, len(stakerPks), len(results))
	// check if the data is inserted correctly
	for _, result := range results {
		_, ok := stakerPks[result.PkHex]
		assert.True(t, ok)
		addresses, err := utils.DeriveAddressesFromNoCoordPk(
			result.PkHex, cfg.Server.BTCNetParam,
		)
		assert.Nil(t, err)
		assert.Equal(t, result.Taproot, addresses.Taproot)
		assert.Equal(t, result.NativeSegwitOdd, addresses.NativeSegwitOdd)
		assert.Equal(t, result.NativeSegwitEven, addresses.NativeSegwitEven)
	}

	// Run the script again, the result should be the same as it does not
	// change the existing data
	err = scripts.BackfillPubkeyAddressesMappings(ctx, cfg)
	assert.Nil(t, err)
	results2, err := testutils.InspectDbDocuments[model.PkAddressMapping](
		cfg,
		model.PkAddressMappingsCollection,
	)
	assert.Nil(t, err)
	assert.Equal(t, len(results), len(results2))
}
