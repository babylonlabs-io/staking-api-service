package db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (db *Database) InsertPkAddressMappings(
	ctx context.Context, pkHex, taproot, nativeSigwitOdd, nativeSigwitEven string,
) error {
	client := db.Client.Database(db.DbName).Collection(model.V1PkAddressMappingsCollection)
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

func (db *Database) FindPkMappingsByTaprootAddress(
	ctx context.Context, taprootAddresses []string,
) ([]*model.PkAddressMapping, error) {
	client := db.Client.Database(db.DbName).Collection(model.V1PkAddressMappingsCollection)
	filter := bson.M{"taproot": bson.M{"$in": taprootAddresses}}

	addressMapping := []*model.PkAddressMapping{}
	cursor, err := client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &addressMapping); err != nil {
		return nil, err
	}
	return addressMapping, nil
}

func (db *Database) FindPkMappingsByNativeSegwitAddress(
	ctx context.Context, nativeSegwitAddresses []string,
) ([]*model.PkAddressMapping, error) {
	client := db.Client.Database(db.DbName).Collection(
		model.V1PkAddressMappingsCollection,
	)
	filter := bson.M{
		"$or": []bson.M{
			{"native_segwit_even": bson.M{"$in": nativeSegwitAddresses}},
			{"native_segwit_odd": bson.M{"$in": nativeSegwitAddresses}},
		},
	}

	addressMapping := []*model.PkAddressMapping{}
	cursor, err := client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &addressMapping); err != nil {
		return nil, err
	}
	return addressMapping, nil
}
