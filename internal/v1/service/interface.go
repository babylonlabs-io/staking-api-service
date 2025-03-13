package v1service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type V1ServiceProvider interface {
	service.SharedServiceProvider
	// Delegation
	DelegationsByStakerPk(ctx context.Context, stakerPk string, states []types.DelegationState, pageToken string) ([]*DelegationPublic, string, *types.Error)
	SaveActiveStakingDelegation(ctx context.Context, txHashHex, stakerPkHex, finalityProviderPkHex string, value, startHeight uint64, stakingTimestamp int64, timeLock, stakingOutputIndex uint64, stakingTxHex string, isOverflow bool) *types.Error
	IsDelegationPresent(ctx context.Context, txHashHex string) (bool, *types.Error)
	GetDelegation(ctx context.Context, txHashHex string) (*DelegationPublic, *types.Error)
	CheckStakerHasActiveDelegationByPk(ctx context.Context, stakerPkHex string, afterTimestamp int64) (bool, *types.Error)
	TransitionToUnbondingState(ctx context.Context, txHashHex string, startHeight, timelock, outputIndex uint64, txHex string, startTimestamp int64) *types.Error
	TransitionToWithdrawnState(ctx context.Context, txHashHex string) *types.Error
	UnbondDelegation(ctx context.Context, stakingTxHashHex, unbondingTxHashHex, unbondingTxHex, signatureHex string) *types.Error
	IsEligibleForUnbondingRequest(ctx context.Context, stakingTxHashHex string) *types.Error
	// Finality Provider
	GetFinalityProvidersFromGlobalParams() []*FpParamsPublic
	GetFinalityProvider(ctx context.Context, finalityProviderPkHex string) (*FpDetailsPublic, *types.Error)
	GetFinalityProviders(ctx context.Context, pageToken string) ([]*FpDetailsPublic, string, *types.Error)
	FindRegisteredFinalityProvidersNotInUse(ctx context.Context, fpParams []*FpParamsPublic) ([]*FpDetailsPublic, error)
	// Global Params
	GetGlobalParamsPublic() *GlobalParamsPublic
	GetVersionedGlobalParamsByHeight(height uint64) *types.VersionedGlobalParams
	// Staker
	ProcessAndSaveBtcAddresses(ctx context.Context, stakerPkHex string) *types.Error
	GetStakerPublicKeysByAddresses(ctx context.Context, addresses []string) (map[string]string, *types.Error)
	// Stats
	ProcessStakingStatsCalculation(ctx context.Context, stakingTxHashHex, stakerPkHex, fpPkHex string, state types.DelegationState, amount uint64) *types.Error
	GetOverallStats(ctx context.Context) (*OverallStatsPublic, *types.Error)
	GetStakerStats(ctx context.Context, stakerPkHex string) (*StakerStatsPublic, *types.Error)
	GetTopStakersByActiveTvl(ctx context.Context, pageToken string) ([]StakerStatsPublic, string, *types.Error)
	ProcessBtcInfoStats(ctx context.Context, btcHeight uint64, confirmedTvl uint64, unconfirmedTvl uint64) *types.Error
	// Timelock
	ProcessExpireCheck(ctx context.Context, stakingTxHashHex string, startHeight, timelock uint64, txType types.StakingTxType) *types.Error
	TransitionToUnbondedState(ctx context.Context, stakingType types.StakingTxType, stakingTxHashHex string) *types.Error
}
