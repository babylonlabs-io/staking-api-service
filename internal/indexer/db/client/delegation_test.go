//go:build integration

package indexerdbclient_test

import (
	"context"
	"encoding/json"
	model "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/pkg"
	"github.com/cometbft/cometbft/libs/os"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestDelegations(t *testing.T) {
	ctx := context.Background()

	const (
		stakerPKHex = "1cb8800d69c22978cbfe4874e132f6a0735880d49b3fecf2543e50b8b16fde57"
		bbnAddress1 = "bbn1vp2grtx8yjlj7zpkjeaf5pf6cquym2c88j92p2"
	)

	fixtures := loadTestData(t, "btc_delegation_details.json")
	// in order to test pagination limit must be less than the number of fixtures
	require.Less(t, maxPaginationLimit, len(fixtures))

	collection := testDB.Client.Database(testDB.Cfg.DbName).Collection(model.BTCDelegationDetailsCollection)
	_, err := collection.InsertMany(ctx, fixtures)
	require.NoError(t, err)

	t.Run("no babylon_address", func(t *testing.T) {
		// ids of found records
		ids := make(map[string]bool, len(fixtures))

		var token string
		for {
			result, err := testDB.GetDelegations(ctx, stakerPKHex, nil, token)
			require.NoError(t, err)

			// for simplicity we just collect ids of found records in ids map
			for _, res := range result.Data {
				ids[res.StakingTxHashHex] = true
			}

			token = result.PaginationToken
			if token == "" {
				break
			}
		}

		// check that number of found records is equal of stored ones
		assert.Equal(t, len(fixtures), len(ids))
	})
	t.Run("with babylon_address", func(t *testing.T) {
		var token string
		var numOfFoundRecords int

		for {
			result, err := testDB.GetDelegations(ctx, stakerPKHex, pkg.Ptr(bbnAddress1), token)
			require.NoError(t, err)

			numOfFoundRecords += len(result.Data)

			token = result.PaginationToken
			if token == "" {
				break
			}
		}

		// for now 3 is just hardcoded number of records with this babylon address
		assert.Equal(t, 3, numOfFoundRecords)
	})

}

func loadTestData(t *testing.T, filename string) []any {
	buff, err := os.ReadFile(filepath.Join("testdata", filename))
	require.NoError(t, err)

	var fixtures []any
	err = json.Unmarshal(buff, &fixtures)
	require.NoError(t, err)

	return fixtures
}
