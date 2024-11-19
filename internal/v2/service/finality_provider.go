package v2service

import (
	"context"
	"net/http"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

type FinalityProviderPublic struct {
	BtcPk             string                              `json:"btc_pk"`
	State             types.FinalityProviderQueryingState `json:"state"`
	Description       types.FinalityProviderDescription   `json:"description"`
	Commission        string                              `json:"commission"`
	ActiveTvl         int64                               `json:"active_tvl"`
	TotalTvl          int64                               `json:"total_tvl"`
	ActiveDelegations int64                               `json:"active_delegations"`
	TotalDelegations  int64                               `json:"total_delegations"`
}

type FinalityProvidersPublic struct {
	FinalityProviders []FinalityProviderPublic `json:"finality_providers"`
}

func mapToFinalityProviderPublic(provider indexerdbmodel.IndexerFinalityProviderDetails) *FinalityProviderPublic {
	return &FinalityProviderPublic{
		BtcPk:       provider.BtcPk,
		State:       types.FinalityProviderQueryingState(provider.State),
		Description: types.FinalityProviderDescription(provider.Description),
		Commission:  provider.Commission,
		// TODO: add active_tvl, total_tvl, active_delegations, total_delegations from statistic data field
		ActiveTvl:         0,
		TotalTvl:          0,
		ActiveDelegations: 0,
		TotalDelegations:  0,
	}
}

// GetFinalityProviders gets a list of finality providers with optional filters
func (s *V2Service) GetFinalityProviders(ctx context.Context, state types.FinalityProviderQueryingState, paginationKey string) ([]*FinalityProviderPublic, string, *types.Error) {
	resultMap, err := s.DbClients.IndexerDBClient.GetFinalityProviders(ctx, state, paginationKey)
	if err != nil {
		if db.IsNotFoundError(err) {
			log.Ctx(ctx).Warn().Err(err).Msg("Finality providers not found")
			return nil, "", types.NewErrorWithMsg(http.StatusNotFound, types.NotFound, "finality providers not found, please retry")
		}
		return nil, "", types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "failed to get finality providers")
	}

	providersPublic := make([]*FinalityProviderPublic, 0, len(resultMap.Data))
	for _, provider := range resultMap.Data {
		providersPublic = append(providersPublic, mapToFinalityProviderPublic(provider))
	}
	return providersPublic, resultMap.PaginationToken, nil
}

// SearchFinalityProviders searches for finality providers with optional filters
func (s *V2Service) SearchFinalityProviders(ctx context.Context, searchQuery string, paginationKey string) ([]*FinalityProviderPublic, string, *types.Error) {
	resultMap, err := s.DbClients.IndexerDBClient.SearchFinalityProviders(ctx, searchQuery, paginationKey)
	if err != nil {
		if db.IsNotFoundError(err) {
			log.Ctx(ctx).Warn().Err(err).Str("searchQuery", searchQuery).Msg("Finality providers not found")
			return nil, "", types.NewErrorWithMsg(http.StatusNotFound, types.NotFound, "finality providers not found, please retry")
		}
		return nil, "", types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "failed to search finality providers")
	}

	providersPublic := make([]*FinalityProviderPublic, 0, len(resultMap.Data))
	for _, provider := range resultMap.Data {
		providersPublic = append(providersPublic, mapToFinalityProviderPublic(provider))
	}
	return providersPublic, resultMap.PaginationToken, nil
}
