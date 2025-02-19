package config

import (
	"fmt"
	"time"
)

type ExternalAPIsConfig struct {
	CoinMarketCap *CoinMarketCapConfig `mapstructure:"coinmarketcap"`
	ChainAnalysis *ChainAnalysisConfig `mapstructure:"chainanalysis"`
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
	if cfg.CoinMarketCap == nil && cfg.ChainAnalysis == nil {
		return fmt.Errorf("missing external api configuration")
	}

	if cfg.CoinMarketCap != nil {
		err := cfg.CoinMarketCap.Validate()
		if err != nil {
			return err
		}
	}

	if cfg.ChainAnalysis != nil {
		err := cfg.ChainAnalysis.Validate()
		if err != nil {
			return err
		}
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

	if cfg.CacheTTL <= 0 {
		return fmt.Errorf("invalid coinmarketcap cache ttl")
	}

	return nil
}

func (cfg *ChainAnalysisConfig) Validate() error {
	if cfg.APIKey == "" {
		return fmt.Errorf("missing chainanalysis api key")
	}

	if cfg.BaseURL == "" {
		return fmt.Errorf("missing chainanalysis base url")
	}

	return nil
}
