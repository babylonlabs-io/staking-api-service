package indexerdbmodel

import (
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
)

type IndexerStakerDelegationPagination struct {
	StakingTxHashHex string `json:"staking_tx_hash_hex"`
	StartHeight      uint32 `json:"start_height"`
}

type IndexerStakerDelegationDetails struct {
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

func BuildStakerDelegationPaginationToken(d IndexerStakerDelegationDetails) (string, error) {
	page := &IndexerStakerDelegationPagination{
		StakingTxHashHex: d.StakingTxHashHex,
		StartHeight:      d.StartHeight,
	}
	token, err := dbmodel.GetPaginationToken(page)
	if err != nil {
		return "", err
	}

	return token, nil
}
