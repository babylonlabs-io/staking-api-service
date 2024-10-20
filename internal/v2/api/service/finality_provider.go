package v2service

import "github.com/babylonlabs-io/staking-api-service/internal/shared/types"

type FinalityProviderPublic struct {
	BtcPK             string                            `json:"btc_pk"`
	State             string                            `json:"state"`
	Description       types.FinalityProviderDescription `json:"description"`
	Commission        string                            `json:"commission"`
	ActiveTVL         int64                             `json:"active_tvl"`
	TotalTVL          int64                             `json:"total_tvl"`
	ActiveDelegations int64                             `json:"active_delegations"`
	TotalDelegations  int64                             `json:"total_delegations"`
	// FinalityProviderDelegations int64                `json:"finality_provider_delegations,omitempty"`
}

type FinalityProvidersPublic struct {
	FinalityProviders []FinalityProviderPublic `json:"finality_providers"`
}
