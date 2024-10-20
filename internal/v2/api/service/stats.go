package v2service

type OverallStatsPublic struct {
	ActiveTVL               int64 `json:"active_tvl"`
	TotalTVL                int64 `json:"total_tvl"`
	ActiveDelegations       int64 `json:"active_delegations"`
	TotalDelegations        int64 `json:"total_delegations"`
	ActiveStakers           int64 `json:"active_stakers"`
	TotalStakers            int64 `json:"total_stakers"`
	ActiveFinalityProviders int64 `json:"active_finality_providers"`
	TotalFinalityProviders  int64 `json:"total_finality_providers"`
}


