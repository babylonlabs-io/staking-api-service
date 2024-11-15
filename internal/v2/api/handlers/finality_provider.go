package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetFinalityProviders gets a list of finality providers with optional filters
// @Summary List Finality Providers
// @Description Fetches finality providers with optional filtering and pagination
// @Produce json
// @Tags v2
// @Param pagination_key query string false "Pagination key to fetch the next page"
// @Param state query string false "Filter by state" Enums(active, standby)
// @Success 200 {object} handler.PublicResponse[[]v2service.FinalityProviderPublic]{array} "List of finality providers and pagination token"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Failure 404 {object} types.Error "Error: Not Found"
// @Failure 500 {object} types.Error "Error: Internal Server Error"
// @Router /v2/finality-providers [get]
func (h *V2Handler) GetFinalityProviders(request *http.Request) (*handler.Result, *types.Error) {
	state, err := handler.ParseFPStateQuery(request, true)
	if err != nil {
		return nil, err
	}

	paginationKey, err := handler.ParsePaginationQuery(request)
	if err != nil {
		return nil, err
	}

	// Get all finality providers with optional state filter
	providers, paginationToken, err := h.Service.GetFinalityProviders(request.Context(), state, paginationKey)

	if err != nil {
		return nil, err
	}
	return handler.NewResultWithPagination(providers, paginationToken), nil
}
