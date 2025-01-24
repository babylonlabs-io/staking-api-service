package v1handlers

import (
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
)

type UnbondDelegationRequestPayload struct {
	StakingTxHashHex         string `json:"staking_tx_hash_hex"`
	UnbondingTxHashHex       string `json:"unbonding_tx_hash_hex"`
	UnbondingTxHex           string `json:"unbonding_tx_hex"`
	StakerSignedSignatureHex string `json:"staker_signed_signature_hex"`
}

func parseUnbondDelegationRequestPayload(request *http.Request) (*UnbondDelegationRequestPayload, *types.Error) {
	payload := &UnbondDelegationRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(payload)
	if err != nil {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "invalid request payload")
	}
	// Validate the payload fields
	if !utils.IsValidTxHash(payload.StakingTxHashHex) {
		return nil, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "invalid staking transaction hash",
		)
	}
	if !utils.IsValidTxHash(payload.UnbondingTxHashHex) {
		return nil, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "invalid unbonding transaction hash",
		)
	}
	if !utils.IsValidTxHex(payload.UnbondingTxHex) {
		return nil, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "invalid unbonding transaction hex",
		)
	}
	if !utils.IsValidSignatureFormat(payload.StakerSignedSignatureHex) {
		return nil, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "invalid staker signed signature hex",
		)
	}

	return payload, nil
}

// UnbondDelegation godoc
// @Summary Unbond phase-1 delegation
// @Description Unbonds a phase-1 delegation by processing the provided transaction details. This endpoint will be deprecated once all phase-1 delegations are either withdrawn or registered into phase-2.
// @Description This is an async operation.
// @Accept json
// @Produce json
// @Tags v1
// @Param payload body UnbondDelegationRequestPayload true "Unbonding Request Payload"
// @Success 202 "Request accepted and will be processed asynchronously"
// @Failure 400 {object} types.Error "Invalid request payload"
// @Router /v1/unbonding [post]
func (h *V1Handler) UnbondDelegation(request *http.Request) (*handler.Result, *types.Error) {
	payload, err := parseUnbondDelegationRequestPayload(request)
	if err != nil {
		return nil, err
	}
	unbondErr := h.Service.UnbondDelegation(
		request.Context(), payload.StakingTxHashHex,
		payload.UnbondingTxHashHex, payload.UnbondingTxHex,
		payload.StakerSignedSignatureHex,
	)
	if unbondErr != nil {
		return nil, unbondErr
	}

	return &handler.Result{Status: http.StatusAccepted}, nil
}

// GetUnbondingEligibility godoc
// @Summary Check unbonding eligibility
// @Description Checks if a delegation identified by its staking transaction hash is eligible for unbonding. This endpoint will be deprecated once all phase-1 delegations are either withdrawn or registered into phase-2.
// @Produce json
// @Tags v1
// @Param staking_tx_hash_hex query string true "Staking Transaction Hash Hex"
// @Success 200 "The delegation is eligible for unbonding"
// @Failure 400 {object} types.Error "Missing or invalid 'staking_tx_hash_hex' query parameter"
// @Router /v1/unbonding/eligibility [get]
func (h *V1Handler) GetUnbondingEligibility(request *http.Request) (*handler.Result, *types.Error) {
	stakingTxHashHex, err := handler.ParseTxHashQuery(request, "staking_tx_hash_hex")
	if err != nil {
		return nil, err
	}
	err = h.Service.IsEligibleForUnbondingRequest(request.Context(), stakingTxHashHex)
	if err != nil {
		return nil, err
	}

	return &handler.Result{Status: http.StatusOK}, nil
}
