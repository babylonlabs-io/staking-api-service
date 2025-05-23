package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetStakerStats gets staker stats for babylon staking
// @Summary Get Staker Stats
// @Description Fetches staker stats for babylon staking including active tvl,
// active delegations, unbonding tvl, unbonding delegations, withdrawable tvl,
// withdrawable delegations, slashed tvl and slashed delegations. If the babylon
// address is not provided, the stats will be calculated for all the delegations
// of the staker based on the staker's BTC public key.
// @Produce json
// @Tags v2
// @Param staker_pk_hex query string true "Public key of the staker to fetch"
// @Param babylon_address query string false "Babylon address of the staker to fetch"
// @Success 200 {object} handler.PublicResponse[v2service.StakerStatsPublic] "Staker stats"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Failure 404 {object} types.Error "Error: Not Found"
// @Failure 500 {object} types.Error "Error: Internal Server Error"
// @Router /v2/staker/stats [get]
func (h *V2Handler) GetStakerStats(request *http.Request) (*handler.Result, *types.Error) {
	stakerPKHex, err := handler.ParsePublicKeyQuery(request, "staker_pk_hex", false)
	if err != nil {
		return nil, err
	}

	bbnAddress, err := handler.ParseBabylonAddressQuery(
		request, "babylon_address", true,
	)
	if err != nil {
		return nil, err
	}

	stats, err := h.Service.GetStakerStats(
		request.Context(), stakerPKHex, bbnAddress,
	)
	if err != nil {
		return nil, err
	}
	return handler.NewResult(stats), nil
}

// GetStats @Summary Get overall system stats
// @Description Overall system stats
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[v2service.OverallStatsPublic] ""
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v2/stats [get]
func (h *V2Handler) GetOverallStats(request *http.Request) (*handler.Result, *types.Error) {
	stats, err := h.Service.GetOverallStats(request.Context())
	if err != nil {
		return nil, err
	}
	return handler.NewResult(stats), nil
}

// GetPrices @Summary Get latest prices for all available symbols
// @Description Get latest prices for all available symbols
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[map[string]float64] ""
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v2/prices [get]
func (h *V2Handler) GetPrices(request *http.Request) (*handler.Result, *types.Error) {
	prices, err := h.Service.GetLatestPrices(request.Context())
	if err != nil {
		return nil, err
	}

	return handler.NewResult(prices), nil
}
