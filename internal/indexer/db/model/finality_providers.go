package indexerdbmodel

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
