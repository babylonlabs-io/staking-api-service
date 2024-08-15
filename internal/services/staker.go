package services

import (
	"context"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"github.com/babylonlabs-io/staking-api-service/internal/utils"
	"github.com/rs/zerolog/log"
)

// Given the staker public key, transform into multiple btc addresses and save them in the db.
func (s *Services) ProcessAndSaveBtcAddresses(
	ctx context.Context, stakerPkHex string,
) *types.Error {
	// Prepare the btc addresses
	addresses, err := utils.DeriveAddressesFromNoCoordPk(stakerPkHex, s.cfg.Server.BTCNetParam)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to derive addresses from staker pk")
		return types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "failed to derive addresses from staker pk",
		)
	}

	// Try to save the btc addresses, ignore if they already exist
	err = s.DbClient.InsertPkAddressMappings(
		ctx, stakerPkHex, addresses.Taproot,
		addresses.NativeSegwitOdd, addresses.NativeSegwitEven,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to save btc addresses")
		return types.NewInternalServiceError(err)
	}
	return nil
}
