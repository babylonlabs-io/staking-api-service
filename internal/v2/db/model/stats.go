package v2dbmodel

import dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"

// StatsLockDocument represents the document in the stats lock collection
// It's used as a lock to prevent concurrent stats calculation for the same staking tx hash
// As well as to prevent the same staking tx hash + txType to be processed multiple times
// The already processed stats will be marked as true in the document
type V2StatsLockDocument struct {
	Id                    string `bson:"_id"`
	OverallStats          bool   `bson:"overall_stats"`
	StakerStats           bool   `bson:"staker_stats"`
	FinalityProviderStats bool   `bson:"finality_provider_stats"`
}

func NewV2StatsLockDocument(
	id string, overallStats, stakerStats, finalityProviderStats bool,
) *V2StatsLockDocument {
	return &V2StatsLockDocument{
		Id:                    id,
		OverallStats:          overallStats,
		StakerStats:           stakerStats,
		FinalityProviderStats: finalityProviderStats,
	}
}

type V2OverallStatsDocument struct {
	Id                      string `bson:"_id"`
	ActiveTvl               int64  `bson:"active_tvl"`
	TotalTvl                int64  `bson:"total_tvl"`
	ActiveDelegations       int64  `bson:"active_delegations"`
	TotalDelegations        int64  `bson:"total_delegations"`
	ActiveStakers           uint64 `bson:"active_stakers"`
	TotalStakers            uint64 `bson:"total_stakers"`
	ActiveFinalityProviders uint64 `bson:"active_finality_providers"`
	TotalFinalityProviders  uint64 `bson:"total_finality_providers"`
}

type V2FinalityProviderStatsDocument struct {
	FinalityProviderPkHex string `bson:"_id"` // FinalityProviderPkHex
	ActiveTvl             int64  `bson:"active_tvl"`
	TotalTvl              int64  `bson:"total_tvl"`
	ActiveDelegations     int64  `bson:"active_delegations"`
	TotalDelegations      int64  `bson:"total_delegations"`
}

type V2FinalityProviderStatsPagination struct {
	FinalityProviderPkHex string `json:"finality_provider_pk_hex"`
	ActiveTvl             int64  `json:"active_tvl"`
}

func BuildV2FinalityProviderStatsPaginationToken(d *V2FinalityProviderStatsDocument) (string, error) {
	page := V2FinalityProviderStatsPagination{
		ActiveTvl:             d.ActiveTvl,
		FinalityProviderPkHex: d.FinalityProviderPkHex,
	}
	token, err := dbmodel.GetPaginationToken(page)
	if err != nil {
		return "", err
	}
	return token, nil
}

type StakingTx struct {
	TxHashHex string `bson:"tx_hash_hex"`
}

type V2StakerStatsDocument struct {
	StakerPkHex                  string      `bson:"_id"` // StakerPkHex
	StakingTxs                   []StakingTx `bson:"staking_txs"`
	ActiveTvl                    int64       `bson:"active_tvl"`
	WithdrawableTvl              int64       `bson:"withdrawable_tvl"`
	SlashedTvl                   int64       `bson:"slashed_tvl"`
	TotalActiveDelegations       int64       `bson:"total_active_delegations"`
	TotalWithdrawableDelegations int64       `bson:"total_withdrawable_delegations"`
	TotalSlashedDelegations      int64       `bson:"total_slashed_delegations"`
}

// StakerStatsByStakerPagination is used to paginate the top stakers by active tvl
// ActiveTvl is used as the sorting key, whereas StakerPkHex is used as the secondary sorting key
type V2StakerStatsByStakerPagination struct {
	StakerPkHex string `json:"staker_pk_hex"`
	ActiveTvl   int64  `json:"active_tvl"`
}

func BuildV2StakerStatsByStakerPaginationToken(d *V2StakerStatsDocument) (string, error) {
	page := V2StakerStatsByStakerPagination{
		StakerPkHex: d.StakerPkHex,
		ActiveTvl:   d.ActiveTvl,
	}
	token, err := dbmodel.GetPaginationToken(page)
	if err != nil {
		return "", err
	}
	return token, nil
}
