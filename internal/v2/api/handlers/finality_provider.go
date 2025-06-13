package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/pkg"
)

// GetFinalityProviders gets a list of finality providers with its stats
// @Summary List Finality Providers
// @Description Fetches finality providers with its stats, currently does not support pagination
// the response contains a field for pagination token, but it's not used yet
// this is for the future when we will support pagination
// @Param consumer_id query string false "Filter by consumer id"
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[[]v2service.FinalityProviderPublic] "List of finality providers with its stats"
// @Failure 404 {object} types.Error "No finality providers found"
// @Failure 500 {object} types.Error "Internal server error occurred"
// @Router /v2/finality-providers [get]
func (h *V2Handler) GetFinalityProviders(request *http.Request) (*handler.Result, *types.Error) {
	consumerID := request.URL.Query().Get("consumer_id")
	// todo add validation for consumer_id

	providers, err := h.Service.GetFinalityProvidersWithStats(request.Context(), pkg.PtrIfNonZero(consumerID))
	if err != nil {
		return nil, err
	}

	return handler.NewResultWithPagination(providers, ""), nil
}
