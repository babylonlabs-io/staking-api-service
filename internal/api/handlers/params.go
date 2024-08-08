package handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

// GetBabylonGlobalParams godoc
// @Summary Get Babylon global parameters
// @Description Retrieves the global parameters for Babylon, including finality provider details.
// @Produce json
// @Success 200 {object} PublicResponse[services.GlobalParamsPublic] "Global parameters"
// @Router /v1/global-params [get]
func (h *Handler) GetBabylonGlobalParams(request *http.Request) (*Result, *types.Error) {
	params := h.services.GetGlobalParamsPublic()
	return NewResult(params), nil
}
