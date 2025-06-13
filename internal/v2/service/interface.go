package v2service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/integrations/chainalysis"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type V2ServiceProvider interface {
	chainalysis.Interface

	// todo rename this method ?
	GetFinalityProvidersWithStats(ctx context.Context, consumerID *string) (
		[]*FinalityProviderPublic, *types.Error,
	)
	GetNetworkInfo(ctx context.Context) (*NetworkInfoPublic, *types.Error)
	GetDelegation(ctx context.Context, stakingTxHashHex string) (*DelegationPublic, *types.Error)
	GetDelegations(ctx context.Context, stakerPKHex string, stakerBabylonAddress *string, paginationKey string) ([]*DelegationPublic, string, *types.Error)
	MarkV1DelegationAsTransitioned(
		ctx context.Context,
		stakingTxHashHex, stakerPkHex, fpPkHex string,
		stakingValue uint64,
	) *types.Error
	GetOverallStats(ctx context.Context) (*OverallStatsPublic, *types.Error)
	GetLatestPrices(ctx context.Context) (map[string]float64, *types.Error)
	/*
		Returns the staker stats for the given staker PK hex and babylon address.
		If the babylon address is not provided, the stats will be calculated for
		all the delegations.
	*/
	GetStakerStats(
		ctx context.Context,
		stakerPKHex string,
		stakerBabylonAddress *string,
	) (*StakerStatsPublic, *types.Error)
	ProcessAndSaveBtcAddresses(ctx context.Context, stakerPkHex string) *types.Error
	SaveUnprocessableMessages(ctx context.Context, messageBody, receipt string) *types.Error
	ProcessActiveDelegationStats(ctx context.Context, stakingTxHashHex, stakerPkHex string, fpBtcPkHexes []string, amount uint64) *types.Error
	ProcessUnbondingDelegationStats(ctx context.Context, stakingTxHashHex, stakerPkHex string, fpBtcPkHexes []string, amount uint64, stateHistory []string) *types.Error
	ProcessWithdrawableDelegationStats(ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64, stateHistory []string) *types.Error
	ProcessWithdrawnDelegationStats(ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64, stateHistory []string) *types.Error
	GetAllBSN(ctx context.Context) ([]BSN, error)
}
