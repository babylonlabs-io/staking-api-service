package types

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/babylonlabs-io/networks/parameters/parser"
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
