package indexerdbclient

import (
	"context"
	"errors"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (indexerdbclient *IndexerDatabase) GetDelegation(ctx context.Context, stakingTxHashHex string) (*indexerdbmodel.IndexerDelegationDetails, error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).Collection(indexerdbmodel.BTCDelegationDetailsCollection)
	filter := bson.M{"_id": stakingTxHashHex}
	var delegation indexerdbmodel.IndexerDelegationDetails
	err := client.FindOne(ctx, filter).Decode(&delegation)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &db.NotFoundError{
				Key:     stakingTxHashHex,
				Message: "Delegation not found",
			}
		}
		return nil, err
	}
	return &delegation, nil
}

func (indexerdbclient *IndexerDatabase) GetDelegations(
	ctx context.Context, stakerPKHex string, paginationToken string,
) (*db.DbResultMap[indexerdbmodel.IndexerDelegationDetails], error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).Collection(indexerdbmodel.BTCDelegationDetailsCollection)

	// Base filter with stakingTxHashHex
	filter := bson.M{"staker_btc_pk_hex": stakerPKHex}

	// Default sort by start_height for stable sorting
	options := options.Find().SetSort(bson.D{
		{Key: "start_height", Value: 1},
	})

	// Decode the pagination token if it exists
	if paginationToken != "" {
		decodedToken, err := dbmodel.DecodePaginationToken[indexerdbmodel.IndexerDelegationPagination](paginationToken)
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
		indexerdbmodel.BuildDelegationPaginationToken,
	)
}

func (indexerdbclient *IndexerDatabase) GetSlashedFpDelegations(ctx context.Context, fpBtcPkHex string) ([]*indexerdbmodel.IndexerDelegationDetails, error) {
	collection := indexerdbclient.Client.
		Database(indexerdbclient.DbName).
		Collection(indexerdbmodel.BTCDelegationDetailsCollection)

	filter := bson.M{
		"finality_provider_btc_pk_hex": fpBtcPkHex,
		"state":                        indexertypes.StateSlashed.String(),
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var delegations []*indexerdbmodel.IndexerDelegationDetails
	if err := cursor.All(ctx, &delegations); err != nil {
		return nil, err
	}

	return delegations, nil
}
