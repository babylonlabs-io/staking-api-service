package v2dbclient

import (
	"context"

	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
)

//go:generate mockery --name=V2DBClient --output=../../../../tests/mocks --outpkg=mocks --filename=mock_v2_db_client.go
type V2DBClient interface {
	dbclient.DBClient
	GetOverallStats(ctx context.Context) (*v2dbmodel.V2OverallStatsDocument, error)
	GetStakerStats(ctx context.Context, stakerPKHex string) (*v2dbmodel.V2StakerStatsDocument, error)
	GetFinalityProviderStats(ctx context.Context) ([]*v2dbmodel.V2FinalityProviderStatsDocument, error)
	GetOrCreateStatsLock(
		ctx context.Context, stakingTxHashHex string, state string,
	) (*v2dbmodel.V2StatsLockDocument, error)
	IncrementOverallStats(
		ctx context.Context, stakingTxHashHex string, amount uint64,
	) error
	SubtractOverallStats(
		ctx context.Context, stakingTxHashHex string, amount uint64,
	) error
	HandleActiveStakerStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
	) error
	HandleUnbondingStakerStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64, stateHistory []string,
	) error
	HandleWithdrawableStakerStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64, stateHistory []string,
	) error
	HandleWithdrawnStakerStats(
		ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64, stateHistory []string,
	) error
	IncrementFinalityProviderStats(
		ctx context.Context, stakingTxHashHex string, fpPkHexes []string, amount uint64,
	) error
	SubtractFinalityProviderStats(
		ctx context.Context, stakingTxHashHex string, fpPkHexes []string, amount uint64,
	) error
	GetActiveStakersCount(ctx context.Context) (int64, error)
}
