package config

import (
	"fmt"
	"os"
	"strings"

	queue "github.com/babylonlabs-io/staking-queue-client/config"
	"github.com/spf13/viper"
)

type Config struct {
	Server                 *ServerConfig               `mapstructure:"server"`
	StakingDb              *DbConfig                   `mapstructure:"staking-db"`
	IndexerDb              *DbConfig                   `mapstructure:"indexer-db"`
	Queue                  *queue.QueueConfig          `mapstructure:"queue"`
	Metrics                *MetricsConfig              `mapstructure:"metrics"`
	Assets                 *AssetsConfig               `mapstructure:"assets"`
	DelegationTransition   *DelegationTransitionConfig `mapstructure:"delegation-transition"`
	ExternalAPIs           *ExternalAPIsConfig         `mapstructure:"external_apis"`
	TermsAcceptanceLogging *TermsAcceptanceConfig      `mapstructure:"terms_acceptance_logging"`
	AddressScreeningConfig *AddressScreeningConfig     `mapstructure:"address_screening"`
	BBN                    *BBNConfig                  `mapstructure:"bbn"`
	NetworkUpgrade         *NetworkUpgrade             `mapstructure:"network_upgrade,omitempty"`
	AllowList              *AllowList                  `mapstructure:"staking-expansion-allow-list"`
}

func (cfg *Config) Validate() error {
	type configValidator interface {
		Validate() error
	}

	configs := []configValidator{cfg.Server, cfg.StakingDb, cfg.IndexerDb, cfg.Metrics, cfg.Queue, cfg.NetworkUpgrade}
	for _, config := range configs {
		err := config.Validate()
		if err != nil {
			return err
		}
	}

	// Assets is optional
	if cfg.Assets != nil {
		if err := cfg.Assets.Validate(); err != nil {
			return err
		}
	}

	if cfg.DelegationTransition != nil {
		if err := cfg.DelegationTransition.Validate(); err != nil {
			return err
		}
	}

	// ExternalAPIs is optional
	if cfg.ExternalAPIs != nil {
		if err := cfg.ExternalAPIs.Validate(); err != nil {
			return err
		}
	}

	if cfg.BBN != nil {
		if err := cfg.BBN.Validate(); err != nil {
			return err
		}
	}

	if cfg.AllowList != nil {
		if err := cfg.AllowList.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// New returns a fully parsed Config object from a given file directory
func New(cfgFile string) (*Config, error) {
	_, err := os.Stat(cfgFile)
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile(cfgFile)

	viper.AutomaticEnv()
	/*
		Below code will replace nested fields in yml into `_` and any `-` into `__` when you try to override this config via env variable
		To give an example:
		1. `some.config.a` can be overridden by `SOME_CONFIG_A`
		2. `some.config-a` can be overridden by `SOME_CONFIG__A`
		This is to avoid using `-` in the environment variable as it's not supported in all os terminal/bash
		Note: vipner package use `.` as delimitter by default. Read more here: https://pkg.go.dev/github.com/spf13/viper#readme-accessing-nested-keys
	*/
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "__"))

	err = viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err = viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
