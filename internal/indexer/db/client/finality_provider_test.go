//go:build integration

package indexerdbclient_test

import (
	"testing"

	model "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFinalityProvidersByPks(t *testing.T) {
	ctx := t.Context()

	fixtures := []any{
		&model.IndexerFinalityProviderDetails{
			BtcPk:          "aaa111",
			BabylonAddress: "bbn1address1",
			Commission:     "0.05",
			State:          model.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_ACTIVE,
			Description: model.Description{
				Moniker: "FP1",
			},
		},
		&model.IndexerFinalityProviderDetails{
			BtcPk:          "bbb222",
			BabylonAddress: "bbn1address2",
			Commission:     "0.10",
			State:          model.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_ACTIVE,
			Description: model.Description{
				Moniker: "FP2",
			},
		},
		&model.IndexerFinalityProviderDetails{
			BtcPk:          "ccc333",
			BabylonAddress: "bbn1address3",
			Commission:     "0.15",
			State:          model.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_INACTIVE,
			Description: model.Description{
				Moniker: "FP3",
			},
		},
	}

	collection := testDB.Client.Database(testDB.Cfg.DbName).Collection(model.FinalityProviderDetailsCollection)
	_, err := collection.InsertMany(ctx, fixtures)
	require.NoError(t, err)
	defer resetDatabase(t)

	t.Run("fetch multiple finality providers", func(t *testing.T) {
		pks := []string{"aaa111", "bbb222"}
		results, err := testDB.GetFinalityProvidersByPks(ctx, pks)
		require.NoError(t, err)

		assert.Len(t, results, 2)

		// Create a map for easier lookup
		resultMap := make(map[string]*model.IndexerFinalityProviderDetails)
		for _, r := range results {
			resultMap[r.BtcPk] = r
		}

		assert.Contains(t, resultMap, "aaa111")
		assert.Contains(t, resultMap, "bbb222")
		assert.Equal(t, "FP1", resultMap["aaa111"].Description.Moniker)
		assert.Equal(t, "FP2", resultMap["bbb222"].Description.Moniker)
	})

	t.Run("fetch with case insensitivity", func(t *testing.T) {
		// Test lowercase conversion - uppercase input should still find lowercase stored keys
		pks := []string{"AAA111", "BBB222"}
		results, err := testDB.GetFinalityProvidersByPks(ctx, pks)
		require.NoError(t, err)

		assert.Len(t, results, 2)

		resultMap := make(map[string]*model.IndexerFinalityProviderDetails)
		for _, r := range results {
			resultMap[r.BtcPk] = r
		}

		assert.Contains(t, resultMap, "aaa111")
		assert.Contains(t, resultMap, "bbb222")
	})

	t.Run("empty input returns empty slice", func(t *testing.T) {
		results, err := testDB.GetFinalityProvidersByPks(ctx, []string{})
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("non-existent public keys", func(t *testing.T) {
		pks := []string{"nonexistent1", "nonexistent2"}
		results, err := testDB.GetFinalityProvidersByPks(ctx, pks)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("mixed existent and non-existent keys", func(t *testing.T) {
		pks := []string{"aaa111", "nonexistent", "ccc333"}
		results, err := testDB.GetFinalityProvidersByPks(ctx, pks)
		require.NoError(t, err)

		assert.Len(t, results, 2)

		resultMap := make(map[string]*model.IndexerFinalityProviderDetails)
		for _, r := range results {
			resultMap[r.BtcPk] = r
		}

		assert.Contains(t, resultMap, "aaa111")
		assert.Contains(t, resultMap, "ccc333")
		assert.NotContains(t, resultMap, "nonexistent")
	})
}
