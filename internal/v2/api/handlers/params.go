package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetGlobalParams gets global parameters
// @Summary Get Global Parameters
// @Description Fetches global parameters for babylon chain and BTC chain
// @Produce json
// @Success 200 {object} PublicResponse[v2service.GlobalParamsPublic] "Global parameters"
// @Router /v2/global-params [get]
func (h *V2Handler) GetGlobalParams(request *http.Request) (*handler.Result, *types.Error) {
	// TODO: Implement the logic to get global parameters
	// mock data response
	return handler.NewResult(map[string]string{"message": "V2 Get Global Params"}), nil
}
