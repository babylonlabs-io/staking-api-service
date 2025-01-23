package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
)

// GetStakerStats gets staker stats for babylon staking
// @Summary Get Staker Stats
// @Description Fetches staker stats for babylon staking including active tvl and active delegations.
// @Produce json
// @Tags v2
// @Param staker_pk_hex query string true "Public key of the staker to fetch"
// @Success 200 {object} handler.PublicResponse[v2service.StakerStatsPublic] "Staker stats"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Failure 404 {object} types.Error "Error: Not Found"
// @Failure 500 {object} types.Error "Error: Internal Server Error"
// @Router /v2/staker/stats [get]
func (h *V2Handler) GetStakerStats(request *http.Request) (*handler.Result, *types.Error) {
	stakerPKHex := request.URL.Query().Get("staker_pk_hex")
	if stakerPKHex == "" {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "staker_pk_hex is required")
	}
	stats, err := h.Service.GetStakerStats(request.Context(), stakerPKHex)
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

// CheckStakerDelegationExist @Summary Check if a staker has an active delegation
// @Description Check if a staker has an active delegation by the staker BTC address (Taproot or Native Segwit)
// @Description Optionally, you can provide a timeframe to check if the delegation is active within the provided timeframe
// @Description The available timeframe is "today" which checks after UTC 12AM of the current day
// @Produce json
// @Tags v1
// @Param address query string true "Staker BTC address in Taproot/Native Segwit format"
// @Param timeframe query string false "Check if the delegation is active within the provided timeframe" Enums(today)
// @Success 200 {object} DelegationCheckPublicResponse "Delegation check result"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/staker/delegation/check [get]
func (h *V2Handler) CheckStakerDelegationExist(request *http.Request) (*handler.Result, *types.Error) {
	address, err := handler.ParseBtcAddressQuery(request, "address", h.Handler.Config.Server.BTCNetParam)
	if err != nil {
		return nil, err
	}

	afterTimestamp, err := parseTimeframeToAfterTimestamp(request.URL.Query().Get("timeframe"))
	if err != nil {
		return nil, err
	}

	addressToPkMapping, err := h.Service.GetStakerPublicKeysByAddresses(request.Context(), []string{address})
	if err != nil {
		return nil, err
	}
	if _, exist := addressToPkMapping[address]; !exist {
		return buildDelegationCheckResponse(false), nil
	}

	exist, err := h.Service.CheckStakerHasActiveDelegationByPk(
		request.Context(), addressToPkMapping[address], afterTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return buildDelegationCheckResponse(exist), nil
}

func parseTimeframeToAfterTimestamp(timeframe string) (int64, *types.Error) {
	switch timeframe {
	case "": // We ignore and return 0 if no timeframe is provided
		return 0, nil
	case "today":
		return utils.GetTodayStartTimestampInSeconds(), nil
	default:
		return 0, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "invalid timeframe value",
		)
	}
}

type DelegationCheckPublicResponse struct {
	Data bool `json:"data"`
	Code int  `json:"code"`
}

func buildDelegationCheckResponse(exist bool) *handler.Result {
	return &handler.Result{
		Data: &DelegationCheckPublicResponse{
			Data: exist, Code: 0,
		},
		Status: http.StatusOK,
	}
}
