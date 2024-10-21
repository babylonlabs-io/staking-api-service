package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetStats @Summary Get overall system stats
// @Description Overall system stats
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[v2service.OverallStatsPublic] ""
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v2/stats [get]
func (h *V2Handler) GetStats(request *http.Request) (*handler.Result, *types.Error) {
	stats, err := h.Service.GetOverallStats(request.Context())
	if err != nil {
		return nil, err
	}
	return handler.NewResult(stats), nil
}
