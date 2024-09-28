package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/services"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"github.com/babylonlabs-io/staking-api-service/internal/utils"
	"github.com/btcsuite/btcd/chaincfg"
)

type Handler struct {
	config   *config.Config
	services *services.Services
}

type ResultOptions struct {
	Code int
}

type paginationResponse struct {
	NextKey string `json:"next_key"`
}

type PublicResponse[T any] struct {
	Data       T                   `json:"data"`
	Pagination *paginationResponse `json:"pagination,omitempty"`
}

type Result struct {
	Data   interface{}
	Status int
}

// NewResult returns a successful result, with default status code 200
func NewResultWithPagination[T any](data T, pageToken string) *Result {
	res := &PublicResponse[T]{Data: data, Pagination: &paginationResponse{NextKey: pageToken}}
	return &Result{Data: res, Status: http.StatusOK}
}

func NewResult[T any](data T) *Result {
	res := &PublicResponse[T]{Data: data}
	return &Result{Data: res, Status: http.StatusOK}
}

func New(
	ctx context.Context, cfg *config.Config, services *services.Services,
) (*Handler, error) {
	return &Handler{
		config:   cfg,
		services: services,
	}, nil
}

func parsePaginationQuery(r *http.Request) (string, *types.Error) {
	pageKey := r.URL.Query().Get("pagination_key")
	if pageKey == "" {
		return "", nil
	}
	if !utils.IsBase64Encoded(pageKey) {
		return "", types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "invalid pagination key format",
		)
	}
	return pageKey, nil
}

func parsePublicKeyQuery(r *http.Request, queryName string, isOptional bool) (string, *types.Error) {
	pkHex := r.URL.Query().Get(queryName)
	if pkHex == "" {
		if isOptional {
			return "", nil
		}
		return "", types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, queryName+" is required",
		)
	}
	_, err := utils.GetSchnorrPkFromHex(pkHex)
	if err != nil {
		return "", types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "invalid "+queryName,
		)
	}
	return pkHex, nil
}

func parseTxHashQuery(r *http.Request, queryName string) (string, *types.Error) {
	txHashHex := r.URL.Query().Get(queryName)
	if txHashHex == "" {
		return "", types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, queryName+" is required",
		)
	}
	if !utils.IsValidTxHash(txHashHex) {
		return "", types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "invalid "+queryName,
		)
	}
	return txHashHex, nil
}

func parseBtcAddressQuery(
	r *http.Request, queryName string, netParam *chaincfg.Params,
) (string, *types.Error) {
	address := r.URL.Query().Get(queryName)
	if address == "" {
		return "", types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, queryName+" is required",
		)
	}
	_, err := utils.CheckBtcAddressType(address, netParam)
	if err != nil {
		return "", types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, err.Error(),
		)
	}
	return address, nil
}

func parseBtcAddressesQuery(
	r *http.Request, queryName string, netParam *chaincfg.Params, limit int,
) ([]string, *types.Error) {
	// Get all the values for the queryName
	addresses := r.URL.Query()[queryName]
	// Check if no addresses were provided
	if len(addresses) == 0 {
		return nil, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, queryName+" is required",
		)
	}

	// Check if the number of addresses exceeds the limit
	if len(addresses) > limit {
		return nil, types.NewErrorWithMsg(
			http.StatusBadRequest,
			types.BadRequest,
			fmt.Sprintf("Maximum %d %s allowed", limit, queryName),
		)
	}

	// Validate each address
	for _, address := range addresses {
		_, err := utils.CheckBtcAddressType(address, netParam)
		if err != nil {
			return nil, types.NewErrorWithMsg(
				http.StatusBadRequest, types.BadRequest, err.Error(),
			)
		}
	}

	return addresses, nil
}

// parseStateFilterQuery parses the state filter query and returns the state enum
// If the state is not provided, it returns an empty string
func parseStateFilterQuery(
	r *http.Request, queryName string,
) (types.DelegationState, *types.Error) {
	state := r.URL.Query().Get(queryName)
	if state == "" {
		return "", nil
	}
	stateEnum, err := types.FromStringToDelegationState(state)
	if err != nil {
		return "", types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, err.Error(),
		)
	}
	return stateEnum, nil
}
