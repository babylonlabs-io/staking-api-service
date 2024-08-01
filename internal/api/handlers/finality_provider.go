package handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

// GetFinalityProviders gets active finality providers sorted by ActiveTvl.
// @Summary Get Active Finality Providers
// @Description Fetches details of all active finality providers sorted by their active total value locked (ActiveTvl) in descending order.
// @Produce json
// @Param pagination_key query string false "Pagination key to fetch the next page of finality providers"
// @Success 200 {object} PublicResponse[[]services.FpDetailsPublic] "A list of finality providers sorted by ActiveTvl in descending order"
// @Router /v1/finality-providers [get]
func (h *Handler) GetFinalityProviders(request *http.Request) (*Result, *types.Error) {
	paginationKey, err := parsePaginationQuery(request)
	if err != nil {
		return nil, err
	}
	fps, paginationToken, err := h.services.GetFinalityProviders(request.Context(), paginationKey)
	if err != nil {
		return nil, err
	}
	return NewResultWithPagination(fps, paginationToken), nil
}
