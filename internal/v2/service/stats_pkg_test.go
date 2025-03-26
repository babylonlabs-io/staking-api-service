package v2service

import (
	"context"
	"errors"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	v2dbclient "github.com/babylonlabs-io/staking-api-service/internal/v2/db/client"
	"github.com/babylonlabs-io/staking-api-service/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetActiveStakersCount(t *testing.T) {
	ctx := context.Background()

	newService := func(db v2dbclient.V2DBClient) *V2Service {
		cfg := &config.Config{}
		dbClients := &dbclients.DbClients{
			V2DBClient: db,
		}

		sharedService, err := service.New(ctx, cfg, nil, nil, nil, dbClients)
		require.NoError(t, err)

		v2, err := New(sharedService)
		require.NoError(t, err)

		return v2
	}

	t.Run("ok", func(t *testing.T) {
		baseCount := int64(777)

		db := mocks.NewV2DBClient(t)
		// it's important to call .Once() in the end!
		db.On("GetActiveStakersCount", ctx).Return(baseCount, nil).Once()

		s := newService(db)

		const ttl = time.Minute
		doc, err := s.getActiveStakersCount(ctx, ttl)
		require.NoError(t, err)
		assert.Equal(t, baseCount, doc)

		doc, err = s.getActiveStakersCount(ctx, ttl)
		require.NoError(t, err)
		assert.Equal(t, baseCount, doc)
	})
	t.Run("ok (expired)", func(t *testing.T) {
		// this subtest tests 2 things:
		// 1. Cache expiration takes place
		// 2. TTL of stored item is not updated with each access
		//    otherwise if there is 1 request per TTL period cached item won't expire at all
		baseCount := int64(888)

		db := mocks.NewV2DBClient(t)
		db.On("GetActiveStakersCount", ctx).Return(baseCount, nil).Twice() // only 2 db accesses are expected

		s := newService(db)

		const ttl = time.Second
		const halfTTL = ttl / 2

		// 1. we reach db and store item in cache
		doc, err := s.getActiveStakersCount(ctx, ttl)
		require.NoError(t, err)
		assert.Equal(t, baseCount, doc)

		// 2. we first wait for half of TTL (explained later)
		// and then call the function again (this time it will use cache only)
		time.Sleep(halfTTL)
		doc, err = s.getActiveStakersCount(ctx, ttl)
		require.NoError(t, err)
		assert.Equal(t, baseCount, doc)

		// 3. again we wait for half of TTL (by that time TTL + some additional time will pass which means
		// cached item must be expired already) and call the function again (this time it will reach db)
		time.Sleep(halfTTL)
		doc, err = s.getActiveStakersCount(ctx, ttl)
		require.NoError(t, err)
		assert.Equal(t, baseCount, doc)
	})
	t.Run("error", func(t *testing.T) {
		baseErr := errors.New("some error")

		db := mocks.NewV2DBClient(t)
		db.On("GetActiveStakersCount", ctx).Return(int64(0), baseErr).Once()

		s := newService(db)

		const ttl = time.Minute
		doc, err := s.getActiveStakersCount(ctx, ttl)
		require.ErrorIs(t, err, baseErr)
		assert.Zero(t, doc)
	})
}
