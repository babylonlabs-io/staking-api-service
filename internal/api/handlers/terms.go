package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

type TermsAcceptanceRequest struct {
	TermsAccepted bool `json:"terms_accepted"`
}

type TermsAcceptancePublic struct {
	Status bool `json:"status"`
}

// AcceptTerms @Summary Accept terms
// @Description Track terms acceptance by the staker's BTC address (Taproot or Native Segwit)
// @Produce json
// @Param address query string true "Staker BTC address in Taproot/Native Segwit format"
// @Param terms_accepted body TermsAcceptanceRequest true "Terms acceptance request"
// @Success 200 {object} TermsAcceptancePublic "Terms acceptance result"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /terms-acceptance [post]
func (h *Handler) AcceptTerms(request *http.Request) (*Result, *types.Error) {
	address, err := parseBtcAddressQuery(request, "address", h.config.Server.BTCNetParam)
	if err != nil {
		return nil, err
	}

	publicKey, err := parsePublicKeyQuery(request, "public_key", false)
	if err != nil {
		return nil, err
	}

	var req TermsAcceptanceRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "Invalid request payload")
	}

	if err := h.services.AcceptTerms(request.Context(), address, publicKey, req.TermsAccepted); err != nil {
		return nil, err
	}

	return NewResult(TermsAcceptancePublic{Status: true}), nil
}
