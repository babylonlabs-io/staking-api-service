package v2service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type V2ServiceProvider interface {
	GetFinalityProvidersWithStats(ctx context.Context) (
		[]*FinalityProviderStatsPublic, *types.Error,
	)
	GetParams(ctx context.Context) (*ParamsPublic, *types.Error)
	GetDelegation(ctx context.Context, stakingTxHashHex string) (*StakerDelegationPublic, *types.Error)
	GetDelegations(ctx context.Context, stakerPKHex string, paginationKey string) ([]*StakerDelegationPublic, string, *types.Error)
	GetOverallStats(ctx context.Context) (*OverallStatsPublic, *types.Error)
	GetStakerStats(ctx context.Context, stakerPKHex string) (*StakerStatsPublic, *types.Error)
}
