package handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/services"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

// GetDelegationByTxHash @Summary Get a delegation
// @Description Retrieves a delegation by a given transaction hash
// @Produce json
// @Param staking_tx_hash_hex query string true "Staking transaction hash in hex format"
// @Success 200 {object} PublicResponse[services.DelegationPublic] "Delegation"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/delegation [get]
func (h *Handler) GetDelegationByTxHash(request *http.Request) (*Result, *types.Error) {
	stakingTxHash, err := parseTxHashQuery(request, "staking_tx_hash_hex")
	if err != nil {
		return nil, err
	}
	delegation, err := h.services.GetDelegation(request.Context(), stakingTxHash)
	if err != nil {
		return nil, err
	}

	return NewResult(services.FromDelegationDocument(delegation)), nil
}
