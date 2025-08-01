package indexerdbclient

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"errors"
)

func (db *IndexerDatabase) GetLastProcessedBbnHeight(ctx context.Context) (uint64, error) {
	// If not in context, query from database
	var result indexerdbmodel.LastProcessedHeight
	err := db.Client.Database(db.DbName).Collection(
		indexerdbmodel.LastProcessedHeightCollection,
	).FindOne(ctx, bson.M{}).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// If no document exists, return 0
			return 0, nil
		}
		return 0, err
	}
	return result.Height, nil
}
