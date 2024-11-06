package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2service "github.com/babylonlabs-io/staking-api-service/internal/v2/service"
)

// GetFinalityProviders gets a list of finality providers with optional filters
// @Summary List Finality Providers
// @Description Fetches finality providers with optional filtering and pagination
// @Produce json
// @Tags v2
// @Param pagination_key query string false "Pagination key to fetch the next page"
// @Param search query string false "Search by moniker, finality provider PK"
// @Param state query string false "Filter by state" Enums(active, standby)
// @Success 200 {object} handler.PublicResponse[[]v2service.FinalityProviderPublic]{array} "List of finality providers and pagination token"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v2/finality-providers [get]
func (h *V2Handler) GetFinalityProviders(request *http.Request) (*handler.Result, *types.Error) {
	searchQuery, err := h.ParseFPSearchQuery(request, "search", true)
	if err != nil {
		return nil, err
	}

	state, err := handler.ParseFPStateQuery(request, "state", true)
	if err != nil {
		return nil, err
	}

	paginationKey, err := handler.ParsePaginationQuery(request)
	if err != nil {
		return nil, err
	}

	var providers []*v2service.FinalityProviderPublic
	var paginationToken string

	if searchQuery != "" {
		// Search by moniker or finality provider PK
		providers, paginationToken, err = h.Service.SearchFinalityProviders(request.Context(), searchQuery, paginationKey)
	} else {
		// Get all finality providers with optional state filter
		providers, paginationToken, err = h.Service.GetFinalityProviders(request.Context(), state, paginationKey)
	}

	if err != nil {
		return nil, err
	}
	return handler.NewResultWithPagination(providers, paginationToken), nil
}
