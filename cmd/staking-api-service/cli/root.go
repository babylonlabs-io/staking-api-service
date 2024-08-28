package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	defaultConfigFileName            = "config.yml"
	defaultGlobalParamsFileName      = "global_params.json"
	defaultFinalityProvidersFileName = "finality_providers.json"
)

var (
	cfgPath                   string
	globalParamsPath          string
	finalityProvidersPath     string
	replayFlag                bool
	backfillPubkeyAddressFlag bool
	rootCmd                   = &cobra.Command{
		Use: "start-server",
	}
)

func Setup() error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	defaultConfigPath := getDefaultConfigFile(homePath, defaultConfigFileName)
	defaultGlobalParamsPath := getDefaultConfigFile(homePath, defaultGlobalParamsFileName)
	defaultFinalityProvidersPath := getDefaultConfigFile(homePath, defaultFinalityProvidersFileName)

	rootCmd.PersistentFlags().StringVar(
		&cfgPath,
		"config",
		defaultConfigPath,
		fmt.Sprintf("config file (default %s)", defaultConfigPath),
	)
	rootCmd.PersistentFlags().StringVar(
		&globalParamsPath,
		"params",
		defaultGlobalParamsPath,
		fmt.Sprintf("global params file (default %s)", defaultGlobalParamsPath),
	)
	rootCmd.PersistentFlags().StringVar(
		&finalityProvidersPath,
		"finality-providers",
		defaultFinalityProvidersPath,
		fmt.Sprintf("finality providers file (default %s)", defaultFinalityProvidersPath),
	)
	rootCmd.PersistentFlags().BoolVar(
		&replayFlag,
		"replay",
		false,
		"Replay unprocessable messages",
	)
	rootCmd.PersistentFlags().BoolVar(
		&backfillPubkeyAddressFlag,
		"backfill-pubkey-address",
		false,
		"Backfill pubkey address mappings",
	)
	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}

func getDefaultConfigFile(homePath, filename string) string {
	return filepath.Join(homePath, filename)
}

func GetConfigPath() string {
	return cfgPath
}

func GetGlobalParamsPath() string {
	return globalParamsPath
}

func GetFinalityProvidersPath() string {
	return finalityProvidersPath
}

func GetReplayFlag() bool {
	return replayFlag
}

func GetBackfillPubkeyAddressFlag() bool {
	return backfillPubkeyAddressFlag
}
