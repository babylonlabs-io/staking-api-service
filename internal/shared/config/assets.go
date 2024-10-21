package config

import "errors"

type AssetsConfig struct {
	MaxUTXOs uint32          `mapstructure:"max_utxos"`
	Ordinals *OrdinalsConfig `mapstructure:"ordinals"`
}

func (cfg *AssetsConfig) Validate() error {
	if err := cfg.Ordinals.Validate(); err != nil {
		return err
	}

	if cfg.MaxUTXOs <= 0 {
		return errors.New("max_utxos cannot be smaller or equal to 0")
	}

	return nil
}
