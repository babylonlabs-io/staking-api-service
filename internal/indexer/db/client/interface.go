package indexerdbclient

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
)

//go:generate mockery --name=IndexerDBClient --output=../../../../tests/mocks --outpkg=mocks --filename=mock_indexer_db_client.go
type IndexerDBClient interface {
	Ping(ctx context.Context) error
	// Params
	GetBbnStakingParams(ctx context.Context) ([]*indexertypes.BbnStakingParams, error)
	GetBtcCheckpointParams(ctx context.Context) ([]*indexertypes.BtcCheckpointParams, error)
	// Finality Providers
	GetFinalityProviders(ctx context.Context, bsnID *string) ([]*indexerdbmodel.IndexerFinalityProviderDetails, error)
	// Staker Delegations
	GetDelegation(ctx context.Context, stakingTxHashHex string) (*indexerdbmodel.IndexerDelegationDetails, error)
	GetDelegations(ctx context.Context, stakerPKHex string, stakerBabylonAddress *string, paginationToken string) (*db.DbResultMap[indexerdbmodel.IndexerDelegationDetails], error)
	// GetLastProcessedBbnHeight retrieves the last processed BBN height.
	GetLastProcessedBbnHeight(ctx context.Context) (lastProcessedHeight uint64, err error)
	CheckDelegationExistByStakerPk(
		ctx context.Context, address string, extraFilter *DelegationFilter,
	) (bool, error)
	GetDelegationsInStates(
		ctx context.Context,
		stakerPKHex string,
		stakerBabylonAddress *string,
		states []indexertypes.DelegationState,
	) ([]indexerdbmodel.IndexerDelegationDetails, error)
	GetAllBSN(ctx context.Context) ([]indexerdbmodel.BSN, error)
}

type DelegationFilter struct {
	AfterTimestamp int64
	States         []indexertypes.DelegationState
}
