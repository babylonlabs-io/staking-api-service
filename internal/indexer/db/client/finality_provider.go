package indexerdbclient

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (indexerdbclient *IndexerDatabase) FindFinalityProviders(
	ctx context.Context, paginationToken string,
) (*db.DbResultMap[indexerdbmodel.IndexerFinalityProviderDetails], error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).Collection(indexerdbmodel.FinalityProviderDetailsCollection)

	filter := bson.M{}

	// Default sort by commission, then by btc_pk for stable sorting
	options := options.Find().SetSort(bson.D{
		{Key: "commission", Value: 1}, // Ascending by commission
		{Key: "_id", Value: 1},        // Then by btc_pk for stable sorting
	})

	// Decode the pagination token if it exists
	if paginationToken != "" {
		decodedToken, err := dbmodel.DecodePaginationToken[indexerdbmodel.IndexerFinalityProviderPagination](paginationToken)
		if err != nil {
			return nil, &db.InvalidPaginationTokenError{
				Message: "Invalid pagination token",
			}
		}

		filter = bson.M{
			"$or": []bson.M{
				{
					"commission": decodedToken.Commission,
					"_id":        bson.M{"$gt": decodedToken.BtcPk},
				},
				{
					"commission": bson.M{"$gt": decodedToken.Commission},
				},
			},
		}
	}

	return db.FindWithPagination(
		ctx, client, filter, options, indexerdbclient.Cfg.MaxPaginationLimit,
		indexerdbmodel.BuildFinalityProviderPaginationToken,
	)
}
