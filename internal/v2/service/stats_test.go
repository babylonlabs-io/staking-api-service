package v2service

import (
	"errors"
	"testing"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients/coinmarketcap"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
	"github.com/babylonlabs-io/staking-api-service/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_GetOverallStats(t *testing.T) {
	ctx := t.Context()

	dbShared := mocks.NewDBClient(t)
	dbV1 := mocks.NewV1DBClient(t)
	dbV2 := mocks.NewV2DBClient(t)
	dbIndexer := mocks.NewIndexerDBClient(t)

	sharedService, err := service.New(&config.Config{}, nil, nil, &clients.Clients{
		CoinMarketCap: coinmarketcap.NewClient("", 0),
	}, &dbclients.DbClients{
		SharedDBClient:  dbShared,
		V1DBClient:      dbV1,
		V2DBClient:      dbV2,
		IndexerDBClient: dbIndexer,
	})
	require.NoError(t, err)

	s, err := New(sharedService, nil, nil)
	require.NoError(t, err)

	t.Run("V2 DB failure", func(t *testing.T) {
		err := errors.New("v2 err")
		dbV2.On("GetOverallStats", ctx).Return(nil, err).Once()

		resp, respErr := s.GetOverallStats(ctx)
		assert.Equal(t, types.NewInternalServiceError(err), respErr)
		assert.Nil(t, resp)
	})
	t.Run("Indexer DB failure", func(t *testing.T) {
		// we pass zero value as 1st return value which is ok - we won't use its values anyway
		dbV2.On("GetOverallStats", ctx).Return(&v2dbmodel.V2OverallStatsDocument{}, nil).Once()
		err := errors.New("indexer err")
		dbIndexer.On("CountFinalityProvidersByStatus", ctx).Return(nil, err).Once()

		resp, respErr := s.GetOverallStats(ctx)
		assert.Equal(t, types.NewInternalServiceError(err), respErr)
		assert.Nil(t, resp)
	})
	t.Run("V1 DB failure", func(t *testing.T) {
		// we pass zero value as 1st return value which is ok - we won't use its values anyway
		dbV2.On("GetOverallStats", ctx).Return(&v2dbmodel.V2OverallStatsDocument{}, nil).Once()
		dbIndexer.On("CountFinalityProvidersByStatus", ctx).Return(map[indexerdbmodel.FinalityProviderState]uint64{}, nil).Once()
		err := errors.New("v1 err")
		dbV1.On("GetOverallStats", ctx).Return(nil, err).Once()

		resp, respErr := s.GetOverallStats(ctx)
		assert.Equal(t, types.NewInternalServiceError(err), respErr)
		assert.Nil(t, resp)
	})
	t.Run("Ok with GetLatestPrice failure", func(t *testing.T) {
		dbV2.On("GetOverallStats", ctx).Return(&v2dbmodel.V2OverallStatsDocument{
			ActiveTvl: 777, // here is important to pass non-zero tvl so it triggers staking BTC calculation
		}, nil).Once()
		dbIndexer.On("CountFinalityProvidersByStatus", ctx).Return(map[indexerdbmodel.FinalityProviderState]uint64{}, nil).Once()
		dbV1.On("GetOverallStats", ctx).Return(&v1dbmodel.OverallStatsDocument{}, nil).Once()
		err := errors.New("db err")
		// this error shouldn't trigger error in GetOverallStats method
		// We now call GetLatestPrice multiple times: for BTC APR calculation and for max staking APR calculation
		dbShared.On("GetLatestPrice", ctx, mock.Anything).Return(float64(0), err).Times(3)

		resp, respErr := s.GetOverallStats(ctx)
		assert.Nil(t, respErr)
		assert.Equal(t, &OverallStatsPublic{
			ActiveTvl:      777,
			TotalActiveTvl: 777,
		}, resp)
	})
}

func Test_ProcessActiveDelegationStats(t *testing.T) {
	ctx := t.Context()

	dbShared := mocks.NewDBClient(t)
	dbV1 := mocks.NewV1DBClient(t)
	dbV2 := mocks.NewV2DBClient(t)
	dbIndexer := mocks.NewIndexerDBClient(t)

	sharedService, err := service.New(&config.Config{}, nil, nil, &clients.Clients{
		CoinMarketCap: coinmarketcap.NewClient("", 0),
	}, &dbclients.DbClients{
		SharedDBClient:  dbShared,
		V1DBClient:      dbV1,
		V2DBClient:      dbV2,
		IndexerDBClient: dbIndexer,
	})
	require.NoError(t, err)

	s, err := New(sharedService, nil, nil)
	require.NoError(t, err)

	t.Run("V2 DB failure", func(t *testing.T) {
		stakingTxHashHex := `19caaf9dcf7be81120a503b8e007189ecee53e5912c8fa542b187224ce45000a`
		stakerPkHex := `21d17b47e1d763f478cba5c414b7adf2778fa4ff6a5ba3d79f08f7a494781e06`
		err := errors.New("some error")

		dbV2.On("GetOrCreateStatsLock", ctx, stakingTxHashHex, "active").Return(nil, err).Once()
		statsErr := s.ProcessActiveDelegationStats(ctx, stakingTxHashHex, stakerPkHex, nil, 30)
		require.Error(t, statsErr)
	})
	t.Run("Active stats", func(t *testing.T) {
		stakingTxHashHex := `19caaf9dcf7be81120a503b8e007189ecee53e5912c8fa542b187224ce45000a`
		stakerPkHex := `21d17b47e1d763f478cba5c414b7adf2778fa4ff6a5ba3d79f08f7a494781e06`
		amount := uint64(77)

		const (
			fp1ID = "fp1"
			fp2ID = "fp2"
		)

		locks := &v2dbmodel.V2StatsLockDocument{
			Id:                    stakingTxHashHex + ":active",
			OverallStats:          true,
			StakerStats:           true,
			FinalityProviderStats: true,
		}
		dbV2.On("GetOrCreateStatsLock", ctx, stakingTxHashHex, "active").Return(locks, nil).Once()

		fpBtcPkHexes := []string{fp1ID, fp2ID}
		statsErr := s.ProcessActiveDelegationStats(ctx, stakingTxHashHex, stakerPkHex, fpBtcPkHexes, amount)
		require.Error(t, statsErr)
	})
}
