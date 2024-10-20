package v2service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/service"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type V2ServiceInterface interface {
	service.ServiceInterface
	GetFinalityProviders(ctx context.Context, paginationKey string) ([]FinalityProviderPublic, string, *types.Error)
	GetGlobalParams(ctx context.Context) (GlobalParamsPublic, *types.Error)
	GetOverallStats(ctx context.Context) (OverallStatsPublic, *types.Error)
	GetStakerDelegations(ctx context.Context, paginationKey string) ([]StakerDelegationPublic, string, *types.Error)
	GetStakerStats(ctx context.Context, stakerPKHex string) (StakerStatsPublic, *types.Error)
}
