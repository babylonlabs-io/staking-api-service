package v1model

import (
	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

type TimelockTransaction struct {
	TxHex          string `bson:"tx_hex"`
	OutputIndex    uint64 `bson:"output_index"`
	StartTimestamp int64  `bson:"start_timestamp"`
	StartHeight    uint64 `bson:"start_height"`
	TimeLock       uint64 `bson:"timelock"`
}

type DelegationDocument struct {
	StakingTxHashHex      string                `bson:"_id"` // Primary key
	StakerPkHex           string                `bson:"staker_pk_hex"`
	FinalityProviderPkHex string                `bson:"finality_provider_pk_hex"`
	StakingValue          uint64                `bson:"staking_value"`
	State                 types.DelegationState `bson:"state"`
	StakingTx             *TimelockTransaction  `bson:"staking_tx"` // Always exist
	UnbondingTx           *TimelockTransaction  `bson:"unbonding_tx,omitempty"`
	IsOverflow            bool                  `bson:"is_overflow"`
}

type DelegationByStakerPagination struct {
	StakingTxHashHex   string `json:"staking_tx_hash_hex"`
	StakingStartHeight uint64 `json:"staking_start_height"`
}

func BuildDelegationByStakerPaginationToken(d DelegationDocument) (string, error) {
	page := &DelegationByStakerPagination{
		StakingTxHashHex:   d.StakingTxHashHex,
		StakingStartHeight: d.StakingTx.StartHeight,
	}
	token, err := model.GetPaginationToken(page)
	if err != nil {
		return "", err
	}
	return token, nil
}

type DelegationScanPagination struct {
	StakingTxHashHex string `json:"staking_tx_hash_hex"`
}

func BuildDelegationScanPaginationToken(d DelegationDocument) (string, error) {
	page := &DelegationScanPagination{
		StakingTxHashHex: d.StakingTxHashHex,
	}
	token, err := model.GetPaginationToken(page)
	if err != nil {
		return "", err
	}
	return token, nil
}
