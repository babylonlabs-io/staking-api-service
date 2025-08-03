package services

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/rs/zerolog/log"
)

// loadAllowList loads allow-list from configuration at the application level.
// Returns a non-nil map, empty if no allow-list is configured.
func loadAllowList(cfg *config.Config) map[string]bool {
	if cfg.AllowList == nil {
		log.Debug().Msg("No allow-list configured, canExpand will default to true for Active delegations with >1 finality providers")
		return make(map[string]bool)
	}

	stakingHashes, err := loadAllowListFile(cfg.AllowList.FilePath)
	if err != nil {
		log.Error().Err(err).Str("path", cfg.AllowList.FilePath).Msg("Failed to load allow-list file, continuing without allow-list")
		return make(map[string]bool)
	}

	allowList := make(map[string]bool, len(stakingHashes))
	for _, hash := range stakingHashes {
		allowList[hash] = true
	}

	log.Info().
		Int("count", len(stakingHashes)).
		Str("file", cfg.AllowList.FilePath).
		Msg("Allow-list loaded successfully during application initialization")

	return allowList
}

func loadAllowListFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open allow-list file %q: %w", filePath, err)
	}
	defer file.Close()

	var stakingHashes []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		stakingHashes = append(stakingHashes, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading allow-list file: %w", err)
	}

	return stakingHashes, nil
}
