package config

import (
	"fmt"
	"time"
)

type ExternalAPIsConfig struct {
	CoinMarketCap *CoinMarketCapConfig `mapstructure:"coinmarketcap"`
}

type CoinMarketCapConfig struct {
	APIKey   string        `mapstructure:"api_key"`
	BaseURL  string        `mapstructure:"base_url"`
	Timeout  time.Duration `mapstructure:"timeout"`
	CacheTTL time.Duration `mapstructure:"cache_ttl"`
}

func (cfg *ExternalAPIsConfig) Validate() error {
	if cfg.CoinMarketCap == nil {
		return fmt.Errorf("missing coinmarketcap config")
	}

	return cfg.CoinMarketCap.Validate()
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

	if cfg.CacheTTL <= 0 {
		return fmt.Errorf("invalid coinmarketcap cache ttl")
	}

	return nil
}
