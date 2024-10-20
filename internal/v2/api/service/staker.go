package v2service

type StakerDelegationPublic struct {
	StakingTxHashHex      string          `json:"staking_tx_hash_hex"`
	StakerPKHex           string          `json:"staker_pk_hex"`
	FinalityProviderPKHex string          `json:"finality_provider_pk_hex"`
	StakingStartHeight    int64           `json:"staking_start_height"`
	UnbondingStartHeight  int64           `json:"unbonding_start_height"`
	Timelock              int64           `json:"timelock"`
	StakingValue          int64           `json:"staking_value"`
	State                 string          `json:"state"`
	StakingTx             TransactionInfo `json:"staking_tx"`
	UnbondingTx           TransactionInfo `json:"unbonding_tx"`
}

type TransactionInfo struct {
	TxHex       string `json:"tx_hex"`
	OutputIndex int    `json:"output_index"`
}

type StakerStatsPublic struct {
	StakerPKHex       string `json:"staker_pk_hex"`
	ActiveTVL         int64  `json:"active_tvl"`
	TotalTVL          int64  `json:"total_tvl"`
	ActiveDelegations int64  `json:"active_delegations"`
	TotalDelegations  int64  `json:"total_delegations"`
}