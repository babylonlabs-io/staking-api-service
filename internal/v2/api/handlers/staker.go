package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetStakerStats gets staker stats for babylon staking
// @Summary Get Staker Stats
// @Description Fetches staker stats for babylon staking including active tvl, total tvl, active delegations and total delegations.
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[v2service.StakerStatsPublic] "Staker stats"
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
