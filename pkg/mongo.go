package pkg

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func FetchAll[T any](ctx context.Context, collection *mongo.Collection, filter bson.M) ([]T, error) {
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []T
	for cursor.Next(ctx) {
		var doc T

		err = cursor.Decode(&doc)
		if err != nil {
			return nil, err
		}

		result = append(result, doc)
	}

	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	return result, nil
}
