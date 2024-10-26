package handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

type TermsAcceptanceLoggingRequest struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
}

type TermsAcceptancePublic struct {
	Status bool `json:"status"`
}

func (h *Handler) LogTermsAcceptance(request *http.Request) (*Result, *types.Error) {
	address, publicKey, err := parseTermsAcceptanceLoggingRequest(request, h.config.Server.BTCNetParam)
	if err != nil {
		return nil, err
	}

	if err := h.services.AcceptTerms(request.Context(), address, publicKey); err != nil {
		return nil, err
	}

	return NewResult(TermsAcceptancePublic{Status: true}), nil
}
