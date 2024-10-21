package v1dbmodel

const (
	UnbondingInitialState = "INSERTED"
)

type UnbondingDocument struct {
	StakerPkHex        string `bson:"staker_pk_hex"`
	FinalityPkHex      string `bson:"finality_pk_hex"`
	UnbondingTxSigHex  string `bson:"unbonding_tx_sig_hex"`
	State              string `bson:"state"`
	UnbondingTxHashHex string `bson:"unbonding_tx_hash_hex"` // Unique Index
	UnbondingTxHex     string `bson:"unbonding_tx_hex"`
	StakingTxHex       string `bson:"staking_tx_hex"`
	StakingOutputIndex uint64 `bson:"staking_output_index"`
	StakingTimelock    uint64 `bson:"staking_timelock"`
	StakingAmount      uint64 `bson:"staking_amount"`
	StakingTxHashHex   string `json:"staking_tx_hash_hex"`
}
