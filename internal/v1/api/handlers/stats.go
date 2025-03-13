package v1handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1service "github.com/babylonlabs-io/staking-api-service/internal/v1/service"
)

// GetOverallStats gets overall stats for babylon staking
// @Summary Get Overall Stats (Deprecated)
// @Description [DEPRECATED] Fetches overall stats for babylon staking including tvl, total delegations, active tvl, active delegations and total stakers. Please use /v2/stats instead.
// @Produce json
// @Tags v1
// @Deprecated
// @Success 200 {object} handler.PublicResponse[v1service.OverallStatsPublic] "Overall stats for babylon staking"
// @Router /v1/stats [get]
func (h *V1Handler) GetOverallStats(request *http.Request) (*handler.Result, *types.Error) {
	stats, err := h.Service.GetOverallStats(request.Context())
	if err != nil {
		return nil, err
	}

	return handler.NewResult(stats), nil
}

// GetStakersStats gets staker stats for babylon staking
// @Summary Get Staker Stats (Deprecated)
// @Description [DEPRECATED] Fetches staker stats for babylon staking including tvl, total delegations, active tvl and active delegations. Please use /v2/staker/stats instead.
// @Description If staker_btc_pk query parameter is provided, it will return stats for the specific staker.
// @Description Otherwise, it will return the top stakers ranked by active tvl.
// @Produce json
// @Tags v1
// @Deprecated
// @Param  staker_btc_pk query string false "Public key of the staker to fetch"
// @Param  pagination_key query string false "Pagination key to fetch the next page of top stakers"
// @Success 200 {object} handler.PublicResponse[[]v1service.StakerStatsPublic]{array} "List of top stakers by active tvl"
// @Failure 400 {object} types.Error "Error: Bad Request"
// @Router /v1/stats/staker [get]
func (h *V1Handler) GetStakersStats(request *http.Request) (*handler.Result, *types.Error) {
	// Check if the request is for a specific staker
	stakerPk, err := handler.ParsePublicKeyQuery(request, "staker_btc_pk", true)
	if err != nil {
		return nil, err
	}
	if stakerPk != "" {
		var result []v1service.StakerStatsPublic
		stakerStats, err := h.Service.GetStakerStats(request.Context(), stakerPk)
		if err != nil {
			return nil, err
		}
		if stakerStats != nil {
			result = append(result, *stakerStats)
		}

		return handler.NewResult(result), nil
	}

	// Otherwise, fetch the top stakers ranked by active tvl
	paginationKey, err := handler.ParsePaginationQuery(request)
	if err != nil {
		return nil, err
	}
	topStakerStats, paginationToken, err := h.Service.GetTopStakersByActiveTvl(request.Context(), paginationKey)
	if err != nil {
		return nil, err
	}

	return handler.NewResultWithPagination(topStakerStats, paginationToken), nil
}
