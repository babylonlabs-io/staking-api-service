package types

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

// NewAllowList loads allow-list from file path and returns a map for O(1) lookup.
// Returns an empty map if the file path is empty or file doesn't exist.
func NewAllowList(path string) (map[string]bool, error) {
	if path == "" {
		log.Debug().Msg("No allow-list path provided, canExpand will use default logic")
		return make(map[string]bool), nil
	}

	stakingHashes, err := loadAllowListFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load allow-list from %q: %w", path, err)
	}

	allowList := make(map[string]bool, len(stakingHashes))
	for _, hash := range stakingHashes {
		allowList[hash] = true
	}

	log.Info().
		Int("count", len(stakingHashes)).
		Str("file", path).
		Msg("Allow-list loaded successfully during application initialization")

	return allowList, nil
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
