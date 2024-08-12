package types

import (
	"encoding/json"
	"os"
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

func NewFinalityProviders(filePath string) ([]FinalityProviderDetails, error) {
	data, err := os.ReadFile(filePath)
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
