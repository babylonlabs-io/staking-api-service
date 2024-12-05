package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetDelegation @Summary Get a delegation
// @Summary Get a delegation
// @Description Retrieves a delegation by a given transaction hash
// @Produce json
// @Tags v2
// @Param staking_tx_hash_hex query string true "Staking transaction hash in hex format"
// @Success 200 {object} handler.PublicResponse[v2service.DelegationPublic] "Staker delegation"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Failure 404 {object} types.Error "Error: Not Found"
// @Failure 500 {object} types.Error "Error: Internal Server Error"
// @Router /v2/delegation [get]
func (h *V2Handler) GetDelegation(request *http.Request) (*handler.Result, *types.Error) {
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

// GetDelegations gets delegations for babylon staking
// @Summary Get Delegations
// @Description Fetches delegations for babylon staking including tvl, total delegations, active tvl, active delegations and total stakers.
// @Produce json
// @Tags v2
// @Param staker_pk_hex query string true "Staker public key in hex format"
// @Param pagination_key query string false "Pagination key to fetch the next page of delegations"
// @Success 200 {object} handler.PublicResponse[[]v2service.DelegationPublic]{array} "List of staker delegations and pagination token"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Failure 404 {object} types.Error "Error: Not Found"
// @Failure 500 {object} types.Error "Error: Internal Server Error"
// @Router /v2/delegations [get]
func (h *V2Handler) GetDelegations(request *http.Request) (*handler.Result, *types.Error) {
	const stakerPKHexKey string = "staker_pk_hex"
	stakerPKHex, err := handler.ParsePublicKeyQuery(request, stakerPKHexKey, false)
	if err != nil {
		return nil, err
	}
	paginationKey, err := handler.ParsePaginationQuery(request)
	if err != nil {
		return nil, err
	}
	delegations, paginationToken, err := h.Service.GetDelegations(request.Context(), stakerPKHex, paginationKey)
	if err != nil {
		return nil, err
	}
	return handler.NewResultWithPagination(delegations, paginationToken), nil
}
