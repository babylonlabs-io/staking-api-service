package indexerdbclient

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *IndexerDatabase) GetEventConsumers(ctx context.Context) ([]indexerdbmodel.EventConsumer, error) {
	cursor, err := db.collection(indexerdbmodel.EventConsumerCollection).
		Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var consumers []indexerdbmodel.EventConsumer
	for cursor.Next(ctx) {
		var consumer indexerdbmodel.EventConsumer
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
