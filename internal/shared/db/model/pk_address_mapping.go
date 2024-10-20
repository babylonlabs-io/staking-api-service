package dbmodel

type PkAddressMapping struct {
	PkHex            string `bson:"_id"`
	Taproot          string `bson:"taproot"`
	NativeSegwitEven string `bson:"native_segwit_even"`
	NativeSegwitOdd  string `bson:"native_segwit_odd"`
}
