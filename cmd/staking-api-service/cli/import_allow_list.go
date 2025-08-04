package cli

import (
	"fmt"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func ImportAllowListCmd() *cobra.Command {
	var allowListPath string

	cmd := &cobra.Command{
		Use:   "import-allow-list",
		Short: "Import and validate allow-list file",
		Long:  "Imports and validates the format of an allow-list file, checking for proper formatting and duplicate entries.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return importAllowList(allowListPath)
		},
	}

	cmd.Flags().StringVar(
		&allowListPath,
		"file",
		"",
		"Path to the allow-list file to import and validate (required)",
	)
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func importAllowList(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("allow-list file path is required")
	}

	log.Info().Str("file", filePath).Msg("Starting allow-list import and validation")

	// Load and validate the allow-list file
	allowList, err := types.NewAllowList(filePath)
	if err != nil {
		log.Error().Err(err).Str("file", filePath).Msg("Failed to load allow-list file")
		return fmt.Errorf("failed to load allow-list file: %w", err)
	}

	// Check for duplicates and validate format
	if len(allowList) == 0 {
		log.Warn().Str("file", filePath).Msg("Allow-list file is empty or contains no valid entries")
	} else {
		log.Info().
			Int("entries", len(allowList)).
			Str("file", filePath).
			Msg("Allow-list file validation completed successfully")
	}

	// Log some sample entries (first 5) for verification
	count := 0
	for hash := range allowList {
		if count >= 5 {
			break
		}
		log.Debug().Str("entry", hash).Msg("Sample allow-list entry")
		count++
	}

	return nil
}
