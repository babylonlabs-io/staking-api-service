package indexerdbclient

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (indexerdbclient *IndexerDatabase) GetStakerDelegations(
	ctx context.Context, stakerPKHex string, paginationToken string,
) (*db.DbResultMap[indexerdbmodel.IndexerStakerDelegationDetails], error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).Collection(indexerdbmodel.BTCDelegationDetailsCollection)

	// Base filter with stakingTxHashHex
	filter := bson.M{"staker_btc_pk_hex": stakerPKHex}

	// Default sort by start_height for stable sorting
	options := options.Find().SetSort(bson.D{
		{Key: "start_height", Value: 1},
	})

	// Decode the pagination token if it exists
	if paginationToken != "" {
		decodedToken, err := dbmodel.DecodePaginationToken[indexerdbmodel.IndexerStakerDelegationPagination](paginationToken)
		if err != nil {
			return nil, &db.InvalidPaginationTokenError{
				Message: "Invalid pagination token",
			}
		}

		// Add start_height filter while maintaining the stakingTxHashHex filter
		filter = bson.M{
			"staker_btc_pk_hex": stakerPKHex,
			"start_height":      bson.M{"$gt": decodedToken.StartHeight},
		}
	}

	return db.FindWithPagination(
		ctx, client, filter, options, indexerdbclient.Cfg.MaxPaginationLimit,
		indexerdbmodel.BuildStakerDelegationPaginationToken,
	)
}
