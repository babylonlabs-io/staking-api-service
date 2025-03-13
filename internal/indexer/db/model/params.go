package indexerdbmodel

import (
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
)

type IndexerGlobalParamsDocument struct {
	Type    indexertypes.GlobalParamsType `bson:"type"`
	Version uint32                        `bson:"version"`
	Params  interface{}                   `bson:"params"`
}

type IndexerBbnStakingParamsDocument struct {
	CovenantPks                  []string `bson:"covenant_pks"`
	CovenantQuorum               uint32   `bson:"covenant_quorum"`
	MinStakingValueSat           int64    `bson:"min_staking_value_sat"`
	MaxStakingValueSat           int64    `bson:"max_staking_value_sat"`
	MinStakingTimeBlocks         uint32   `bson:"min_staking_time_blocks"`
	MaxStakingTimeBlocks         uint32   `bson:"max_staking_time_blocks"`
	SlashingPkScript             string   `bson:"slashing_pk_script"`
	MinSlashingTxFeeSat          int64    `bson:"min_slashing_tx_fee_sat"`
	SlashingRate                 string   `bson:"slashing_rate"`
	UnbondingTimeBlocks          uint32   `bson:"unbonding_time_blocks"`
	UnbondingFeeSat              int64    `bson:"unbonding_fee_sat"`
	MinCommissionRate            string   `bson:"min_commission_rate"`
	MaxActiveFinalityProviders   uint32   `bson:"max_active_finality_providers"`
	DelegationCreationBaseGasFee uint64   `bson:"delegation_creation_base_gas_fee"`
	AllowListExpirationHeight    uint64   `bson:"allow_list_expiration_height"`
	BtcActivationHeight          uint32   `bson:"btc_activation_height"`
}

type IndexerBtcCheckpointParamsDocument struct {
	Version              uint32 `bson:"version"`
	BtcConfirmationDepth uint64 `bson:"btc_confirmation_depth"`
}
