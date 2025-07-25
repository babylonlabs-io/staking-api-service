package config

import "fmt"

type NetworkUpgrade struct {
	POPUpgradeHeight uint64 `mapstructure:"pop_upgrade_height"`
}

func (cfg *NetworkUpgrade) Validate() error {
	if cfg == nil {
		return fmt.Errorf("empty network upgrade config")
	}
	
	if cfg.POPUpgradeHeight == 0 {
		return fmt.Errorf("POP upgrade height must be greater than zero")
	}

	return nil
}
