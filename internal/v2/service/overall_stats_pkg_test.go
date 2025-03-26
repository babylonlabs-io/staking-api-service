package v2service

import (
	"context"
	"errors"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
	"github.com/babylonlabs-io/staking-api-service/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestOverallStats(t *testing.T) {
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		baseDoc := &v2dbmodel.V2OverallStatsDocument{
			Id:                "some_id",
			ActiveTvl:         10,
			ActiveDelegations: 20,
		}

		db := mocks.NewV2DBClient(t)
		// it's important to call .Once() in the end!
		db.On("GetOverallStats", ctx).Return(baseDoc, nil).Once()

		s := newOverallStatsService(db, time.Minute)

		doc, err := s.getOverallStatsFromDB(ctx)
		require.NoError(t, err)
		assert.Equal(t, baseDoc, doc)

		doc, err = s.getOverallStatsFromDB(ctx)
		require.NoError(t, err)
		assert.Equal(t, baseDoc, doc)
	})
	t.Run("ok (expired)", func(t *testing.T) {
		baseDoc := &v2dbmodel.V2OverallStatsDocument{
			Id:                "some_id",
			ActiveTvl:         111,
			ActiveDelegations: 222,
		}

		db := mocks.NewV2DBClient(t)
		db.On("GetOverallStats", ctx).Return(baseDoc, nil).Twice()

		const ttl = time.Second
		s := newOverallStatsService(db, ttl)

		doc, err := s.getOverallStatsFromDB(ctx)
		require.NoError(t, err)
		assert.Equal(t, baseDoc, doc)

		time.Sleep(ttl)

		doc, err = s.getOverallStatsFromDB(ctx)
		require.NoError(t, err)
		assert.Equal(t, baseDoc, doc)
	})
	t.Run("error", func(t *testing.T) {
		baseErr := errors.New("some error")

		db := mocks.NewV2DBClient(t)
		db.On("GetOverallStats", ctx).Return(nil, baseErr).Once()

		s := newOverallStatsService(db, time.Minute)

		doc, err := s.getOverallStatsFromDB(ctx)
		require.ErrorIs(t, err, baseErr)
		assert.Nil(t, doc)
	})
}
