package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetStakerDelegations gets staker delegations for babylon staking
// @Summary Get Staker Delegations
// @Description Fetches staker delegations for babylon staking including tvl, total delegations, active tvl, active delegations and total stakers.
// @Produce json
// @Param staking_tx_hash_hex query string true "Staking transaction hash in hex format"
// @Param pagination_key query string false "Pagination key to fetch the next page of delegations"
// @Success 200 {object} PublicResponse[[]v2service.StakerDelegationsPublic]{array} "List of staker delegations and pagination token"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v2/staker/delegations [get]
func (h *V2Handler) GetStakerDelegations(request *http.Request) (*handler.Result, *types.Error) {
	paginationKey, err := handler.ParsePaginationQuery(request)
	if err != nil {
		return nil, err
	}
	delegations, paginationToken, err := h.Service.GetStakerDelegations(request.Context(), paginationKey)
	if err != nil {
		return nil, err
	}
	return handler.NewResultWithPagination(delegations, paginationToken), nil
}

// GetStakerStats gets staker stats for babylon staking
// @Summary Get Staker Stats
// @Description Fetches staker stats for babylon staking including active tvl, total tvl, active delegations and total delegations.
// @Produce json
// @Success 200 {object} PublicResponse[v2service.StakerStatsPublic] "Staker stats"
// @Router /v2/staker/stats [get]
func (h *V2Handler) GetStakerStats(request *http.Request) (*handler.Result, *types.Error) {
	stakerPKHex := request.URL.Query().Get("staker_pk_hex")
	if stakerPKHex == "" {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "staker_pk_hex is required")
	}
	stats, err := h.Service.GetStakerStats(request.Context(), stakerPKHex)
	if err != nil {
		return nil, err
	}
	return handler.NewResult(stats), nil
}
