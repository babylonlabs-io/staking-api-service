package config

import "fmt"

type AllowList struct {
	ExpirationBlock uint64 `mapstructure:"expiration_block"`
}

func (cfg *AllowList) Validate() error {
	if cfg.ExpirationBlock == 0 {
		return fmt.Errorf("allow-list: expiration block cannot be zero")
	}

	return nil
}
