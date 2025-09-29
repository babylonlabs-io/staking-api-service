package config

import (
	"fmt"
	"time"
)

type ExternalAPIsConfig struct {
	CoinMarketCap *CoinMarketCapConfig `mapstructure:"coinmarketcap"`
	Chainalysis   *ChainAnalysisConfig `mapstructure:"chainalysis"`
}

type CoinMarketCapConfig struct {
	APIKey   string        `mapstructure:"api_key"`
	BaseURL  string        `mapstructure:"base_url"`
	Timeout  time.Duration `mapstructure:"timeout"`
	CacheTTL time.Duration `mapstructure:"cache_ttl"`
}

type ChainAnalysisConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

func (cfg *ExternalAPIsConfig) Validate() error {
	if cfg == nil {
		return fmt.Errorf("external api configuration is required")
	}

	// coinmarketcap is required
	err := cfg.CoinMarketCap.Validate()
	if err != nil {
		return err
	}

	// chainalysis is optional
	if cfg.Chainalysis != nil {
		err := cfg.Chainalysis.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (cfg *CoinMarketCapConfig) Validate() error {
	if cfg == nil {
		return fmt.Errorf("coinmarketcap configuration is required")
	}

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

func (cfg *ChainAnalysisConfig) Validate() error {
	if cfg.APIKey == "" {
		return fmt.Errorf("missing chainalysis api key")
	}

	if cfg.BaseURL == "" {
		return fmt.Errorf("missing chainalysis base url")
	}

	return nil
}
