package v1dbclient

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
)

//go:generate mockery --name=V1DBClient --output=../../../../tests/mocks --outpkg=mocks --filename=mock_v1_db_client.go
type V1DBClient interface {
	dbclient.DBClient
	SaveActiveStakingDelegation(
		ctx context.Context, stakingTxHashHex, stakerPkHex, fpPkHex string,
		stakingTxHex string, amount, startHeight, timelock, outputIndex uint64,
		startTimestamp int64, isOverflow bool,
	) error
	// FindDelegationsByStakerPk finds all delegations by the staker's public key.
	// The extraFilter parameter can be used to filter the results by the delegation's
	// properties. The paginationToken parameter is used to fetch the next page of results.
	// If the paginationToken is empty, the first page of results will be fetched.
	// The returned DbResultMap will contain the next pagination token if there are more
	// results to fetch.
	FindDelegationsByStakerPk(
		ctx context.Context, stakerPk string,
		extraFilter *DelegationFilter, paginationToken string,
	) (*db.DbResultMap[v1dbmodel.DelegationDocument], error)
	SaveUnbondingTx(
		ctx context.Context, stakingTxHashHex, unbondingTxHashHex, txHex, signatureHex string,
	) error
	FindDelegationByTxHashHex(ctx context.Context, txHashHex string) (*v1dbmodel.DelegationDocument, error)
	TransitionToTransitionedState(ctx context.Context, stakingTxHashHex string) error
	SaveTimeLockExpireCheck(ctx context.Context, stakingTxHashHex string, expireHeight uint64, txType string) error
	TransitionToUnbondedState(
		ctx context.Context, stakingTxHashHex string, eligiblePreviousState []types.DelegationState,
	) error
	TransitionToUnbondingState(
		ctx context.Context, txHashHex string, startHeight, timelock, outputIndex uint64, txHex string, startTimestamp int64,
	) error
	TransitionToWithdrawnState(ctx context.Context, txHashHex string) error
	GetOrCreateStatsLock(
		ctx context.Context, stakingTxHashHex string, state string,
	) (*v1dbmodel.StatsLockDocument, error)
	SubtractOverallStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
	) error
	IncrementOverallStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
	) error
	GetOverallStats(ctx context.Context) (*v1dbmodel.OverallStatsDocument, error)
	IncrementFinalityProviderStats(
		ctx context.Context, stakingTxHashHex, fpPkHex string, amount uint64,
	) error
	SubtractFinalityProviderStats(
		ctx context.Context, stakingTxHashHex, fpPkHex string, amount uint64,
	) error
	FindFinalityProviderStats(ctx context.Context, paginationToken string) (*db.DbResultMap[*v1dbmodel.FinalityProviderStatsDocument], error)
	FindFinalityProviderStatsByFinalityProviderPkHex(
		ctx context.Context, finalityProviderPkHex []string,
	) ([]*v1dbmodel.FinalityProviderStatsDocument, error)
	IncrementStakerStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
	) error
	SubtractStakerStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
	) error
	FindTopStakersByTvl(ctx context.Context, paginationToken string) (*db.DbResultMap[*v1dbmodel.StakerStatsDocument], error)
	// GetStakerStats fetches the staker stats by the staker's public key.
	GetStakerStats(
		ctx context.Context, stakerPkHex string,
	) (*v1dbmodel.StakerStatsDocument, error)
	UpsertLatestBtcInfo(
		ctx context.Context, height uint64, confirmedTvl uint64, unconfirmedTvl uint64,
	) error
	GetLatestBtcInfo(ctx context.Context) (*v1dbmodel.BtcInfo, error)
	CheckDelegationExistByStakerPk(
		ctx context.Context, address string, extraFilter *DelegationFilter,
	) (bool, error)
	// ScanDelegationsPaginated scans the delegation collection in a paginated way
	// without applying any filters or sorting, ensuring that all existing items
	// are eventually fetched.
	ScanDelegationsPaginated(
		ctx context.Context,
		paginationToken string,
	) (*db.DbResultMap[v1dbmodel.DelegationDocument], error)
}

type DelegationFilter struct {
	AfterTimestamp int64
	States         []types.DelegationState
}
