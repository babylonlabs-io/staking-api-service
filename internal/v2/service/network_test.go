package v2service

import (
	"context"
	"slices"
	"testing"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/babylonlabs-io/staking-api-service/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNetworkInfo(t *testing.T) {
	ctx := context.Background() // todo(Kirill) replace with t.Context() after go 1.24 release
	t.Run("BBN params are sorted", func(t *testing.T) {
		indexerDB := &mocks.IndexerDBClient{}
		defer indexerDB.AssertExpectations(t)

		bbnStakingParams := []*indexertypes.BbnStakingParams{
			// other values are not important in this test, focus only on version
			{Version: 33},
			{Version: 0},
			{Version: 7},
			{Version: 9},
		}
		indexerDB.On("GetBbnStakingParams", ctx).Return(bbnStakingParams, nil).Once()
		indexerDB.On("GetBtcCheckpointParams", ctx).Return(nil, nil).Once()

		cfg := &config.Config{}
		dbClients := &dbclients.DbClients{
			IndexerDBClient: indexerDB,
		}

		sharedService, err := service.New(cfg, nil, nil, nil, dbClients)
		require.NoError(t, err)

		service, err := New(sharedService)
		require.NoError(t, err)

		resp, rpcErr := service.GetNetworkInfo(ctx)
		require.Nil(t, rpcErr)

		var versions []uint32
		for _, param := range resp.Params.Bbn {
			versions = append(versions, param.Version)
		}

		assert.NotEmpty(t, versions)
		assert.True(t, slices.IsSorted(versions))
	})
}
