package indexerdbmodel

import (
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
)

type IndexerGlobalParamsDocument struct {
	Type    indexertypes.GlobalParamsType `bson:"type"`
	Version uint32                        `bson:"version"`
	Params  interface{}                   `bson:"params"`
}
