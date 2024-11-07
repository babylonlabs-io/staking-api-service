package v2service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type V2ServiceProvider interface {
	service.SharedServiceProvider
	GetFinalityProviders(ctx context.Context, state types.FinalityProviderQueryingState, paginationKey string) ([]*FinalityProviderPublic, string, *types.Error)
	SearchFinalityProviders(ctx context.Context, searchQuery string, paginationKey string) ([]*FinalityProviderPublic, string, *types.Error)
	GetGlobalParams(ctx context.Context) (*GlobalParamsPublic, *types.Error)
	GetOverallStats(ctx context.Context) (OverallStatsPublic, *types.Error)
	GetStakerDelegations(ctx context.Context, paginationKey string) ([]StakerDelegationPublic, string, *types.Error)
	GetStakerStats(ctx context.Context, stakerPKHex string) (StakerStatsPublic, *types.Error)
}
