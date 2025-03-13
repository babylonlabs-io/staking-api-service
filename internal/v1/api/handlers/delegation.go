package v1handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetDelegationByTxHash @Summary Get a delegation (Deprecated)
// @Description [DEPRECATED] Retrieves a delegation by a given transaction hash. Please use /v2/delegation instead.
// @Produce json
// @Tags v1
// @Deprecated
// @Param staking_tx_hash_hex query string true "Staking transaction hash in hex format"
// @Success 200 {object} handler.PublicResponse[v1service.DelegationPublic] "Delegation"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/delegation [get]
func (h *V1Handler) GetDelegationByTxHash(request *http.Request) (*handler.Result, *types.Error) {
	stakingTxHash, err := handler.ParseTxHashQuery(request, "staking_tx_hash_hex")
	if err != nil {
		return nil, err
	}
	delegation, err := h.Service.GetDelegation(request.Context(), stakingTxHash)
	if err != nil {
		return nil, err
	}

	return handler.NewResult(delegation), nil
}
