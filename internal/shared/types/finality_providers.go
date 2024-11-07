package types

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type FinalityProviderQueryingState string

const (
	FinalityProviderStateActive  FinalityProviderQueryingState = "active"
	FinalityProviderStateStandby FinalityProviderQueryingState = "standby"
)

type FinalityProviderDescription struct {
	Moniker         string `json:"moniker"`
	Identity        string `json:"identity"`
	Website         string `json:"website"`
	SecurityContact string `json:"security_contact"`
	Details         string `json:"details"`
}

type FinalityProviderDetails struct {
	Description FinalityProviderDescription `json:"description"`
	Commission  string                      `json:"commission"`
	BtcPk       string                      `json:"btc_pk"`
}

type FinalityProviderFromFile struct {
	Description FinalityProviderDescription `json:"description"`
	Commission  string                      `json:"commission"`
	BtcPk       string                      `json:"btc_pk"`
	EotsPk      string                      `json:"eots_pk"` // eots is the default field for the pk
}

type FinalityProviders struct {
	FinalityProviders []FinalityProviderFromFile `json:"finality_providers"`
}

func NewFinalityProviders(path string) ([]FinalityProviderDetails, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	var finalityProviders FinalityProviders
	err = json.Unmarshal(data, &finalityProviders)
	if err != nil {
		return nil, err
	}

	// Convert FinalityProviderFromFile to FinalityProviderDetails
	var finalityProviderDetails []FinalityProviderDetails
	for _, fp := range finalityProviders.FinalityProviders {
		btcPk := fp.EotsPk
		if btcPk == "" {
			btcPk = fp.BtcPk
		}

		finalityProviderDetails = append(finalityProviderDetails, FinalityProviderDetails{
			Description: fp.Description,
			Commission:  fp.Commission,
			BtcPk:       btcPk,
		})
	}

	return finalityProviderDetails, nil
}
