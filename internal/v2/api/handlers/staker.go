package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetStakerDelegations gets staker delegations for babylon staking
// @Summary Get Staker Delegations
// @Description Fetches staker delegations for babylon staking including tvl, total delegations, active tvl, active delegations and total stakers.
// @Produce json
// @Param staking_tx_hash_hex query string true "Staking transaction hash in hex format"
// @Param pagination_key query string false "Pagination key to fetch the next page of delegations"
// @Success 200 {object} PublicResponse[[]v2service.StakerDelegationsPublic]{array} "List of staker delegations and pagination token"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v2/staker/delegations [get]
func (h *V2Handler) GetStakerDelegations(request *http.Request) (*handler.Result, *types.Error) {
	// TODO: Implement the logic to get staker delegations
	// mock data response
	return handler.NewResult(map[string]string{"message": "V2 Get Staker Delegations"}), nil
}

// GetStakerStats gets staker stats for babylon staking
// @Summary Get Staker Stats
// @Description Fetches staker stats for babylon staking including active tvl, total tvl, active delegations and total delegations.
// @Produce json
// @Success 200 {object} PublicResponse[v2service.StakerStatsPublic] "Staker stats"
// @Router /v2/staker/stats [get]
func (h *V2Handler) GetStakerStats(request *http.Request) (*handler.Result, *types.Error) {
	// TODO: Implement the logic to get staker stats
	// mock data response
	return handler.NewResult(map[string]string{"message": "V2 Get Staker Stats"}), nil
}
