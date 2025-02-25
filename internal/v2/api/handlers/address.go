package v2handlers

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"net/http"
)

type AddressScreeningResponse struct {
	BTCAddress struct {
		Risk string `json:"risk"`
	} `json:"btc_address"`
}

// AddressScreening checks address risk against address screening providers
// @Summary Checks address risk
// @Description Checks address risk
// @Produce json
// @Tags v2
// @Param btc_address query string true "BTC address to check"
// @Success 200 {object} handler.PublicResponse[AddressScreeningResponse] "Risk of provided address"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Failure 404 {object} types.Error "Error: Not Found"
// @Failure 500 {object} types.Error "Error: Internal Server Error"
// @Router /address/screening [get]
func (h *V2Handler) AddressScreening(request *http.Request) (*handler.Result, *types.Error) {
	btcAddress := request.URL.Query().Get("btc_address")
	if btcAddress == "" {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "btc_address is required")
	}

	result, err := h.Service.AssessAddress(nil, btcAddress)
	if err != nil {
		return nil, types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "error assessing address")
	}

	var data AddressScreeningResponse
	data.BTCAddress.Risk = result.Risk
	return handler.NewResult(data), nil
}
