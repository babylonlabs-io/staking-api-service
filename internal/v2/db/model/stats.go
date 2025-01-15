package v2dbmodel

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
	Id                string `bson:"_id"`
	ActiveTvl         int64  `bson:"active_tvl"`
	ActiveDelegations int64  `bson:"active_delegations"`
}

type V2StakerStatsDocument struct {
	StakerPkHex             string `bson:"_id"`
	ActiveTvl               int64  `bson:"active_tvl"`
	ActiveDelegations       int64  `bson:"active_delegations"`
	UnbondingTvl            int64  `bson:"unbonding_tvl"`
	UnbondingDelegations    int64  `bson:"unbonding_delegations"`
	WithdrawableTvl         int64  `bson:"withdrawable_tvl"`
	WithdrawableDelegations int64  `bson:"withdrawable_delegations"`
}

type V2FinalityProviderStatsDocument struct {
	FinalityProviderPkHex string `bson:"_id"`
	ActiveTvl             int64  `bson:"active_tvl"`
	ActiveDelegations     int64  `bson:"active_delegations"`
}
