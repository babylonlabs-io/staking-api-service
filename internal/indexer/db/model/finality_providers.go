package indexerdbmodel

import (
	"encoding/json"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
)

type IndexerFinalityProviderDetails struct {
	BtcPk          string      `bson:"_id"` // Primary key
	BabylonAddress string      `bson:"babylon_address"`
	Commission     string      `bson:"commission"`
	State          string      `bson:"state"`
	Description    Description `bson:"description"`
}

// Description represents the nested description field
type Description struct {
	Moniker         string `bson:"moniker"`
	Identity        string `bson:"identity"`
	Website         string `bson:"website"`
	SecurityContact string `bson:"security_contact"`
	Details         string `bson:"details"`
}

type IndexerFinalityProviderPagination struct {
	BtcPk      string `json:"btc_pk"`
	Commission string `json:"commission"`
}

func BuildFinalityProviderPaginationToken(f IndexerFinalityProviderDetails) (string, error) {
	page := &IndexerFinalityProviderPagination{
		BtcPk:      f.BtcPk,
		Commission: f.Commission,
	}
	token, err := dbmodel.GetPaginationToken(page)
	if err != nil {
		return "", err
	}

	return token, nil
}

func DecodeFinalityProviderPaginationToken(token string) (*IndexerFinalityProviderPagination, error) {
	var pagination IndexerFinalityProviderPagination
	err := json.Unmarshal([]byte(token), &pagination)
	return &pagination, err
}
