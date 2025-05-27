package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type AddressScreeningResponse struct {
	BTCAddress struct {
		Risk string `json:"risk"`
	} `json:"btc_address"`
}

func (h *V2Handler) AddressScreening(request *http.Request) (*handler.Result, *types.Error) {
	btcAddress := request.URL.Query().Get("btc_address")
	if btcAddress == "" {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "btc_address is required")
	}

	ctx := request.Context()
	result, err := h.Service.AssessAddress(ctx, btcAddress)
	if err != nil {
		return nil, types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "error assessing address")
	}

	var data AddressScreeningResponse
	data.BTCAddress.Risk = result.Risk
	return handler.NewResult(data), nil
}
