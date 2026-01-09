package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetFinalityProviders gets a list of finality providers with their stats, sorted by active TVL
//
//	@Summary		List Finality Providers
//	@Description	Fetches finality providers with their stats, sorted by active_tvl in descending order (highest TVL first). Pagination is supported via cursor-based pagination tokens.
//	@Produce		json
//	@Tags			v2
//	@Param			pagination_key	query		string													false	"Pagination key to fetch the next page of finality providers"
//	@Success		200				{object}	handler.PublicResponse[[]v2service.FinalityProviderPublic]	"List of finality providers with stats, sorted by active_tvl DESC"
//	@Failure		400				{object}	types.Error													"Invalid pagination token"
//	@Failure		404				{object}	types.Error													"No finality providers found"
//	@Failure		500				{object}	types.Error													"Internal server error occurred"
//	@Router			/v2/finality-providers [get]
func (h *V2Handler) GetFinalityProviders(request *http.Request) (*handler.Result, *types.Error) {
	paginationToken, err := handler.ParsePaginationQuery(request)
	if err != nil {
		return nil, err
	}

	providers, nextPaginationToken, err := h.Service.GetFinalityProvidersWithStats(
		request.Context(),
		paginationToken,
	)
	if err != nil {
		return nil, err
	}

	return handler.NewResultWithPagination(providers, nextPaginationToken), nil
}
