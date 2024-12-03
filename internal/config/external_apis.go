package config

import (
	"fmt"
	"time"
)

type ExternalAPIsConfig struct {
	CoinMarketCap *CoinMarketCapConfig `mapstructure:"coinmarketcap"`
}

type CoinMarketCapConfig struct {
	APIKey  string        `mapstructure:"api_key"`
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

func (cfg *ExternalAPIsConfig) Validate() error {
	if cfg.CoinMarketCap == nil {
		return fmt.Errorf("missing coinmarketcap config")
	}

	if err := cfg.CoinMarketCap.Validate(); err != nil {
		return err
	}

	return nil
}

func (cfg *CoinMarketCapConfig) Validate() error {
	if cfg.APIKey == "" {
		return fmt.Errorf("missing coinmarketcap api key")
	}

	if cfg.BaseURL == "" {
		return fmt.Errorf("missing coinmarketcap base url")
	}

	if cfg.Timeout <= 0 {
		return fmt.Errorf("invalid coinmarketcap timeout")
	}

	return nil
}
