package cli

import (
	"os"

	"fmt"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func ImportAllowListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-allow-list",
		Short: "Import allow-list",
		Run:   importAllowList,
	}

	return cmd
}

func importAllowList(cmd *cobra.Command, args []string) {
	err := importAllowListE(cmd, args)
	// because of current architecture we need to stop execution of the program
	// otherwise existing main logic will be called
	if err != nil {
		log.Err(err).Msg("Failed to update overall stats")
		os.Exit(1)
	}

	os.Exit(0)
}

func importAllowListE(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	cfg, err := config.New(GetConfigPath())
	if err != nil {
		return err
	}

	dbClients, err := dbclients.New(ctx, cfg)
	if err != nil {
		return err
	}
	sharedDB := dbClients.SharedDBClient

	log.Info().Msg("Importing allow-list")
	stakingTxHashes := []string{
		"02000000000101933cf39b909c9a6b07925ccfadf088559afdac369ac7cec9702e075d3bbeb6090200000000fdffffff0350c30000000000002251201f719f4cacb26c3059abffa5538ee0bbba6d028a9c0a6f754cf46b3f986040f90000000000000000496a4762627434001e06e1ef408126703ed66447cd6972434396b252a22e843d8295d55ae7a9cfd1c20acf33c17e5a6c92cced9f1d530cccab7aa3e53400456202f02fac95e9c481fa00f8fa0300000000002251207754ae229380bdacf33b9011e997655dfb042e01a5ea07ce98b2e716e05b7df601405711f80b8bff9d4751680809b3eea627b042ec82fb12bced7d110f9d9e251fda83e3f3d181de19bc6646efc72ea3ae41bc8043adc20db112f687111368dfb77408080300",
	}
	for _, stakingTxHash := range stakingTxHashes {
		err = sharedDB.SaveTxInAllowList(ctx, stakingTxHash)
		if err != nil {
			fmt.Printf("Failed to save staking tx %q in allow-list: %v\n", stakingTxHash, err)
		}
	}

	return nil
}
