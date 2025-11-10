package indexerdbmodel

// IndexerStatsDocument represents the overall stats document from the indexer
type IndexerStatsDocument struct {
	Id                string `bson:"_id"`                // Always "overall_stats"
	ActiveTvl         uint64 `bson:"active_tvl"`         // Active TVL in satoshis
	ActiveDelegations uint64 `bson:"active_delegations"` // Active delegation count
	LastUpdated       int64  `bson:"last_updated"`       // Unix timestamp
}

// IndexerFinalityProviderStatsDocument represents per-FP stats from the indexer
type IndexerFinalityProviderStatsDocument struct {
	FpBtcPkHex        string `bson:"_id"`                // FP BTC public key (lowercase)
	ActiveTvl         uint64 `bson:"active_tvl"`         // Active TVL for this FP in satoshis
	ActiveDelegations uint64 `bson:"active_delegations"` // Active delegation count for this FP
	LastUpdated       int64  `bson:"last_updated"`       // Unix timestamp
}
