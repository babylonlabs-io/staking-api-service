package handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"github.com/babylonlabs-io/staking-api-service/internal/utils"
)

type DelegationCheckPublicResponse struct {
	Data bool `json:"data"`
	Code int  `json:"code"`
}

// GetStakerDelegations @Summary Get staker delegations
// @Description Retrieves delegations for a given staker
// @Produce json
// @Param staker_btc_pk query string true "Staker BTC Public Key"
// @Param state query types.DelegationState false "Filter by state"
// @Param pagination_key query string false "Pagination key to fetch the next page of delegations"
// @Success 200 {object} PublicResponse[[]services.DelegationPublic]{array} "List of delegations and pagination token"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/staker/delegations [get]
func (h *Handler) GetStakerDelegations(request *http.Request) (*Result, *types.Error) {
	stakerBtcPk, err := parsePublicKeyQuery(request, "staker_btc_pk", false)
	if err != nil {
		return nil, err
	}
	paginationKey, err := parsePaginationQuery(request)
	if err != nil {
		return nil, err
	}
	stateFilter, err := parseStateFilterQuery(request, "state")
	if err != nil {
		return nil, err
	}
	delegations, newPaginationKey, err := h.services.DelegationsByStakerPk(
		request.Context(), stakerBtcPk, stateFilter, paginationKey,
	)
	if err != nil {
		return nil, err
	}

	return NewResultWithPagination(delegations, newPaginationKey), nil
}

// CheckStakerDelegationExist @Summary Check if a staker has an active delegation
// @Description Check if a staker has an active delegation by the staker BTC address (Taproot or Native Segwit)
// @Description Optionally, you can provide a timeframe to check if the delegation is active within the provided timeframe
// @Description The available timeframe is "today" which checks after UTC 12AM of the current day
// @Produce json
// @Param address query string true "Staker BTC address in Taproot/Native Segwit format"
// @Param timeframe query string false "Check if the delegation is active within the provided timeframe" Enums(today)
// @Success 200 {object} DelegationCheckPublicResponse "Delegation check result"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/staker/delegation/check [get]
func (h *Handler) CheckStakerDelegationExist(request *http.Request) (*Result, *types.Error) {
	address, err := parseBtcAddressQuery(request, "address", h.config.Server.BTCNetParam)
	if err != nil {
		return nil, err
	}

	afterTimestamp, err := parseTimeframeToAfterTimestamp(request.URL.Query().Get("timeframe"))
	if err != nil {
		return nil, err
	}

	addressToPkMapping, err := h.services.GetStakerPublicKeysByAddresses(request.Context(), []string{address})
	if err != nil {
		return nil, err
	}
	if _, exist := addressToPkMapping[address]; !exist {
		return buildDelegationCheckResponse(false), nil
	}

	exist, err := h.services.CheckStakerHasActiveDelegationByPk(
		request.Context(), addressToPkMapping[address], afterTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return buildDelegationCheckResponse(exist), nil
}

func buildDelegationCheckResponse(exist bool) *Result {
	return &Result{
		Data: &DelegationCheckPublicResponse{
			Data: exist, Code: 0,
		},
		Status: http.StatusOK,
	}
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
