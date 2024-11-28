package v2service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type V2ServiceProvider interface {
	GetFinalityProvidersWithStats(ctx context.Context) (
		[]*FinalityProviderStatsPublic, *types.Error,
	)
	GetNetworkInfo(ctx context.Context) (*NetworkInfoPublic, *types.Error)
	IsDelegationPresent(ctx context.Context, txHashHex string) (bool, *types.Error)
	GetDelegation(ctx context.Context, stakingTxHashHex string) (*StakerDelegationPublic, *types.Error)
	GetDelegations(ctx context.Context, stakerPKHex string, paginationKey string) ([]*StakerDelegationPublic, string, *types.Error)
	GetOverallStats(ctx context.Context) (*OverallStatsPublic, *types.Error)

	// Staker
	ProcessAndSaveBtcAddresses(ctx context.Context, stakerPkHex string) *types.Error
	ProcessStakingStatsCalculation(
		ctx context.Context,
		stakingTxHashHex, stakerPkHex string,
		finalityProviderBtcPksHex []string,
		state types.DelegationState, amount uint64,
	) *types.Error
}
