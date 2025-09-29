package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2service "github.com/babylonlabs-io/staking-api-service/internal/v2/service"
)

// GetDelegation @Summary Get a delegation
//
//	@Summary		Get a delegation
//	@Description	Retrieves a delegation by a given transaction hash
//	@Produce		json
//	@Tags			v2
//	@Param			staking_tx_hash_hex	query		string												true	"Staking transaction hash in hex format"
//	@Success		200					{object}	handler.PublicResponse[v2service.DelegationPublic]	"Staker delegation"
//	@Failure		400					{object}	types.Error											"Error: Bad Request"
//	@Failure		404					{object}	types.Error											"Error: Not Found"
//	@Failure		500					{object}	types.Error											"Error: Internal Server Error"
//	@Router			/v2/delegation [get]
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
//
//		@Summary		Get Delegations
//		@Description	Fetches delegations for babylon staking including tvl, total delegations, active tvl, active delegations and total stakers.
//		@Produce		json
//		@Tags			v2
//		@Param			staker_pk_hex	query		string														false	"Staker public key in hex format. If omitted, babylon_address is used as the main query parameter and state is required."
//		@Param			babylon_address	query		string														false	"Babylon address. Required if staker_pk_hex is omitted."
//		@Param			state			query		string														false	"State of delegations (only 'active' is supported). Required if staker_pk_hex is omitted."
//		@Param			pagination_key	query		string														false	"Pagination key to fetch the next page of delegations"
//		@Success		200				{object}	handler.PublicResponse[[]v2service.DelegationPublic]{array}	"List of staker delegations and pagination token"
//		@Failure		400				{object}	types.Error													"Error: Bad Request"
//		@Failure		404				{object}	types.Error													"Error: Not Found"
//		@Failure		500				{object}	types.Error													"Error: Internal Server Error"
//		@Router			/v2/delegations [get]
func (h *V2Handler) GetDelegations(request *http.Request) (*handler.Result, *types.Error) {
	stakerPKHex, err := handler.ParsePublicKeyQuery(request, "staker_pk_hex", true)
	if err != nil {
		return nil, err
	}

	bbnAddress, err := handler.ParseBabylonAddressQuery(
		request, "babylon_address", true,
	)
	if err != nil {
		return nil, err
	}

	if stakerPKHex == "" && bbnAddress == nil {
		return nil, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "staker_pk_hex or babylon_address is required",
		)
	}

	paginationKey, err := handler.ParsePaginationQuery(request)
	if err != nil {
		return nil, err
	}

	ctx := request.Context()

	var (
		delegations         []*v2service.DelegationPublic
		paginationKeyResult string // pagination key that is returned from API
	)
	if stakerPKHex == "" {
		// if staker pk is omitted then babylon_address is the main query param and state is required as well
		state, parseErr := handler.ParseDelegationStateQuery(request)
		if parseErr != nil {
			return nil, parseErr
		}

		// only support active state for now
		if state != types.Active {
			return nil, types.NewErrorWithMsg(
				http.StatusBadRequest, types.BadRequest, "state is not supported",
			)
		}

		delegations, paginationKeyResult, err = h.Service.GetDelegationsByBabylonAddress(ctx, *bbnAddress, state, paginationKey)
	} else {
		delegations, paginationKeyResult, err = h.Service.GetDelegationsByStakerPKHex(
			request.Context(), stakerPKHex, bbnAddress, paginationKey,
		)
	}

	if err != nil {
		return nil, err
	}

	return handler.NewResultWithPagination(delegations, paginationKeyResult), nil
}
