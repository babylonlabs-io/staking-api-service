package v1handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
)

type DelegationCheckPublicResponse struct {
	Data bool `json:"data"`
	Code int  `json:"code"`
}

// GetStakerDelegations @Summary Get phase-1 staker delegations
// @Description Retrieves phase-1 delegations for a given staker. This endpoint will be deprecated once all phase-1 delegations are either withdrawn or registered into phase-2.
// @Description This endpoint is only used to show legacy phase-1 delegations for the purpose of unbonding or registering into phase-2.
// @Produce json
// @Tags v1
// @Param staker_btc_pk query string true "Staker BTC Public Key"
// @Param pending_action query boolean false "Only return delegations with pending actions which include active, unbonding, unbonding_requested, unbonded"
// @Param pagination_key query string false "Pagination key to fetch the next page of delegations"
// @Success 200 {object} handler.PublicResponse[[]v1service.DelegationPublic]{array} "List of delegations and pagination token"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/staker/delegations [get]
func (h *V1Handler) GetStakerDelegations(request *http.Request) (*handler.Result, *types.Error) {
	stakerBtcPk, err := handler.ParsePublicKeyQuery(request, "staker_btc_pk", false)
	if err != nil {
		return nil, err
	}
	paginationKey, err := handler.ParsePaginationQuery(request)
	if err != nil {
		return nil, err
	}
	pendingAction, err := handler.ParseBooleanQuery(request, "pending_action", true)
	if err != nil {
		return nil, err
	}
	stateFilter := []types.DelegationState{}
	if pendingAction {
		// We only fetch for states that can have pending actions.
		// We don't care terminal states such as "withdrawn" or "transitioned".
		stateFilter = append(
			stateFilter,
			types.Active, types.UnbondingRequested, types.Unbonded, types.Unbonding,
		)
	}

	delegations, newPaginationKey, err := h.Service.DelegationsByStakerPk(
		request.Context(), stakerBtcPk, stateFilter, paginationKey,
	)
	if err != nil {
		return nil, err
	}

	return handler.NewResultWithPagination(delegations, newPaginationKey), nil
}

// CheckStakerDelegationExist @Summary Check if a staker has an active delegation
// @Description Check if a staker has an active delegation by the staker BTC address (Taproot or Native Segwit).
// @Description Optionally, you can provide a timeframe to check if the delegation is active within the provided timeframe
// @Description The available timeframe is "today" which checks after UTC 12AM of the current day
// @Produce json
// @Tags shared
// @Param address query string true "Staker BTC address in Taproot/Native Segwit format"
// @Param timeframe query string false "Check if the delegation is active within the provided timeframe" Enums(today)
// @Success 200 {object} DelegationCheckPublicResponse "Delegation check result"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/staker/delegation/check [get]
func (h *V1Handler) CheckStakerDelegationExist(request *http.Request) (*handler.Result, *types.Error) {
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

func buildDelegationCheckResponse(exist bool) *handler.Result {
	return &handler.Result{
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
