package handlers

import (
	"net/http"

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

// GetTopStakerStats gets top stakers by active tvl
// @Summary Get Top Staker Stats by Active TVL
// @Description Fetches details of top stakers by their active total value locked (ActiveTvl) in descending order.
// @Produce json
// @Param  pagination_key query string false "Pagination key to fetch the next page of top stakers"
// @Success 200 {object} PublicResponse[[]services.StakerStatsPublic]{array} "List of top stakers by active tvl"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/stats/staker [get]
func (h *Handler) GetTopStakerStats(request *http.Request) (*Result, *types.Error) {
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
