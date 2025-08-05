package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// NewAllowList loads allow-list from JSON file and returns a map lookup.
// Returns an empty map if the file path is empty.
// Expects JSON file to be an array of staking transaction hashes
func NewAllowList(path string) (map[string]bool, error) {
	if path == "" {
		log.Debug().Msg("No allow-list path provided, canExpand will use default logic")
		return make(map[string]bool), nil
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read allow-list file %q: %w", path, err)
	}

	var stakingHashes []string
	err = json.Unmarshal(data, &stakingHashes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse allow-list JSON from %q: %w", path, err)
	}

	// Convert slice to map for fast lookup in runtime
	allowList := make(map[string]bool, len(stakingHashes))
	for _, hash := range stakingHashes {
		if hash != "" { // Skip empty strings
			allowList[hash] = true
		}
	}

	log.Info().
		Int("count", len(allowList)).
		Str("file", path).
		Msg("Allow-list loaded successfully from JSON file")

	return allowList, nil
}
