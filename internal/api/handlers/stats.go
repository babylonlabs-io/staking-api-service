package handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/services"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

// GetOverallStats gets overall stats for babylon staking
// @Summary Get Overall Stats
// @Description Fetches overall stats for babylon staking including tvl, total delegations, active tvl, active delegations and total stakers.
// @Produce json
// @Success 200 {object} PublicResponse[services.OverallStatsPublic] "Overall stats for babylon staking"
// @Router /v1/stats [get]
func (h *Handler) GetOverallStats(request *http.Request) (*Result, *types.Error) {
	stats, err := h.services.GetOverallStats(request.Context())
	if err != nil {
		return nil, err
	}

	return NewResult(stats), nil
}

// GetStakersStats gets staker stats for babylon staking
// @Summary Get Staker Stats
// @Description Fetches staker stats for babylon staking including tvl, total delegations, active tvl and active delegations.
// @Description If staker_btc_pk query parameter is provided, it will return stats for the specific staker.
// @Description Otherwise, it will return the top stakers ranked by active tvl.
// @Produce json
// @Param  staker_btc_pk query string false "Public key of the staker to fetch"
// @Param  pagination_key query string false "Pagination key to fetch the next page of top stakers"
// @Success 200 {object} PublicResponse[[]services.StakerStatsPublic]{array} "List of top stakers by active tvl"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/stats/staker [get]
func (h *Handler) GetStakersStats(request *http.Request) (*Result, *types.Error) {
	// Check if the request is for a specific staker
	stakerPk, err := parsePublicKeyQuery(request, "staker_btc_pk", true)
	if err != nil {
		return nil, err
	}
	if stakerPk != "" {
		var result []services.StakerStatsPublic
		stakerStats, err := h.services.GetStakerStats(request.Context(), stakerPk)
		if err != nil {
			return nil, err
		}
		if stakerStats != nil {
			result = append(result, *stakerStats)
		}

		return NewResult(result), nil
	}

	// Otherwise, fetch the top stakers ranked by active tvl
	paginationKey, err := parsePaginationQuery(request)
	if err != nil {
		return nil, err
	}
	topStakerStats, paginationToken, err := h.services.GetTopStakersByActiveTvl(request.Context(), paginationKey)
	if err != nil {
		return nil, err
	}

	return NewResultWithPagination(topStakerStats, paginationToken), nil
}
