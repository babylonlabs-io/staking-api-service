package v2service

import (
	"context"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type FinalityProviderPublic struct {
	BtcPK             string                            `json:"btc_pk"`
	State             types.FinalityProviderState       `json:"state"`
	Description       types.FinalityProviderDescription `json:"description"`
	Commission        string                            `json:"commission"`
	ActiveTVL         int64                             `json:"active_tvl"`
	TotalTVL          int64                             `json:"total_tvl"`
	ActiveDelegations int64                             `json:"active_delegations"`
	TotalDelegations  int64                             `json:"total_delegations"`
}

type FinalityProvidersPublic struct {
	FinalityProviders []FinalityProviderPublic `json:"finality_providers"`
}

// GetFinalityProviders gets a list of finality providers with optional filters
func (s *V2Service) GetFinalityProviders(ctx context.Context, state types.FinalityProviderState, paginationKey string) ([]*FinalityProviderPublic, string, *types.Error) {
	resultMap, err := s.DbClients.IndexerDBClient.GetFinalityProviders(ctx, state, paginationKey)
	if err != nil {
		return nil, "", types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "failed to get finality providers")
	}

	providersPublic := make([]*FinalityProviderPublic, 0, len(resultMap.Data))
	for _, provider := range resultMap.Data {
		providersPublic = append(providersPublic, &FinalityProviderPublic{
			BtcPK:       provider.BtcPk,
			State:       types.FinalityProviderState(provider.State),
			Description: types.FinalityProviderDescription(provider.Description),
			Commission:  provider.Commission,
			// TODO: add active_tvl, total_tvl, active_delegations, total_delegations from statistic data field
			ActiveTVL:         0,
			TotalTVL:          0,
			ActiveDelegations: 0,
			TotalDelegations:  0,
		})
	}
	return providersPublic, resultMap.PaginationToken, nil
}

// SearchFinalityProviders searches for finality providers with optional filters
func (s *V2Service) SearchFinalityProviders(ctx context.Context, searchQuery string, paginationKey string) ([]*FinalityProviderPublic, string, *types.Error) {
	resultMap, err := s.DbClients.IndexerDBClient.SearchFinalityProviders(ctx, searchQuery, paginationKey)
	if err != nil {
		return nil, "", types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "failed to search finality providers")
	}

	providersPublic := make([]*FinalityProviderPublic, 0, len(resultMap.Data))
	for _, provider := range resultMap.Data {
		providersPublic = append(providersPublic, &FinalityProviderPublic{
			BtcPK:       provider.BtcPk,
			State:       types.FinalityProviderState(provider.State),
			Description: types.FinalityProviderDescription(provider.Description),
			Commission:  provider.Commission,
			// TODO: add active_tvl, total_tvl, active_delegations, total_delegations from statistic data field
			ActiveTVL:         0,
			TotalTVL:          0,
			ActiveDelegations: 0,
			TotalDelegations:  0,
		})
	}
	return providersPublic, resultMap.PaginationToken, nil
}
