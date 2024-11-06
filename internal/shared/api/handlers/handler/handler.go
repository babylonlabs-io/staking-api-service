package handler

import (
	"context"
	"fmt"
	"net/http"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	"github.com/btcsuite/btcd/chaincfg"
)

type Handler struct {
	Config  *config.Config
	Service service.SharedServiceProvider
}

func New(ctx context.Context, config *config.Config, service service.SharedServiceProvider) (*Handler, error) {
	return &Handler{Config: config, Service: service}, nil
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

func ParsePaginationQuery(r *http.Request) (string, *types.Error) {
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

func ParsePublicKeyQuery(r *http.Request, queryName string, isOptional bool) (string, *types.Error) {
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

func ParseTxHashQuery(r *http.Request, queryName string) (string, *types.Error) {
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

func ParseBtcAddressQuery(
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

func ParseBtcAddressesQuery(
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

// ParseStateFilterQuery parses the state filter query and returns the state enum
// If the state is not provided, it returns an empty string
func ParseStateFilterQuery(
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

func (h *Handler) ParseFPSearchQuery(r *http.Request, queryName string, isOptional bool) (string, *types.Error) {
	str := r.URL.Query().Get(queryName)
	if str == "" {
		if isOptional {
			return "", nil
		}
		return "", types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, queryName + " is required",
		)
	}

	if len(str) < 1 || len(str) > h.Config.Server.MaxSearchQueryLength {
		return "", types.NewErrorWithMsg(
			http.StatusBadRequest,
			types.BadRequest,
			fmt.Sprintf("search query must be between 1 and %d characters", h.Config.Server.MaxSearchQueryLength),
		)
	}

	for _, char := range str {
		if char < 32 || char > 126 {
			return "", types.NewErrorWithMsg(
				http.StatusBadRequest,
				types.BadRequest, 
				fmt.Sprintf("%s contains invalid characters", queryName),
			)
		}
	}

	return str, nil
}

func ParseFPStateQuery(r *http.Request, queryName string, isOptional bool) (types.FinalityProviderState, *types.Error) {
	state := r.URL.Query().Get(queryName)
	if state == "" {
		if isOptional {
			return "", nil
		}
	}
	stateEnum, err := indexerdbmodel.FromStringToFinalityProviderState(state)
	if err != nil {
		return "", types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, err.Error(),
		)
	}
	return stateEnum, nil
}
