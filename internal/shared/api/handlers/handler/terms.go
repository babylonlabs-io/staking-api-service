package handler

import (
	"encoding/json"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	"github.com/btcsuite/btcd/chaincfg"
)

type TermsAcceptanceLoggingRequest struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
}

type TermsAcceptancePublic struct {
	Status bool `json:"status"`
}

func (h *Handler) LogTermsAcceptance(request *http.Request) (*Result, *types.Error) {
	address, publicKey, err := parseTermsAcceptanceLoggingRequest(request, h.Config.Server.BTCNetParam)
	if err != nil {
		return nil, err
	}

	if err := h.Service.AcceptTerms(request.Context(), address, publicKey); err != nil {
		return nil, types.NewInternalServiceError(err)
	}

	return NewResult(TermsAcceptancePublic{Status: true}), nil
}

// parseTermsAcceptanceLoggingRequest parses the terms acceptance request bdoy and returns the address and public key
func parseTermsAcceptanceLoggingRequest(request *http.Request, btcNetParam *chaincfg.Params) (string, string, *types.Error) {
	var req TermsAcceptanceLoggingRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		return "", "", types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "Invalid request payload")
	}

	// Validate the Bitcoin address
	if _, err := utils.CheckBtcAddressType(req.Address, btcNetParam); err != nil {
		return "", "", types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "Invalid Bitcoin address")
	}

	// Validate the public key
	if _, err := utils.GetSchnorrPkFromHex(req.PublicKey); err != nil {
		return "", "", types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "Invalid public key")
	}

	if req.Address == "" || req.PublicKey == "" {
		return "", "", types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "Address and public key are required")
	}

	return req.Address, req.PublicKey, nil
}
