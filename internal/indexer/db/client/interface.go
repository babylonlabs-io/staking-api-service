package indexerdbclient

import (
	"context"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
)

type IndexerDBClient interface {
	Ping(ctx context.Context) error
	// Params
	GetBbnStakingParams(ctx context.Context) ([]*indexertypes.BbnStakingParams, error)
	GetBtcCheckpointParams(ctx context.Context) ([]*indexertypes.BtcCheckpointParams, error)
}
