package dbclient

import (
	"context"
	"errors"
	model "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (db *Database) SaveTxInAllowList(ctx context.Context, stakingTxHash string) error {
	doc := bson.M{"_id": stakingTxHash}

	collection := db.Client.Database(db.DbName).Collection(model.AllowListCollection)
	_, err := collection.InsertOne(ctx, doc)
	return err
}

func (db *Database) IsTxInAllowList(ctx context.Context, stakingTxHash string) (bool, error) {
	collection := db.Client.Database(db.DbName).Collection(model.AllowListCollection)
	filter := bson.M{"_id": stakingTxHash}

	err := collection.FindOne(ctx, filter).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	found := err == nil

	return found, err
}
