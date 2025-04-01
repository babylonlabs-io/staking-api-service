package cli

import (
	"os"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func UpdateLegacyOverallStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-legacy-overall-stats",
		Short: "Update legacy overall stats",
		Run:   updateLegacyOverallStats,
	}

	return cmd
}

func updateLegacyOverallStats(cmd *cobra.Command, args []string) {
	err := updateLegacyOverallStatsE(cmd, args)
	// because of current architecture we need to stop execution of the program
	// otherwise existing main logic will be called
	if err != nil {
		log.Err(err).Msg("Failed to update overall stats")
		os.Exit(1)
	}

	os.Exit(0)
}

func updateLegacyOverallStatsE(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	cfg, err := config.New(GetConfigPath())
	if err != nil {
		return err
	}

	dbClients, err := dbclients.New(ctx, cfg)
	if err != nil {
		return err
	}

	log.Info().Msg("Updating overall stats")
	stats, err := dbClients.V1DBClient.UpdateLegacyOverallStats(ctx)
	if err != nil {
		return err
	}

	log.Info().Msgf("Updated overall stats: %+v", stats)

	return nil
}
