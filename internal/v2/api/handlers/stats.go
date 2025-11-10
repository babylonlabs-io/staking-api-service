package v2handlers

import (
	"net/http"
	"strconv"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetStakerStats gets staker stats for babylon staking
//
//	@Summary		Get Staker Stats
//	@Description	Fetches staker stats for babylon staking including active tvl,
//
// active delegations, unbonding tvl, unbonding delegations, withdrawable tvl,
// withdrawable delegations, slashed tvl and slashed delegations. If the babylon
// address is not provided, the stats will be calculated for all the delegations
// of the staker based on the staker's BTC public key.
//
//	@Produce		json
//	@Tags			v2
//	@Param			staker_pk_hex	query		string												true	"Public key of the staker to fetch"
//	@Param			babylon_address	query		string												false	"Babylon address of the staker to fetch"
//	@Success		200				{object}	handler.PublicResponse[v2service.StakerStatsPublic]	"Staker stats"
//	@Failure		400				{object}	types.Error											"Error: Bad Request"
//	@Failure		404				{object}	types.Error											"Error: Not Found"
//	@Failure		500				{object}	types.Error											"Error: Internal Server Error"
//	@Router			/v2/staker/stats [get]
func (h *V2Handler) GetStakerStats(request *http.Request) (*handler.Result, *types.Error) {
	stakerPKHex, err := handler.ParsePublicKeyQuery(request, "staker_pk_hex", false)
	if err != nil {
		return nil, err
	}

	bbnAddress, err := handler.ParseBabylonAddressQuery(
		request, "babylon_address", true,
	)
	if err != nil {
		return nil, err
	}

	stats, err := h.Service.GetStakerStats(
		request.Context(), stakerPKHex, bbnAddress,
	)
	if err != nil {
		return nil, err
	}
	return handler.NewResult(stats), nil
}

// GetStats @Summary Get overall system stats
//
//	@Description	Overall system stats including max staking APR (BTC + co-staking)
//	@Produce		json
//	@Tags			v2
//	@Success		200	{object}	handler.PublicResponse[v2service.OverallStatsPublic]	""
//	@Failure		400	{object}	types.Error												"Error: Bad Request"
//	@Router			/v2/stats [get]
func (h *V2Handler) GetOverallStats(request *http.Request) (*handler.Result, *types.Error) {
	stats, err := h.Service.GetOverallStats(request.Context())
	if err != nil {
		return nil, err
	}
	return handler.NewResult(stats), nil
}

// GetPrices @Summary Get latest prices for all available symbols
//
//	@Description	Get latest prices for all available symbols
//	@Produce		json
//	@Tags			v2
//	@Success		200	{object}	handler.PublicResponse[map[string]float64]	""
//	@Failure		400	{object}	types.Error									"Error: Bad Request"
//	@Router			/v2/prices [get]
func (h *V2Handler) GetPrices(request *http.Request) (*handler.Result, *types.Error) {
	prices, err := h.Service.GetLatestPrices(request.Context())
	if err != nil {
		return nil, err
	}

	return handler.NewResult(prices), nil
}

// GetAPR is used by frontend app and external partners (e.g., wallets)
//
//	@Summary		Get personalized staking APR
//	@Description	Get personalized staking APR based on user's BTC and BABY stake amounts
//	@Produce		json
//	@Tags			v2
//	@Param			satoshis_staked	query		int														false	"Total satoshis staked (confirmed + pending)"	default(0)
//	@Param			ubbn_staked	query		int														false	"Total ubbn staked (confirmed + pending)"	default(0)
//	@Success		200			{object}	handler.PublicResponse[v2service.StakingAPRPublic]		""
//	@Failure		400			{object}	types.Error												"Error: Bad Request"
//	@Router			/v2/apr [get]
func (h *V2Handler) GetAPR(request *http.Request) (*handler.Result, *types.Error) {
	// Parse satoshis_staked parameter (optional, defaults to 0)
	satoshisStaked, err := parseInt64Query(request, "satoshis_staked", true)
	if err != nil {
		return nil, err
	}

	// Parse ubbn_staked parameter (optional, defaults to 0)
	ubbnStaked, err := parseInt64Query(request, "ubbn_staked", true)
	if err != nil {
		return nil, err
	}

	stakingAPR, serviceErr := h.Service.GetStakingAPR(request.Context(), satoshisStaked, ubbnStaked)
	if serviceErr != nil {
		return nil, serviceErr
	}

	return handler.NewResult(stakingAPR), nil
}

// parseInt64Query parses an int64 query parameter
func parseInt64Query(r *http.Request, paramName string, isOptional bool) (int64, *types.Error) {
	value := r.URL.Query().Get(paramName)

	// If parameter is missing
	if value == "" {
		if isOptional {
			return 0, nil // Return 0 as default for optional parameters
		}
		return 0, types.NewErrorWithMsg(
			http.StatusBadRequest,
			types.BadRequest,
			"missing required query parameter: "+paramName,
		)
	}

	// Parse the value
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, types.NewErrorWithMsg(
			http.StatusBadRequest,
			types.BadRequest,
			"invalid "+paramName+": must be a valid integer",
		)
	}

	// Validate non-negative
	if parsed < 0 {
		return 0, types.NewErrorWithMsg(
			http.StatusBadRequest,
			types.BadRequest,
			"invalid "+paramName+": must be non-negative",
		)
	}

	return parsed, nil
}
