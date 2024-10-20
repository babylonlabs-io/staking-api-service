package v2handlers

import (
	"go/types"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
)

func (h *V2Handler) GetStats(request *http.Request) (*handler.Result, *types.Error) {
	// TODO: Implement the logic to get overall stats
	// mock data response
	return handler.NewResult(map[string]string{"message": "V2 Get Stats"}), nil
}

func (h *V2Handler) GetStakerStats(request *http.Request) (*handler.Result, *types.Error) {
	// TODO: Implement the logic to get staker stats
	// mock data response
	return handler.NewResult(map[string]string{"message": "V2 Get Staker Stats"}), nil
}
