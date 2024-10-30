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
