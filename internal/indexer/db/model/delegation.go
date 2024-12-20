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

type CovenantSignature struct {
	CovenantBtcPkHex string `bson:"covenant_btc_pk_hex"`
	SignatureHex     string `bson:"signature_hex"`
}

type BTCDelegationCreatedBbnBlock struct {
	Height    int64 `bson:"height"`
	Timestamp int64 `bson:"timestamp"` // epoch time in seconds
}

type SlashingTx struct {
	SlashingTxHex          string `bson:"slashing_tx_hex"`
	UnbondingSlashingTxHex string `bson:"unbonding_slashing_tx_hex"`
	SpendingHeight         uint32 `bson:"spending_height"`
}

type IndexerDelegationDetails struct {
	StakingTxHashHex             string                          `bson:"_id"` // Primary key
	StakingTxHex                 string                          `bson:"staking_tx_hex"`
	ParamsVersion                uint32                          `bson:"params_version"`
	FinalityProviderBtcPksHex    []string                        `bson:"finality_provider_btc_pks_hex"`
	StakerBtcPkHex               string                          `bson:"staker_btc_pk_hex"`
	StakingTimeLock              uint32                          `bson:"staking_time"`
	StakingAmount                uint64                          `bson:"staking_amount"`
	StakingOutputPkScript        string                          `bson:"staking_output_pk_script"`
	StakingOutputIdx             uint32                          `bson:"staking_output_idx"`
	UnbondingTimeLock            uint32                          `bson:"unbonding_time"`
	UnbondingTx                  string                          `bson:"unbonding_tx"`
	State                        indexertypes.DelegationState    `bson:"state"`
	SubState                     indexertypes.DelegationSubState `bson:"sub_state,omitempty"`
	StartHeight                  uint32                          `bson:"start_height"`
	EndHeight                    uint32                          `bson:"end_height"`
	CovenantUnbondingSignatures  []CovenantSignature             `bson:"covenant_unbonding_signatures"`
	BTCDelegationCreatedBbnBlock BTCDelegationCreatedBbnBlock    `bson:"btc_delegation_created_bbn_block"`
	SlashingTx                   SlashingTx                      `bson:"slashing_tx"`
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
