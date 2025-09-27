package indexerdbmodel

type FinalityProviderState string

const (
	FinalityProviderStatus_FINALITY_PROVIDER_STATUS_INACTIVE FinalityProviderState = "FINALITY_PROVIDER_STATUS_INACTIVE"
	FinalityProviderStatus_FINALITY_PROVIDER_STATUS_ACTIVE   FinalityProviderState = "FINALITY_PROVIDER_STATUS_ACTIVE"
	FinalityProviderStatus_FINALITY_PROVIDER_STATUS_JAILED   FinalityProviderState = "FINALITY_PROVIDER_STATUS_JAILED"
	FinalityProviderStatus_FINALITY_PROVIDER_STATUS_SLASHED  FinalityProviderState = "FINALITY_PROVIDER_STATUS_SLASHED"
)

type IndexerFinalityProviderDetails struct {
	BtcPk          string                `bson:"_id"` // Primary key
	BabylonAddress string                `bson:"babylon_address"`
	Commission     string                `bson:"commission"`
	State          FinalityProviderState `bson:"state"`
	Description    Description           `bson:"description"`
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
