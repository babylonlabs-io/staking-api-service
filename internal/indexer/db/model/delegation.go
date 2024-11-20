package indexerdbmodel

import (
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
)

type IndexerDelegationPagination struct {
	StakingTxHashHex string `json:"staking_tx_hash_hex"`
	// TODO: The start height shall be the BBN height https://github.com/babylonlabs-io/babylon-staking-indexer/issues/47
	StartHeight uint32 `json:"start_height"`
}

type IndexerDelegationDetails struct {
	StakingTxHashHex          string                       `bson:"_id"` // Primary key
	StakingTxHex              string                       `bson:"staking_tx_hex"`
	ParamsVersion             uint32                       `bson:"params_version"`
	FinalityProviderBtcPksHex []string                     `bson:"finality_provider_btc_pks_hex"`
	StakerBtcPkHex            string                       `bson:"staker_btc_pk_hex"`
	StakingTime               uint32                       `bson:"staking_time"`
	StakingAmount             uint64                       `bson:"staking_amount"`
	StakingOutputPkScript     string                       `bson:"staking_output_pk_script"`
	StakingOutputIdx          uint32                       `bson:"staking_output_idx"`
	UnbondingTime             uint32                       `bson:"unbonding_time"`
	UnbondingTx               string                       `bson:"unbonding_tx"`
	State                     indexertypes.DelegationState `bson:"state"`
	StartHeight               uint32                       `bson:"start_height"`
	EndHeight                 uint32                       `bson:"end_height"`
}

func BuildDelegationPaginationToken(d IndexerDelegationDetails) (string, error) {
	page := &IndexerDelegationPagination{
		StakingTxHashHex: d.StakingTxHashHex,
		StartHeight:      d.StartHeight,
	}
	token, err := dbmodel.GetPaginationToken(page)
	if err != nil {
		return "", err
	}

	return token, nil
}
