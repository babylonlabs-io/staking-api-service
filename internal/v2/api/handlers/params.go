package v2handlers

import (
	"go/types"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
)

func (h *V2Handler) GetGlobalParams(request *http.Request) (*handler.Result, *types.Error) {
	// TODO: Implement the logic to get global parameters
	// mock data response
	return handler.NewResult(map[string]string{"message": "V2 Get Global Params"}), nil
}
