package indexerdbclient

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// todo use indexerdbclient.Cfg.MaxPaginationLimit once frontend is ready
const finalityProvidersLimit = 150

// CountFinalityProvidersByStatus returns counts of finality providers grouped by status
func (indexerdbclient *IndexerDatabase) CountFinalityProvidersByStatus(
	ctx context.Context,
) (map[indexerdbmodel.FinalityProviderState]uint64, error) {
	client := indexerdbclient.Client.Database(
		indexerdbclient.DbName,
	).Collection(indexerdbmodel.FinalityProviderDetailsCollection)

	pipeline := bson.A{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$state"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := client.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[indexerdbmodel.FinalityProviderState]uint64)
	for cursor.Next(ctx) {
		var doc struct {
			ID    indexerdbmodel.FinalityProviderState `bson:"_id"`
			Count uint64                               `bson:"count"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		result[doc.ID] = doc.Count
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetFinalityProviders retrieves finality providers
func (indexerdbclient *IndexerDatabase) GetFinalityProviders(
	ctx context.Context,
	paginationToken string,
) (*db.DbResultMap[*indexerdbmodel.IndexerFinalityProviderDetails], error) {
	client := indexerdbclient.Client.Database(
		indexerdbclient.DbName,
	).Collection(indexerdbmodel.FinalityProviderDetailsCollection)

	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})

	var filter bson.M

	if paginationToken != "" {
		decodedToken, err := dbmodel.DecodePaginationToken[indexerdbmodel.IndexerFinalityProviderPagination](paginationToken)
		if err != nil {
			return nil, &db.InvalidPaginationTokenError{
				Message: "Invalid pagination token",
			}
		}
		filter = bson.M{"_id": bson.M{"$gt": decodedToken.BtcPk}}
	} else {
		filter = bson.M{}
	}

	return db.FindWithPagination(
		ctx,
		client,
		filter,
		opts,
		finalityProvidersLimit,
		indexerdbmodel.BuildIndexerFinalityProviderPaginationToken,
	)
}
