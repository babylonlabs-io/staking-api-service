package v1service

import (
	"context"
	"net/http"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	"github.com/rs/zerolog/log"
)

type PublicKeyAddressMappingPublic struct {
	PkHex   string `json:"pk_hex"`
	Address string `json:"address"`
}

// Given the staker public key, transform into multiple btc addresses and save them in the db.
func (s *V1Service) ProcessAndSaveBtcAddresses(
	ctx context.Context, stakerPkHex string,
) *types.Error {
	// Prepare the btc addresses
	addresses, err := utils.DeriveAddressesFromNoCoordPk(
		stakerPkHex, s.Service.Cfg.Server.BTCNetParam,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to derive addresses from staker pk")
		return types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest,
			"failed to derive addresses from staker pk",
		)
	}

	// Try to save the btc addresses, ignore if they already exist
	err = s.Service.DbClients.V1DBClient.InsertPkAddressMappings(
		ctx, stakerPkHex, addresses.Taproot,
		addresses.NativeSegwitOdd, addresses.NativeSegwitEven,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to save btc addresses")
		return types.NewInternalServiceError(err)
	}
	return nil
}

// GetStakerPublicKeysByAddresses retrieves the corresponding public keys for a
// list of BTC addresses. It handles both Taproot and Native Segwit addresses by
// first categorizing the addresses, then querying the database for the
// corresponding public keys. The results are returned as a map where the keys
// are the addresses and the values are the corresponding public keys.
// TODO: extract this to a common function in util
func (s *V1Service) GetStakerPublicKeysByAddresses(
	ctx context.Context, addresses []string,
) (map[string]string, *types.Error) {
	// Split the addresses into taproot and native segwit
	var taprootAddresses, nativeSegwitAddresses []string
	for _, addr := range addresses {
		addressType, err := utils.CheckBtcAddressType(addr, s.Service.Cfg.Server.BTCNetParam)
		if err != nil {
			return nil, types.NewErrorWithMsg(
				http.StatusBadRequest, types.BadRequest, "invalid btc address",
			)
		}
		if addressType == utils.Taproot {
			taprootAddresses = append(taprootAddresses, addr)
		} else if addressType == utils.NativeSegwit {
			nativeSegwitAddresses = append(nativeSegwitAddresses, addr)
		} else {
			return nil, types.NewErrorWithMsg(
				http.StatusBadRequest, types.BadRequest, "unsupported address type",
			)
		}
	}

	// map of address to public key
	addressPkMapping := make(map[string]string)
	// Get the public keys from the db by taproot addresses
	if len(taprootAddresses) > 0 {
		mappings, err := s.Service.DbClients.V1DBClient.FindPkMappingsByTaprootAddress(
			ctx, taprootAddresses,
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).
				Msg("Failed to get pk mappings by taproot address")
			return nil, types.NewInternalServiceError(err)
		}
		for _, mapping := range mappings {
			addressPkMapping[mapping.Taproot] = mapping.PkHex
		}
	}
	// Get the public keys from the db by native segwit addresses
	if len(nativeSegwitAddresses) > 0 {
		mappings, err := s.Service.DbClients.V1DBClient.FindPkMappingsByNativeSegwitAddress(
			ctx, nativeSegwitAddresses,
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).
				Msg("Failed to get pk mappings by native segwit address")
			return nil, types.NewInternalServiceError(err)
		}
		// Map each Native Segwit address to its corresponding public key
		for _, nativeSegwitAddress := range nativeSegwitAddresses {
			pkHex := findPublicKeyByNativeSegwitAddress(nativeSegwitAddress, mappings)
			if pkHex != "" {
				addressPkMapping[nativeSegwitAddress] = pkHex
			}
		}
	}
	return addressPkMapping, nil
}

// findPublicKeyByNativeSegwitAddress searches for the corresponding public key
// in the provided mappings for a given Native Segwit address.
// It checks both the "NativeSegwitEven" and "NativeSegwitOdd" fields to find a
// match.
func findPublicKeyByNativeSegwitAddress(
	providedNativeSegwitAddress string, mappings []*dbmodel.PkAddressMapping,
) string {
	for _, mapping := range mappings {
		if mapping.NativeSegwitEven == providedNativeSegwitAddress ||
			mapping.NativeSegwitOdd == providedNativeSegwitAddress {
			return mapping.PkHex
		}
	}
	return ""
}
