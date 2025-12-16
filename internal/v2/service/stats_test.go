package v2service

import (
	"errors"
	"testing"
	"time"

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
		CoinMarketCap: coinmarketcap.NewClient("", 0*time.Second),
	}, &dbclients.DbClients{
		SharedDBClient:  dbShared,
		V1DBClient:      dbV1,
		V2DBClient:      dbV2,
		IndexerDBClient: dbIndexer,
	})
	require.NoError(t, err)

	s, err := New(sharedService, nil, nil)
	require.NoError(t, err)

	t.Run("Indexer DB failure - GetOverallStats", func(t *testing.T) {
		err := errors.New("indexer err")
		dbIndexer.On("GetOverallStats", ctx).Return(nil, err).Once()

		resp, respErr := s.GetOverallStats(ctx)
		assert.Equal(t, types.NewInternalServiceError(err), respErr)
		assert.Nil(t, resp)
	})
	t.Run("Indexer DB failure - CountFinalityProviders", func(t *testing.T) {
		dbIndexer.On("GetOverallStats", ctx).Return(&indexerdbmodel.IndexerStatsDocument{}, nil).Once()
		err := errors.New("indexer count err")
		dbIndexer.On("CountFinalityProvidersByStatus", ctx).Return(nil, err).Once()

		resp, respErr := s.GetOverallStats(ctx)
		assert.Equal(t, types.NewInternalServiceError(err), respErr)
		assert.Nil(t, resp)
	})
	t.Run("V1 DB failure", func(t *testing.T) {
		dbIndexer.On("GetOverallStats", ctx).Return(&indexerdbmodel.IndexerStatsDocument{}, nil).Once()
		dbIndexer.On("CountFinalityProvidersByStatus", ctx).Return(map[indexerdbmodel.FinalityProviderState]uint64{}, nil).Once()
		err := errors.New("v1 err")
		dbV1.On("GetOverallStats", ctx).Return(nil, err).Once()

		resp, respErr := s.GetOverallStats(ctx)
		assert.Equal(t, types.NewInternalServiceError(err), respErr)
		assert.Nil(t, resp)
	})
	t.Run("Ok with GetLatestPrice failure", func(t *testing.T) {
		dbIndexer.On("GetOverallStats", ctx).Return(&indexerdbmodel.IndexerStatsDocument{
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
		CoinMarketCap: coinmarketcap.NewClient("", 0*time.Second),
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

func Test_calculateUserCoStakingAPR(t *testing.T) {
	s := &V2Service{}

	t.Run("zero satoshis returns zero", func(t *testing.T) {
		apr := s.calculateUserCoStakingAPR(
			0,
			1000000,
			1000000000,
			1000,
			1000000,
			50000,
			0.5,
		)
		assert.Equal(t, float64(0), apr)
	})

	t.Run("zero global score returns zero", func(t *testing.T) {
		apr := s.calculateUserCoStakingAPR(
			100000000,
			1000000,
			0,
			1000,
			1000000,
			50000,
			0.5,
		)
		assert.Equal(t, float64(0), apr)
	})

	t.Run("calculates correctly with valid inputs", func(t *testing.T) {
		apr := s.calculateUserCoStakingAPR(
			100000000,
			100000000,
			1000000000,
			1000,
			1000000,
			50000,
			0.5,
		)
		assert.Greater(t, apr, float64(0))
	})

	t.Run("no BABY staked returns zero co-staking APR", func(t *testing.T) {
		apr := s.calculateUserCoStakingAPR(
			100000000,
			0,
			1000000000,
			1000,
			1000000,
			50000,
			0.5,
		)
		assert.Equal(t, float64(0), apr)
	})
}

func Test_calculateBoostCoStakingAPR(t *testing.T) {
	s := &V2Service{}

	t.Run("zero satoshis returns zero", func(t *testing.T) {
		apr := s.calculateBoostCoStakingAPR(
			0,
			1000000,
			1000000000,
			1000,
			1000000,
			50000,
			0.5,
		)
		assert.Equal(t, float64(0), apr)
	})

	t.Run("zero global score returns zero", func(t *testing.T) {
		apr := s.calculateBoostCoStakingAPR(
			100000000,
			1000000,
			0,
			1000,
			1000000,
			50000,
			0.5,
		)
		assert.Equal(t, float64(0), apr)
	})

	t.Run("boost APR >= current APR", func(t *testing.T) {
		satoshisStaked := int64(100000000)
		ubbnStaked := int64(50000000)
		globalTotalScore := int64(1000000000)
		scoreRatio := int64(1000)
		totalCoStakingRewardSupply := float64(1000000)
		btcPrice := float64(50000)
		babyPrice := float64(0.5)

		currentAPR := s.calculateUserCoStakingAPR(
			satoshisStaked, ubbnStaked, globalTotalScore, scoreRatio,
			totalCoStakingRewardSupply, btcPrice, babyPrice,
		)
		boostAPR := s.calculateBoostCoStakingAPR(
			satoshisStaked, ubbnStaked, globalTotalScore, scoreRatio,
			totalCoStakingRewardSupply, btcPrice, babyPrice,
		)

		assert.GreaterOrEqual(t, boostAPR, currentAPR)
	})
}
