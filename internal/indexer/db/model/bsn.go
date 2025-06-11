package indexerdbmodel

type EventConsumerType string

const (
	EventConsumerTypeCosmos EventConsumerType = "cosmos"
	EventConsumerTypeRollup EventConsumerType = "rollup"
)

type EventConsumer struct {
	ID                     string                 `bson:"_id"`
	Name                   string                 `bson:"name"`
	Description            string                 `bson:"description"`
	MaxMultiStakedFPS      uint32                 `bson:"max_multi_staked_fps"` // max number of finality providers from consumer
	Type                   EventConsumerType      `bson:"type"`
	RollupConsumerMetadata *ETHL2ConsumerMetadata `bson:"rollup_consumer_metadata"`
}

type ETHL2ConsumerMetadata struct {
	FinalityContractAddress string `bson:"finality_contract_address"`
}
