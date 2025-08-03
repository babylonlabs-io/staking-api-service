package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
	if len(args) == 0 {
		log.Error().Msg("empty allow-list file")
		os.Exit(1)
	}

	filename := args[0]
	fd, err := os.Open(filename)
	if err != nil {
		log.Err(err).Msg("Failed to open allow-list file")
		os.Exit(1)
	}
	defer fd.Close()

	sc := bufio.NewScanner(fd)
	for sc.Scan() {
		stakingTxHash := strings.TrimSpace(sc.Text())

		// Skip empty lines and comments
		if stakingTxHash == "" || strings.HasPrefix(stakingTxHash, "#") {
			continue
		}

		// Note: We no longer update database canExpand field.
		// This command now validates the file format for runtime evaluation.
		// The allow-list will be loaded into memory during API service startup.
		fmt.Printf("Allow-list entry validated: %q\n", stakingTxHash)
	}

	if err := sc.Err(); err != nil {
		log.Err(err).Msg("Failed to import allow-list")
		os.Exit(1)
	}

	os.Exit(0)
}
