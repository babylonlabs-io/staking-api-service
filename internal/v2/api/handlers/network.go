package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetNetworkInfo @Summary Get network info
// @Description Get network info, including staking status and param
// @Produce json
// @Tags v2
// @Success 200 {object} v2service.NetworkInfoPublic "Network info"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v2/network-info [get]
func (h *V2Handler) GetNetworkInfo(request *http.Request) (*handler.Result, *types.Error) {
	networkInfo, err := h.Service.GetNetworkInfo(request.Context())
	if err != nil {
		return nil, err
	}
	return handler.NewResult(networkInfo), nil
}
