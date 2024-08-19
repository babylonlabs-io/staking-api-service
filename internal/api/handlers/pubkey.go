package handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
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

// GetPubKeys @Summary Get stakers' public keys
// @Description Retrieves public keys for the given BTC addresses. This endpoint
// only returns public keys for addresses that have associated delegations in
// the system. If an address has no associated delegation, it will not be
// included in the response. Supports both Taproot and Native Segwit addresses.
// @Produce json
// @Param address query []string true "List of BTC addresses to look up (up to 10), currently only supports Taproot and Native Segwit addresses"
// @Success 200 {object} Result[map[string]string] "A map of BTC addresses to their corresponding public keys (only addresses with delegations are returned)"
// @Failure 400 {object} types.Error "Bad Request: Invalid input parameters"
// @Failure 500 {object} types.Error "Internal Server Error"
// @Router /v1/staker/pubkey-lookup [get]
func (h *Handler) GetPubKeys(request *http.Request) (*Result, *types.Error) {
	addresses, err := parseBtcAddressesQuery(
		request, "address", h.config.Server.BTCNetParam, MAX_NUM_PK_LOOKUP_ADDRESSES,
	)
	if err != nil {
		return nil, err
	}

	// Get the public keys for the given addresses
	addressToPkMapping, err := h.services.GetStakerPublicKeysByAddresses(request.Context(), addresses)
	if err != nil {
		return nil, err
	}

	return NewResult(addressToPkMapping), nil
}
