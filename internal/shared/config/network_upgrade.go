package config

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// POPUpgrade represents a single POP upgrade configuration
type POPUpgrade struct {
	Height  uint64 `mapstructure:"height"`
	Version uint64 `mapstructure:"version"`
}

// Validate ensures POP upgrade configuration is valid
func (pop *POPUpgrade) Validate() error {
	if pop == nil {
		return nil // POP upgrade is optional
	}

	// If POP exists, height must be configured
	if pop.Height == 0 {
		return fmt.Errorf("POP upgrade height is required when POP is configured")
	}

	// Version can be 0, so we don't need to validate it's greater than 0
	log.Info().Uint64("height", pop.Height).Uint64("version", pop.Version).Msg("POP upgrade configured")
	return nil
}

type NetworkUpgrade struct {
	POP       []POPUpgrade `mapstructure:"pop,omitempty"`
	AllowList *AllowList   `mapstructure:"allow-list"`
}

// network upgrade config is optional.
func (cfg *NetworkUpgrade) Validate() error {
	if cfg == nil {
		log.Info().Msg("empty network upgrade config")
		// Empty network upgrade config is valid
		return nil
	}

	// Validate each POP upgrade if configured
	for i, pop := range cfg.POP {
		if err := pop.Validate(); err != nil {
			return fmt.Errorf("POP upgrade %d validation failed: %w", i, err)
		}
	}

	if cfg.AllowList != nil {
		return cfg.AllowList.Validate()
	}

	return nil
}
