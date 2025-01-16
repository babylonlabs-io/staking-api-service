package indexerdbclient

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
)

type IndexerDBClient interface {
	Ping(ctx context.Context) error
	// Params
	GetBbnStakingParams(ctx context.Context) ([]*indexertypes.BbnStakingParams, error)
	GetBtcCheckpointParams(ctx context.Context) ([]*indexertypes.BtcCheckpointParams, error)
	// Finality Providers
	GetFinalityProviders(ctx context.Context) ([]*indexerdbmodel.IndexerFinalityProviderDetails, error)
	GetFinalityProviderByPk(ctx context.Context, fpPk string) (*indexerdbmodel.IndexerFinalityProviderDetails, error)
	// Staker Delegations
	GetDelegation(ctx context.Context, stakingTxHashHex string) (*indexerdbmodel.IndexerDelegationDetails, error)
	GetDelegations(ctx context.Context, stakerPKHex string, paginationToken string) (*db.DbResultMap[indexerdbmodel.IndexerDelegationDetails], error)
	// GetLastProcessedBbnHeight retrieves the last processed BBN height.
	GetLastProcessedBbnHeight(ctx context.Context) (lastProcessedHeight uint64, err error)
}
