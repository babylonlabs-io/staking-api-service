package v1dbmodel

type TimeLockDocument struct {
	StakingTxHashHex string `bson:"staking_tx_hash_hex"`
	ExpireHeight     uint64 `bson:"expire_height"`
	TxType           string `bson:"tx_type"`
}

func NewTimeLockDocument(stakingTxHashHex string, expireHeight uint64, txType string) *TimeLockDocument {
	return &TimeLockDocument{
		StakingTxHashHex: stakingTxHashHex,
		ExpireHeight:     expireHeight,
		TxType:           txType,
	}
}
