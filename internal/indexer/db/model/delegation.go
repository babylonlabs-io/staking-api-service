package indexerdbmodel

import (
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
)

type BTCDelegationDetails struct {
	StakingTxHashHex          string                       `bson:"_id"` // Primary key
	ParamsVersion             string                       `bson:"params_version"`
	FinalityProviderBtcPksHex []string                     `bson:"finality_provider_btc_pks_hex"`
	StakerBtcPkHex            string                       `bson:"staker_btc_pk_hex"`
	StakingTime               string                       `bson:"staking_time"`
	StakingAmount             string                       `bson:"staking_amount"`
	UnbondingTime             string                       `bson:"unbonding_time"`
	UnbondingTx               string                       `bson:"unbonding_tx"`
	State                     indexertypes.DelegationState `bson:"state"`
	StartHeight               uint32                       `bson:"start_height"`
	EndHeight                 uint32                       `bson:"end_height"`
}
