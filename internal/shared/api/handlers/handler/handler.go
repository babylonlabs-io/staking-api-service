package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/bbnclient"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services/service"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	"github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Handler struct {
	Config    *config.Config
	Service   service.SharedServiceProvider
	bbnClient *bbnclient.BBNClient
}

func New(config *config.Config, service service.SharedServiceProvider) (*Handler, error) {
	var bbnClient *bbnclient.BBNClient
	if config.BBN != nil {
		var err error
		bbnClient, err = bbnclient.New(config.BBN)
		if err != nil {
			return nil, err
		}
	}

	return &Handler{Config: config, Service: service, bbnClient: bbnClient}, nil
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

func ValidateBabylonAddress(address string) error {
	if len(strings.TrimSpace(address)) == 0 {
		return errors.New("empty address string is not allowed")
	}

	const babylonPrefix = "bbn"
	bz, err := sdk.GetFromBech32(address, babylonPrefix)
	if err != nil {
		return err
	}

	return sdk.VerifyAddressFormat(bz)
}

func ParseBabylonAddressQuery(
	r *http.Request, queryName string, isOptional bool,
) (*string, *types.Error) {
	address := r.URL.Query().Get(queryName)
	if address == "" {
		if isOptional {
			return nil, nil
		}
		return nil, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, queryName+" is required",
		)
	}
	err := ValidateBabylonAddress(address)
	if err != nil {
		return nil, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "invalid "+queryName,
		)
	}
	return &address, nil
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
	txHashHex = strings.ToLower(txHashHex)

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

// ParseBooleanQuery parses the boolean query and returns the boolean value
// If the boolean is not provided, it returns false
// If the boolean is not valid, it returns an error
func ParseBooleanQuery(
	r *http.Request, queryName string, isOptional bool,
) (bool, *types.Error) {
	value := r.URL.Query().Get(queryName)
	if value == "" {
		if isOptional {
			return false, nil
		}
		return false, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, queryName+" is required",
		)
	}
	if value != "true" && value != "false" {
		return false, types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "invalid "+queryName,
		)
	}
	return value == "true", nil
}
