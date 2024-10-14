package v2model

type StakerDocument struct {
	ID                           string            `bson:"_id"`
	Addresses                    map[string]string `bson:"addresses"`
	ActiveTVL                    int64             `bson:"active_tvl"`
	WithdrawableTVL              int64             `bson:"withdrawable_tvl"`
	SlashedTVL                   int64             `bson:"slashed_tvl"`
	TotalActiveDelegations       int64             `bson:"total_active_delegations"`
	TotalWithdrawableDelegations int64             `bson:"total_withdrawable_delegations"`
	TotalSlashedDelegations      int64             `bson:"total_slashed_delegations"`
}

