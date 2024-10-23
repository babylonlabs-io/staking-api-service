package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
)

func (db *Database) SaveTermsAcceptance(ctx context.Context, termsAcceptance *model.TermsAcceptance) error {
	collection := db.Client.Database(db.DbName).Collection(model.TermsAcceptanceCollection)

	filter := bson.M{
		"address":    termsAcceptance.Address,
		"public_key": termsAcceptance.PublicKey,
	}

	update := bson.M{
		"$setOnInsert": termsAcceptance,
	}

	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}
