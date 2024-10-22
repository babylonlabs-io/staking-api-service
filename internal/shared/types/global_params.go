package types

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/babylonlabs-io/networks/parameters/parser"
	"github.com/btcsuite/btcd/btcec/v2"
)

type VersionedGlobalParams = parser.VersionedGlobalParams

type GlobalParams = parser.GlobalParams

type BabylonParams struct {
	Version                      int      `json:"version"`
	CovenantPKs                  []string `json:"covenant_pks"`
	CovenantQuorum               int      `json:"covenant_quorum"`
	MaxStakingAmount             int64    `json:"max_staking_amount"`
	MinStakingAmount             int64    `json:"min_staking_amount"`
	MaxStakingTime               int64    `json:"max_staking_time"`
	MinStakingTime               int64    `json:"min_staking_time"`
	SlashingPKScript             string   `json:"slashing_pk_script"`
	MinSlashingTxFee             int64    `json:"min_slashing_tx_fee"`
	SlashingRate                 float64  `json:"slashing_rate"`
	MinUnbondingTime             int64    `json:"min_unbonding_time"`
	UnbondingFee                 int64    `json:"unbonding_fee"`
	MinCommissionRate            float64  `json:"min_commission_rate"`
	MaxActiveFinalityProviders   int      `json:"max_active_finality_providers"`
	DelegationCreationBaseGasFee int64    `json:"delegation_creation_base_gas_fee"`
}

type BTCParams struct {
	Version              int `json:"version"`
	BTCConfirmationDepth int `json:"btc_confirmation_depth"`
}

func NewGlobalParams(path string) (*GlobalParams, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	var globalParams GlobalParams
	err = json.Unmarshal(data, &globalParams)
	if err != nil {
		return nil, err
	}

	_, err = parser.ParseGlobalParams(&globalParams)
	if err != nil {
		return nil, err
	}

	return &globalParams, nil
}

// parseCovenantPubKeyFromHex parses public key string to btc public key
// the input should be 33 bytes
func parseCovenantPubKeyFromHex(pkStr string) (*btcec.PublicKey, error) {
	pkBytes, err := hex.DecodeString(pkStr)
	if err != nil {
		return nil, err
	}

	pk, err := btcec.ParsePubKey(pkBytes)
	if err != nil {
		return nil, err
	}

	return pk, nil
}
