package indexerdbmodel

type BSN struct {
	ID                string         `bson:"_id"`
	Name              string         `bson:"name"`
	Description       string         `bson:"description"`
	MaxMultiStakedFPS uint32         `bson:"max_multi_staked_fps"` // max number of finality providers from consumer
	Type              string         `bson:"type"`
	RollupMetadata    *ETHL2Metadata `bson:"rollup_metadata"`
}

const (
	TypeCosmos = "COSMOS"
	TypeRollup = "ROLLUP"
)

type ETHL2Metadata struct {
	FinalityContractAddress string `bson:"finality_contract_address"`
}
