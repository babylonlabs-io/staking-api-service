package environment

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

// Environment represents the deployment environment
type Environment string

const (
	// Environments
	MockMainnet          Environment = "mock-mainnet"
	Phase2Devnet         Environment = "phase-2-devnet"
	Phase2Testnet        Environment = "phase-2-testnet"
	Phase2PrivateMainnet Environment = "phase-2-private-mainnet"
	Unknown              Environment = "unknown"

	// Environment variable name
	envVarName = "BABYLON_ENVIRONMENT"
)

// GetEnvironment determines the environment based on the environment variable
func GetEnvironment() Environment {
	envStr := os.Getenv(envVarName)
	if envStr == "" {
		log.Warn().Msg("BABYLON_ENVIRONMENT not set, using unknown environment")
		return Unknown
	}

	return environmentFromString(envStr)
}

// environmentFromString converts a string to an Environment
func environmentFromString(envStr string) Environment {
	envStr = strings.ToLower(envStr)

	switch envStr {
	case "mock-mainnet":
		return MockMainnet
	case "phase-2-devnet":
		return Phase2Devnet
	case "phase-2-testnet":
		return Phase2Testnet
	case "phase-2-private-mainnet":
		return Phase2PrivateMainnet
	default:
		log.Warn().Str("environment", envStr).Msg("Unknown environment, using unknown environment")
		return Unknown
	}
}

// String returns the string representation of the environment
func (e Environment) String() string {
	return string(e)
}

// IsValid checks if the environment is valid
func (e Environment) IsValid() bool {
	switch e {
	case MockMainnet, Phase2Devnet, Phase2Testnet, Phase2PrivateMainnet:
		return true
	default:
		return false
	}
}
