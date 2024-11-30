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
	MarkV1DelegationAsTransitioned(ctx context.Context, stakingTxHashHex string) *types.Error
	GetOverallStats(ctx context.Context) (*OverallStatsPublic, *types.Error)
	GetStakerStats(ctx context.Context, stakerPKHex string) (*StakerStatsPublic, *types.Error)
	ProcessAndSaveBtcAddresses(ctx context.Context, stakerPkHex string) *types.Error
	SaveUnprocessableMessages(ctx context.Context, messageBody, receipt string) *types.Error
	ProcessActiveDelegationStats(ctx context.Context, stakingTxHashHex, stakerPkHex string, fpBtcPkHexes []string, amount uint64) *types.Error
	ProcessUnbondingDelegationStats(ctx context.Context, stakingTxHashHex, stakerPkHex string, fpBtcPkHexes []string, amount uint64) *types.Error
}
