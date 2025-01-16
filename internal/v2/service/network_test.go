package v2service

import (
	"testing"
	"context"
	"github.com/stretchr/testify/require"
	"fmt"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
)

func TestGetNetworkInfo(t *testing.T) {
	ctx := context.Background() // todo(Kirill) replace with t.Context() after go 1.24 release
	t.Run("BBN params are sorted", func(t *testing.T) {
		service, err := New(ctx, nil, nil, &dbclients.DbClients{
			StakingMongoClient: nil,
			IndexerMongoClient: nil,
			SharedDBClient:     nil,
			V1DBClient:         nil,
			V2DBClient:         nil,
			IndexerDBClient:    nil,
		})
		require.NoError(t, err)

		resp, rpcErr := service.GetNetworkInfo(ctx)
		require.Nil(t, rpcErr)

		fmt.Println("RESP", resp)
	})
}
