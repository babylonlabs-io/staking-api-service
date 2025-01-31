package dbclient

import (
	"context"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (db *Database) InsertPkAddressMappings(
	ctx context.Context, pkHex, taproot, nativeSigwitOdd, nativeSigwitEven string,
) error {
	client := db.Client.Database(db.DbName).Collection(dbmodel.PkAddressMappingsCollection)
	addressMapping := &dbmodel.PkAddressMapping{
		PkHex:            pkHex,
		Taproot:          taproot,
		NativeSegwitOdd:  nativeSigwitOdd,
		NativeSegwitEven: nativeSigwitEven,
	}
	_, err := client.InsertOne(ctx, addressMapping)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		metrics.RecordDbError("insert_pk_address_mappings")
		return err
	}
	return nil
}

func (db *Database) FindPkMappingsByTaprootAddress(
	ctx context.Context, taprootAddresses []string,
) ([]*dbmodel.PkAddressMapping, error) {
	client := db.Client.Database(db.DbName).Collection(dbmodel.PkAddressMappingsCollection)
	filter := bson.M{"taproot": bson.M{"$in": taprootAddresses}}

	addressMapping := []*dbmodel.PkAddressMapping{}
	cursor, err := client.Find(ctx, filter)
	if err != nil {
		metrics.RecordDbError("find_pk_mappings_by_taproot_address")
		return nil, err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &addressMapping); err != nil {
		metrics.RecordDbError("find_pk_mappings_by_taproot_address")
		return nil, err
	}
	return addressMapping, nil
}

func (db *Database) FindPkMappingsByNativeSegwitAddress(
	ctx context.Context, nativeSegwitAddresses []string,
) ([]*dbmodel.PkAddressMapping, error) {
	client := db.Client.Database(db.DbName).Collection(dbmodel.PkAddressMappingsCollection)
	filter := bson.M{
		"$or": []bson.M{
			{"native_segwit_even": bson.M{"$in": nativeSegwitAddresses}},
			{"native_segwit_odd": bson.M{"$in": nativeSegwitAddresses}},
		},
	}

	addressMapping := []*dbmodel.PkAddressMapping{}
	cursor, err := client.Find(ctx, filter)
	if err != nil {
		metrics.RecordDbError("find_pk_mappings_by_native_segwit_address")
		return nil, err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &addressMapping); err != nil {
		metrics.RecordDbError("find_pk_mappings_by_native_segwit_address")
		return nil, err
	}
	return addressMapping, nil
}
