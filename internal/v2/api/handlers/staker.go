package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetStakerDelegationByTxHash @Summary Get a staker delegation
// @Summary Get a staker delegation
// @Description Retrieves a staker delegation by a given transaction hash
// @Produce json
// @Tags v2
// @Param staking_tx_hash_hex query string true "Staking transaction hash in hex format"
// @Success 200 {object} handler.PublicResponse[v2service.StakerDelegationPublic] "Staker delegation"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v2/staker/delegation [get]
func (h *V2Handler) GetDelegationByTxHash(request *http.Request) (*handler.Result, *types.Error) {
	stakingTxHash, err := handler.ParseTxHashQuery(request, "staking_tx_hash_hex")
	if err != nil {
		return nil, err
	}
	delegation, err := h.Service.GetStakerDelegation(request.Context(), stakingTxHash)
	if err != nil {
		return nil, err
	}

	return handler.NewResult(delegation), nil
}


// GetStakerDelegations gets staker delegations for babylon staking
// @Summary Get Staker Delegations
// @Description Fetches staker delegations for babylon staking including tvl, total delegations, active tvl, active delegations and total stakers.
// @Produce json
// @Tags v2
// @Param staker_pk_hex query string true "Staker public key in hex format"
// @Param pagination_key query string false "Pagination key to fetch the next page of delegations"
// @Success 200 {object} handler.PublicResponse[[]v2service.StakerDelegationPublic]{array} "List of staker delegations and pagination token"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v2/staker/delegations [get]
func (h *V2Handler) GetStakerDelegations(request *http.Request) (*handler.Result, *types.Error) {
	const stakerPKHexKey string = "staker_pk_hex"
	stakerPKHex, err := handler.ParsePublicKeyQuery(request, stakerPKHexKey, false)
	if err != nil {
		return nil, err
	}
	paginationKey, err := handler.ParsePaginationQuery(request)
	if err != nil {
		return nil, err
	}
	delegations, paginationToken, err := h.Service.GetStakerDelegations(request.Context(), stakerPKHex, paginationKey)
	if err != nil {
		return nil, err
	}
	return handler.NewResultWithPagination(delegations, paginationToken), nil
}

// GetStakerStats gets staker stats for babylon staking
// @Summary Get Staker Stats
// @Description Fetches staker stats for babylon staking including active tvl, total tvl, active delegations and total delegations.
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[v2service.StakerStatsPublic] "Staker stats"
// @Router /v2/staker/stats [get]
func (h *V2Handler) GetStakerStats(request *http.Request) (*handler.Result, *types.Error) {
	stakerPKHex := request.URL.Query().Get("staker_pk_hex")
	if stakerPKHex == "" {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, "staker_pk_hex is required")
	}
	stats, err := h.Service.GetStakerStats(request.Context(), stakerPKHex)
	if err != nil {
		return nil, err
	}
	return handler.NewResult(stats), nil
}
