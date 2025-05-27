package v2service

import (
	"errors"
	"testing"

	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
	"github.com/babylonlabs-io/staking-api-service/tests/mocks"
	cmc "github.com/miguelmota/go-coinmarketcap/pro/v1"
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
	s, err := New(&service.Service{
		DbClients: &dbclients.DbClients{
			SharedDBClient:  dbShared,
			V1DBClient:      dbV1,
			V2DBClient:      dbV2,
			IndexerDBClient: dbIndexer,
		},
		Clients: &clients.Clients{
			CoinMarketCap: cmc.NewClient(nil),
		},
	})
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
		dbIndexer.On("GetFinalityProviders", ctx).Return(nil, err).Once()

		resp, respErr := s.GetOverallStats(ctx)
		assert.Equal(t, types.NewInternalServiceError(err), respErr)
		assert.Nil(t, resp)
	})
	t.Run("V1 DB failure", func(t *testing.T) {
		// we pass zero value as 1st return value which is ok - we won't use its values anyway
		dbV2.On("GetOverallStats", ctx).Return(&v2dbmodel.V2OverallStatsDocument{}, nil).Once()
		// note that first return value (finality providers) is nil which is ok (iteration over nil slice is valid)
		dbIndexer.On("GetFinalityProviders", ctx).Return(nil, nil).Once()
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
		dbIndexer.On("GetFinalityProviders", ctx).Return(nil, nil).Once()
		dbV1.On("GetOverallStats", ctx).Return(&v1dbmodel.OverallStatsDocument{}, nil).Once()
		err := errors.New("db err")
		// this error shouldn't trigger error in GetOverallStats method
		dbShared.On("GetLatestPrice", ctx, mock.Anything).Return(float64(0), err).Once()

		resp, respErr := s.GetOverallStats(ctx)
		assert.Nil(t, respErr)
		assert.Equal(t, &OverallStatsPublic{
			ActiveTvl:      777,
			TotalActiveTvl: 777,
		}, resp)
	})
}
