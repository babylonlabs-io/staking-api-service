package indexerdbmodel

import dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"

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

// IndexerFinalityProviderStatsPagination represents pagination token for FP stats
// sorted by active_tvl in descending order with fp_btc_pk_hex as tiebreaker
type IndexerFinalityProviderStatsPagination struct {
	FpBtcPkHex string `json:"fp_btc_pk_hex"`
	ActiveTvl  uint64 `json:"active_tvl"`
}

// BuildIndexerFinalityProviderStatsPaginationToken creates pagination token from stats document
func BuildIndexerFinalityProviderStatsPaginationToken(d *IndexerFinalityProviderStatsDocument) (string, error) {
	page := IndexerFinalityProviderStatsPagination{
		FpBtcPkHex: d.FpBtcPkHex,
		ActiveTvl:  d.ActiveTvl,
	}
	token, err := dbmodel.GetPaginationToken(page)
	if err != nil {
		return "", err
	}
	return token, nil
}
