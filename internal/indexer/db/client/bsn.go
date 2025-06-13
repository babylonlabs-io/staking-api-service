package indexerdbclient

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *IndexerDatabase) GetAllBSN(ctx context.Context) ([]indexerdbmodel.BSN, error) {
	cursor, err := db.collection(indexerdbmodel.BSNCollection).
		Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var consumers []indexerdbmodel.BSN
	for cursor.Next(ctx) {
		var consumer indexerdbmodel.BSN
		if err := cursor.Decode(&consumer); err != nil {
			return nil, err
		}

		consumers = append(consumers, consumer)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return consumers, nil
}
