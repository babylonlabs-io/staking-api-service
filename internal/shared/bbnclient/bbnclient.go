package bbnclient

import (
	"context"
	"fmt"

	"github.com/avast/retry-go/v4"
	bbncfg "github.com/babylonlabs-io/babylon/v4/client/config"
	"github.com/babylonlabs-io/babylon/v4/client/query"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/rs/zerolog/log"
)

type BBNClient struct {
	queryClient *query.QueryClient
	cfg         *config.BBNConfig
}

func New(cfg *config.BBNConfig) (*BBNClient, error) {
	queryClient, err := query.New(&bbncfg.BabylonQueryConfig{
		RPCAddr: cfg.RPCAddr,
		Timeout: cfg.Timeout,
	})
	if err != nil {
		return nil, err
	}

	return &BBNClient{
		queryClient: queryClient,
		cfg:         cfg,
	}, nil
}

func (c *BBNClient) TotalSupply(ctx context.Context, denom string) (types.Coin, error) {
	callForResponse := func() (*banktypes.QuerySupplyOfResponse, error) {
		queryClient := banktypes.NewQueryClient(client.Context{Client: c.queryClient.RPCClient})
		response, err := queryClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{denom})
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	status, err := clientCallWithRetry(ctx, callForResponse, c.cfg)
	if err != nil {
		return types.Coin{}, fmt.Errorf("failed to get total supply: %w", err)
	}
	return status.Amount, nil
}

func (c *BBNClient) StakingPool(ctx context.Context) (stakingtypes.Pool, error) {
	callForResponse := func() (*stakingtypes.QueryPoolResponse, error) {
		queryClient := stakingtypes.NewQueryClient(client.Context{Client: c.queryClient.RPCClient})
		response, err := queryClient.Pool(ctx, &stakingtypes.QueryPoolRequest{})
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	response, err := clientCallWithRetry(ctx, callForResponse, c.cfg)
	if err != nil {
		return stakingtypes.Pool{}, fmt.Errorf("failed to get total supply: %w", err)
	}
	return response.Pool, nil
}

func clientCallWithRetry[T any](
	ctx context.Context, call retry.RetryableFuncWithData[*T], cfg *config.BBNConfig,
) (*T, error) {
	result, err := retry.DoWithData(call, retry.Context(ctx), retry.Attempts(cfg.MaxRetryTimes), retry.Delay(cfg.RetryInterval), retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			log.Ctx(ctx).Debug().
				Uint("attempt", n+1).
				Uint("max_attempts", cfg.MaxRetryTimes).
				Err(err).
				Msg("failed to call the RPC client")
		}))
	if err != nil {
		return nil, err
	}
	return result, nil
}
