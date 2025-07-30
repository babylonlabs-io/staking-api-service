package config

import "fmt"

type AllowList struct {
	ActivationBlock uint64 `mapstructure:"activation_block"`
	ExpirationBlock uint64 `mapstructure:"expiration_block"`
}

func (cfg *AllowList) Validate() error {
	if cfg.ActivationBlock >= cfg.ExpirationBlock {
		return fmt.Errorf("activation block (%d) must be less than expiration block (%d)", cfg.ActivationBlock, cfg.ExpirationBlock)
	}

	return nil
}
