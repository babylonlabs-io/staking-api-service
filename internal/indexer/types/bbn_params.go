package indexertypes

type GlobalParamsType string

const (
	CHECKPOINT_PARAMS_VERSION uint32           = 0
	CHECKPOINT_PARAMS_TYPE    GlobalParamsType = "CHECKPOINT"
	STAKING_PARAMS_TYPE       GlobalParamsType = "STAKING"
)

type BbnStakingParams struct {
	Version                      uint32   `json:"version"`
	CovenantPks                  []string `json:"covenant_pks"`
	CovenantQuorum               uint32   `json:"covenant_quorum"`
	MinStakingValueSat           int64    `json:"min_staking_value_sat"`
	MaxStakingValueSat           int64    `json:"max_staking_value_sat"`
	MinStakingTimeBlocks         uint32   `json:"min_staking_time_blocks"`
	MaxStakingTimeBlocks         uint32   `json:"max_staking_time_blocks"`
	SlashingPkScript             string   `json:"slashing_pk_script"`
	MinSlashingTxFeeSat          int64    `json:"min_slashing_tx_fee_sat"`
	SlashingRate                 string   `json:"slashing_rate"`
	UnbondingTimeBlocks          uint32   `json:"unbonding_time_blocks"`
	UnbondingFeeSat              int64    `json:"unbonding_fee_sat"`
	MinCommissionRate            string   `json:"min_commission_rate"`
	MaxActiveFinalityProviders   uint32   `json:"max_active_finality_providers"`
	DelegationCreationBaseGasFee uint64   `json:"delegation_creation_base_gas_fee"`
	AllowListExpirationHeight    uint64   `json:"allow_list_expiration_height"`
	BtcActivationHeight          uint32   `json:"btc_activation_height"`
}

type BtcCheckpointParams struct {
	Version              uint32 `json:"version"`
	BtcConfirmationDepth uint64 `json:"btc_confirmation_depth"`
}
