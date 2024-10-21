package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetFinalityProviders gets finality providers
// @Summary Get Finality Providers
// @Description Fetches finality providers including their public keys, active tvl, total tvl, descriptions, commission, active delegations and total delegations etc
// @Produce json
// @Param pagination_key query string false "Pagination key to fetch the next page of finality providers"
// @Param finality_provider_pk query string false "Filter by finality provider public key"
// @Param sort_by query string false "Sort by field" Enums(active_tvl, name, commission)
// @Param order query string false "Order" Enums(asc, desc)
// @Success 200 {object} PublicResponse[[]v2service.FinalityProviderPublic]{array} "List of finality providers and pagination token"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v2/finality-providers [get]
func (h *V2Handler) GetFinalityProviders(request *http.Request) (*handler.Result, *types.Error) {
	paginationKey, err := handler.ParsePaginationQuery(request)
	if err != nil {
		return nil, err
	}
	providers, paginationToken, err := h.Service.GetFinalityProviders(request.Context(), paginationKey)
	if err != nil {
		return nil, err
	}
	return handler.NewResultWithPagination(providers, paginationToken), nil
}
