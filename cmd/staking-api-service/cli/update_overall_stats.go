package cli

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func UpdateOverallStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "update-overall-stats",
		Run: updateOverallStats,
	}

	cmd.Flags().Bool("dry-run", false, "Run in simulation mode without making changes")

	return cmd
}

func updateOverallStats(cmd *cobra.Command, args []string) {
	err := updateOverallStatsE(cmd, args)
	// because of current architecture we need to stop execution of the program
	// otherwise existing main logic will be called
	if err != nil {
		log.Err(err).Msg("Failed to update overall stats")
		os.Exit(1)
	}

	os.Exit(0)
}

func updateOverallStatsE(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	cfg, err := config.New(GetConfigPath())
	if err != nil {
		return err
	}

	dbClients, err := dbclients.New(ctx, cfg)
	if err != nil {
		return err
	}

	stats, err := dbClients.V1DBClient.GetOverallStats(ctx)
	if err != nil {
		return err
	}
	_ = stats
	_ = dryRun // don't do any modifications if dryRun is passed

	// which query
	// dbClients.V2DBClient.IncrementOverallStats()

	return nil
}
