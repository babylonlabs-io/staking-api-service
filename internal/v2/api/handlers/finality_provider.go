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
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[[]v2service.FinalityProviderStatsPublic] "List of finality providers with its stats"
// @Failure 400 {object} types.Error "Invalid parameters or malformed request"
// @Failure 404 {object} types.Error "No finality providers found"
// @Failure 500 {object} types.Error "Internal server error occurred"
// @Router /v2/finality-providers [get]
func (h *V2Handler) GetFinalityProviders(request *http.Request) (*handler.Result, *types.Error) {
	providers, err := h.Service.GetFinalityProvidersWithStats(request.Context())

	if err != nil {
		return nil, err
	}
	return handler.NewResultWithPagination(providers, ""), nil
}
