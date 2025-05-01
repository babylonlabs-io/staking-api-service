package config

import (
	"fmt"
	"net/url"
	"time"
)

type BBNConfig struct {
	RPCAddr       string        `mapstructure:"rpc-addr"`
	Timeout       time.Duration `mapstructure:"timeout"`
	MaxRetryTimes uint          `mapstructure:"maxretrytimes"`
	RetryInterval time.Duration `mapstructure:"retryinterval"`
}

func (cfg *BBNConfig) Validate() error {
	if _, err := url.Parse(cfg.RPCAddr); err != nil {
		return fmt.Errorf("cfg.RPCAddr is not correctly formatted: %w", err)
	}

	if cfg.Timeout <= 0 {
		return fmt.Errorf("cfg.Timeout must be positive")
	}

	if cfg.MaxRetryTimes <= 0 {
		return fmt.Errorf("cfg.MaxRetryTimes must be positive")
	}

	if cfg.RetryInterval <= 0 {
		return fmt.Errorf("cfg.RetryInterval must be positive")
	}

	return nil
}
