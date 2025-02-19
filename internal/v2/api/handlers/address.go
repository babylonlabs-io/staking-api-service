package v2handlers

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"net/http"
)

// AddressScreening gets staker stats for babylon staking
// todo change
// @Summary Get Staker Stats
// @Description Fetches staker stats for babylon staking including active tvl and active delegations.
// @Produce json
// @Tags v2
// @Param address query string true "Address to check"
// @Success 200 {object} handler.PublicResponse[v2service.StakerStatsPublic] "Staker stats"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Failure 404 {object} types.Error "Error: Not Found"
// @Failure 500 {object} types.Error "Error: Internal Server Error"
// @Router /v2/address/screening [get]
func (h *V2Handler) AddressScreening(request *http.Request) (*handler.Result, *types.Error) {
	address := request.URL.Query().Get("address")
	if address == "" {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "address is required")
	}

	result, err := h.Service.AssessAddress(address)
	if err != nil {
		return nil, types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "error assessing address")
	}

	return handler.NewResult(result), nil
}
