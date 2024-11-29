package v1handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1service "github.com/babylonlabs-io/staking-api-service/internal/v1/service"
)

// Deprecated: GetFinalityProviders gets active finality providers sorted by ActiveTvl. Querying V2 finality providers is preferred for a combined list of V1 and V2 finality providers.
// @Summary Get Active Finality Providers
// @Description Fetches details of all active finality providers sorted by their active total value locked (ActiveTvl) in descending order.
// @Produce json
// @Tags v1
// @Param fp_btc_pk query string false "Public key of the finality provider to fetch"
// @Param pagination_key query string false "Pagination key to fetch the next page of finality providers"
// @Success 200 {object} handler.PublicResponse[[]v1service.FpDetailsPublic] "A list of finality providers sorted by ActiveTvl in descending order"
// @Router /v1/finality-providers [get]
func (h *V1Handler) GetFinalityProviders(request *http.Request) (*handler.Result, *types.Error) {
	fpPk, err := handler.ParsePublicKeyQuery(request, "fp_btc_pk", true)
	if err != nil {
		return nil, err
	}
	if fpPk != "" {
		var result []*v1service.FpDetailsPublic
		fp, err := h.Service.GetFinalityProvider(request.Context(), fpPk)
		if err != nil {
			return nil, err
		}
		if fp != nil {
			result = append(result, fp)
		}

		return handler.NewResult(result), nil
	}

	paginationKey, err := handler.ParsePaginationQuery(request)
	if err != nil {
		return nil, err
	}
	fps, paginationToken, err := h.Service.GetFinalityProviders(request.Context(), paginationKey)
	if err != nil {
		return nil, err
	}
	return handler.NewResultWithPagination(fps, paginationToken), nil
}
