package config

import (
	"fmt"
	"net"
)

// MetricsConfig defines the server's metric configuration
type MetricsConfig struct {
	// IP of the prometheus server
	Host string `mapstructure:"host"`
	// Port of the prometheus server
	Port int `mapstructure:"port"`
}

func (cfg *MetricsConfig) Validate() error {
	if cfg.Port < 1024 || cfg.Port > 65535 {
		return fmt.Errorf("metrics server port must be between 1024 and 65535 (inclusive)")
	}

	ip := net.ParseIP(cfg.Host)
	if ip == nil {
		return fmt.Errorf("invalid metrics server host: %v", cfg.Host)
	}

	return nil
}

func (cfg *MetricsConfig) GetMetricsPort() int {
	return cfg.Port
}

func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Host: "0.0.0.0",
		Port: 2112,
	}
}
