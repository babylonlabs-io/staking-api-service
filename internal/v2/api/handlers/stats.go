package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetStats @Summary Get overall system stats
// @Description Overall system stats
// @Produce json
// @Success 200 {object} PublicResponse[v2service.OverallStatsPublic] ""
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v2/stats [get]
func (h *V2Handler) GetStats(request *http.Request) (*handler.Result, *types.Error) {
	// TODO: Implement the logic to get overall stats
	// mock data response
	return handler.NewResult(map[string]string{"message": "V2 Get Stats"}), nil
}