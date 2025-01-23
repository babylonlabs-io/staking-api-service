package v1handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

const (
	// MAX_NUM_PK_LOOKUP_ADDRESSES is the maximum number of addresses that can be queried
	// in a single request. This limit helps prevent the URL from becoming too long
	// and potentially being rejected by the server or browser. While this is a soft
	// limit that can be increased if needed, we are setting it conservatively at 10
	// to ensure compatibility. Given a URL length limit of 2048 characters and a
	// worst-case scenario where each address is 64 characters long, we can support
	// up to 28 addresses. However, we limit it to 10 for added safety.
	MAX_NUM_PK_LOOKUP_ADDRESSES = 10
)

// GetPubKeys godoc
// @Summary Get stakers' public keys
// @Description Retrieves public keys for the given BTC addresses. This endpoint
// @Description only returns public keys for addresses that have associated delegations in
// @Description the system. If an address has no associated delegation, it will not be
// @Description included in the response. Supports both Taproot and Native Segwit addresses.
// @Produce json
// @Tags shared
// @Param address query []string true "List of BTC addresses to look up (up to 10), currently only supports Taproot and Native Segwit addresses" collectionFormat(multi)
// @Success 200 {object} handler.PublicResponse[map[string]string] "A map of BTC addresses to their corresponding public keys (only addresses with delegations are returned)"
// @Failure 400 {object} types.Error "Bad Request: Invalid input parameters"
// @Failure 500 {object} types.Error "Internal Server Error"
// @Router /v1/staker/pubkey-lookup [get]
func (h *V1Handler) GetPubKeys(request *http.Request) (*handler.Result, *types.Error) {
	addresses, err := handler.ParseBtcAddressesQuery(
		request, "address", h.Handler.Config.Server.BTCNetParam, MAX_NUM_PK_LOOKUP_ADDRESSES,
	)
	if err != nil {
		return nil, err
	}

	// Get the public keys for the given addresses
	addressToPkMapping, err := h.Service.GetStakerPublicKeysByAddresses(request.Context(), addresses)
	if err != nil {
		return nil, err
	}

	return handler.NewResult(addressToPkMapping), nil
}
