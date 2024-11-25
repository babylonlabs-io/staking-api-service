package indexerdbclient

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Use namespaced key for context to avoid collisions
const ctxKey = "indexer_db.last_processed_height"

func (db *IndexerDatabase) GetLastProcessedBbnHeight(ctx context.Context) (uint64, error) {
	// Check if height is already in context
	if height, ok := ctx.Value(ctxKey).(uint64); ok {
		return height, nil
	}

	// If not in context, query from database
	var result indexerdbmodel.LastProcessedHeight
	err := db.Client.Database(db.DbName).Collection(
		indexerdbmodel.LastProcessedHeightCollection,
	).FindOne(ctx, bson.M{}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		// If no document exists, return 0
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	// Store in context for future use with namespaced key
	ctx = context.WithValue(ctx, ctxKey, result.Height)

	return result.Height, nil
}
