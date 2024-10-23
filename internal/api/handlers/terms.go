package handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

type TermsAcceptanceRequest struct {
	Address       string `json:"address"`
	TermsAccepted bool   `json:"terms_accepted"`
	PublicKey     string `json:"public_key"`
}

type TermsAcceptancePublic struct {
	Status bool `json:"status"`
}

func (h *Handler) AcceptTerms(request *http.Request) (*Result, *types.Error) {
	address, publicKey, termsAccepted, err := parseTermsAcceptanceQuery(request, h.config.Server.BTCNetParam)
	if err != nil {
		return nil, err
	}

	if err := h.services.AcceptTerms(request.Context(), address, publicKey, termsAccepted); err != nil {
		return nil, err
	}

	return NewResult(TermsAcceptancePublic{Status: true}), nil
}
