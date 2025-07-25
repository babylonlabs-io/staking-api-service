package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetFinalityProviders gets a list of finality providers with its stats
// @Summary List Finality Providers
// @Description Fetches finality providers with its stats, currently does not support pagination
// the response contains a field for pagination token, but it's not used yet
// this is for the future when we will support pagination
// @Param bsn_id query string false "Filter by bsn id". `all` for all FPs across all BSNs
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[[]v2service.FinalityProviderPublic] "List of finality providers with its stats"
// @Failure 404 {object} types.Error "No finality providers found"
// @Failure 500 {object} types.Error "Internal server error occurred"
// @Router /v2/finality-providers [get]
func (h *V2Handler) GetFinalityProviders(request *http.Request) (*handler.Result, *types.Error) {
	bsnID := h.getBsnIDFromQuery(request)
	providers, err := h.Service.GetFinalityProvidersWithStats(request.Context(), bsnID)
	if err != nil {
		return nil, err
	}

	return handler.NewResultWithPagination(providers, ""), nil
}

func (h *V2Handler) getBsnIDFromQuery(request *http.Request) *string {
	const paramKey = "bsn_id"
	if !request.URL.Query().Has(paramKey) {
		return nil
	}

	value := request.URL.Query().Get(paramKey)
	return &value
}
