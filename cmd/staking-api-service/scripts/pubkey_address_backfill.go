package scripts

import (
	"context"
	"fmt"

	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db"
	v1db "github.com/babylonlabs-io/staking-api-service/internal/db/v1"
	"github.com/babylonlabs-io/staking-api-service/internal/utils"
	"github.com/rs/zerolog/log"
)

func BackfillPubkeyAddressesMappings(ctx context.Context, cfg *config.Config) error {
	client, err := db.NewMongoClient(ctx, cfg.Db)
	if err != nil {
		return fmt.Errorf("failed to create db client: %w", err)
	}
	v1dbClient, err := v1db.New(ctx, client, cfg.Db)
	if err != nil {
		return fmt.Errorf("failed to create db client: %w", err)
	}
	pageToken := ""
	var count int
	for {
		result, err := v1dbClient.ScanDelegationsPaginated(ctx, pageToken)
		if err != nil {
			return fmt.Errorf("failed to scan delegations: %w", err)
		}
		for _, delegation := range result.Data {
			addresses, err := utils.DeriveAddressesFromNoCoordPk(
				delegation.StakerPkHex, cfg.Server.BTCNetParam,
			)
			if err != nil {
				return fmt.Errorf("failed to derive btc addresses: %w", err)
			}
			if err := v1dbClient.InsertPkAddressMappings(
				ctx, delegation.StakerPkHex, addresses.Taproot,
				addresses.NativeSegwitOdd, addresses.NativeSegwitEven,
			); err != nil {
				return fmt.Errorf("failed to save btc addresses: %w", err)
			}
			log.Info().Msgf("Saved btc addresses for staker %s", delegation.StakerPkHex)
			count++
		}
		pageToken = result.PaginationToken
		if pageToken == "" {
			break
		}
	}
	log.Info().Msgf("Backfilled %d pubkey addresses mappings", count)
	return nil
}
