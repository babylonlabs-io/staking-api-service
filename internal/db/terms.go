package db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
)

func (db *Database) SaveTermsAcceptance(ctx context.Context, termsAcceptance *model.TermsAcceptance) error {
	collection := db.Client.Database(db.DbName).Collection(model.TermsAcceptanceCollection)

	_, err := collection.InsertOne(ctx, termsAcceptance)
	return err
}
