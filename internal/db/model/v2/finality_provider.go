package v2model

type FinalityProviderDocument struct {
	ID                string `bson:"_id"`
	ActiveTVL         int64  `bson:"active_tvl"`
	TotalTVL          int64  `bson:"total_tvl"`
	ActiveDelegations int64  `bson:"active_delegations"`
	TotalDelegations  int64  `bson:"total_delegations"`
	Commission        string `bson:"commission"`
}

