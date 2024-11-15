package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetParams gets system parameters
// @Summary Get Parameters
// @Description Fetches system parameters for babylon chain and BTC chain
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[v2service.ParamsPublic] "Parameters"
// @Failure 404 {object} types.Error "Error: Not Found"
// @Failure 500 {object} types.Error "Error: Internal Server Error"
// @Router /v2/params [get]
func (h *V2Handler) GetParams(request *http.Request) (*handler.Result, *types.Error) {
	params, err := h.Service.GetParams(request.Context())
	if err != nil {
		return nil, err
	}
	return handler.NewResult(params), nil
}
