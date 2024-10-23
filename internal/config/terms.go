package config

type TermsAcceptanceConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

func (cfg *TermsAcceptanceConfig) Validate() error {
	// No validation needed for Enabled field as it can be either true or false

	return nil
}
