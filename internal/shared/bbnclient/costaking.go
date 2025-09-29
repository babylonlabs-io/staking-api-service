package bbnclient

import (
	"context"
	"fmt"

	cosmosMath "cosmossdk.io/math"

	costakingTypes "github.com/babylonlabs-io/babylon/v4/x/costaking/types"
	"github.com/cosmos/cosmos-sdk/client"
)

func (c *BBNClient) CostakingTotalScoreSum(ctx context.Context) (cosmosMath.Int, error) {
	callForResponse := func() (*costakingTypes.QueryCurrentRewardsResponse, error) {
		queryClient := costakingTypes.NewQueryClient(client.Context{Client: c.queryClient.RPCClient})
		response, err := queryClient.CurrentRewards(ctx, &costakingTypes.QueryCurrentRewardsRequest{})
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	response, err := clientCallWithRetry(ctx, callForResponse, c.cfg)
	if err != nil {
		return cosmosMath.Int{}, fmt.Errorf("failed to get costaking total score: %w", err)
	}
	return response.TotalScore, nil
}

func (c *BBNClient) CostakingParams(ctx context.Context) (costakingTypes.Params, error) {
	callForResponse := func() (*costakingTypes.QueryParamsResponse, error) {
		queryClient := costakingTypes.NewQueryClient(client.Context{Client: c.queryClient.RPCClient})
		response, err := queryClient.Params(ctx, &costakingTypes.QueryParamsRequest{})
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	response, err := clientCallWithRetry(ctx, callForResponse, c.cfg)
	if err != nil {
		return costakingTypes.Params{}, fmt.Errorf("failed to get costaking total score: %w", err)
	}
	return response.Params, nil
}
