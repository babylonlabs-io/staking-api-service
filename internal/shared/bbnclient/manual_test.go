//go:build manual

package bbnclient

import (
	"testing"
	"time"

	"fmt"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/pkg"
	"github.com/stretchr/testify/require"
)

func TestBBNClient(t *testing.T) {
	rpcAddr := pkg.Getenv("BABYLON_RPC_ADDR", "https://rpc.devnet.babylonlabs.io/")

	cl, err := New(&config.BBNConfig{
		RPCAddr: rpcAddr,
		Timeout: time.Second,
	})
	require.NoError(t, err)

	ctx := t.Context()
	params, err := cl.BTCStakingRewardsPortion(ctx)
	require.NoError(t, err)

	provisions, err := cl.AnnualProvisions(ctx)
	require.NoError(t, err)

	result, err := params.Mul(provisions).QuoInt64(1e6).Float64()
	require.NoError(t, err)

	fmt.Printf("Baby annual rewards: %f\n", result)
}
