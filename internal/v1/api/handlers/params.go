package v1handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetBabylonGlobalParams godoc
// @Summary Get Babylon global parameters
// @Description Retrieves the global parameters for Babylon, including finality provider details.
// @Produce json
// @Success 200 {object} PublicResponse[services.GlobalParamsPublic] "Global parameters"
// @Router /v1/global-params [get]
func (h *V1Handler) GetBabylonGlobalParams(request *http.Request) (*handler.Result, *types.Error) {
	params := h.Service.GetGlobalParamsPublic()
	return handler.NewResult(params), nil
}
