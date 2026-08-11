//go:build integration

package indexerdbclient_test

import (
	"testing"

	model "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFinalityProviderStatsPaginated(t *testing.T) {
	ctx := t.Context()

	// Create test fixtures with different TVL values
	fixtures := []any{
		&model.IndexerFinalityProviderStatsDocument{
			FpBtcPkHex:        "aaa111",
			ActiveTvl:         1000,
			ActiveDelegations: 10,
			LastUpdated:       1234567890,
		},
		&model.IndexerFinalityProviderStatsDocument{
			FpBtcPkHex:        "bbb222",
			ActiveTvl:         5000,
			ActiveDelegations: 50,
			LastUpdated:       1234567891,
		},
		&model.IndexerFinalityProviderStatsDocument{
			FpBtcPkHex:        "ccc333",
			ActiveTvl:         3000,
			ActiveDelegations: 30,
			LastUpdated:       1234567892,
		},
		&model.IndexerFinalityProviderStatsDocument{
			FpBtcPkHex:        "ddd444",
			ActiveTvl:         5000, // Same TVL as bbb222 to test tiebreaker
			ActiveDelegations: 55,
			LastUpdated:       1234567893,
		},
		&model.IndexerFinalityProviderStatsDocument{
			FpBtcPkHex:        "eee555",
			ActiveTvl:         2000,
			ActiveDelegations: 20,
			LastUpdated:       1234567894,
		},
	}

	// In order to test pagination, limit must be less than the number of fixtures
	require.Less(t, maxPaginationLimit, len(fixtures))

	collection := testDB.Client.Database(testDB.Cfg.DbName).Collection(model.IndexerFinalityProviderStatsCollection)
	_, err := collection.InsertMany(ctx, fixtures)
	require.NoError(t, err)
	defer resetDatabase(t)

	t.Run("sorted by active_tvl descending with tiebreaker", func(t *testing.T) {
		var allResults []*model.IndexerFinalityProviderStatsDocument
		var token string

		// Paginate through all results
		for {
			result, err := testDB.GetFinalityProviderStatsPaginated(ctx, token)
			require.NoError(t, err)

			allResults = append(allResults, result.Data...)

			token = result.PaginationToken
			if token == "" {
				break
			}
		}

		// Check that all records were fetched
		assert.Equal(t, len(fixtures), len(allResults))

		// Verify sorting: active_tvl descending, then _id descending for ties
		// Expected order: bbb222(5000), ddd444(5000), ccc333(3000), eee555(2000), aaa111(1000)
		// For same TVL, _id descending: ddd444 > bbb222
		assert.Equal(t, uint64(5000), allResults[0].ActiveTvl)
		assert.Equal(t, "ddd444", allResults[0].FpBtcPkHex)

		assert.Equal(t, uint64(5000), allResults[1].ActiveTvl)
		assert.Equal(t, "bbb222", allResults[1].FpBtcPkHex)

		assert.Equal(t, uint64(3000), allResults[2].ActiveTvl)
		assert.Equal(t, "ccc333", allResults[2].FpBtcPkHex)

		assert.Equal(t, uint64(2000), allResults[3].ActiveTvl)
		assert.Equal(t, "eee555", allResults[3].FpBtcPkHex)

		assert.Equal(t, uint64(1000), allResults[4].ActiveTvl)
		assert.Equal(t, "aaa111", allResults[4].FpBtcPkHex)

		// Verify all TVLs are in descending order (allowing ties)
		for i := 0; i < len(allResults)-1; i++ {
			assert.GreaterOrEqual(t, allResults[i].ActiveTvl, allResults[i+1].ActiveTvl,
				"active_tvl should be sorted in descending order")
		}
	})

	t.Run("empty results with no data", func(t *testing.T) {
		resetDatabase(t)

		result, err := testDB.GetFinalityProviderStatsPaginated(ctx, "")
		require.NoError(t, err)
		assert.Empty(t, result.Data)
		assert.Empty(t, result.PaginationToken)
	})

	t.Run("invalid pagination token", func(t *testing.T) {
		_, err := testDB.GetFinalityProviderStatsPaginated(ctx, "invalid-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid pagination token")
	})
}
