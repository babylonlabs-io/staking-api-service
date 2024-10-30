package indexerdbmodel

type TimeLockDocument struct {
	StakingTxHashHex string `bson:"_id"` // Primary key
	ExpireHeight     uint32 `bson:"expire_height"`
	TxType           string `bson:"tx_type"`
}
