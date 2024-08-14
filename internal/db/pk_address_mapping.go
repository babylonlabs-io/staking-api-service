package db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"go.mongodb.org/mongo-driver/mongo"
)

func (db *Database) InsertPkAddressMappings(
	ctx context.Context, pkHex, taproot, nativeSigwitOdd, nativeSigwitEven string,
) error {
	client := db.Client.Database(db.DbName).Collection(model.PkAddressMappingsCollection)
	addressMapping := &model.PkAddressMapping{
		PkHex:            pkHex,
		Taproot:          taproot,
		NativeSegwitOdd:  nativeSigwitOdd,
		NativeSegwitEven: nativeSigwitEven,
	}
	_, err := client.InsertOne(ctx, addressMapping)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return err
	}
	return nil
}
