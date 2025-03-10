package environment

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
)

// Service provides environment-specific configuration values
type Service struct {
	environment Environment
}

// NewService creates a new environment service
func NewService() *Service {
	return &Service{
		environment: GetEnvironment(),
	}
}

// GetEnvironmentName returns the current environment name
func (s *Service) GetEnvironmentName() string {
	return s.environment.String()
}

// GetDelegationTransitionParams returns the delegation transition parameters for the current environment
func (s *Service) GetDelegationTransitionParams() DelegationTransitionParams {
	return GetDelegationTransitionParamsForEnvironment(s.environment)
}

// ApplyDelegationTransitionParams applies the environment-specific delegation transition parameters to the config
func (s *Service) ApplyDelegationTransitionParams(cfg *config.Config) {
	if cfg == nil {
		return
	}

	params := s.GetDelegationTransitionParams()

	// If the delegation transition config is not set, create it
	if cfg.DelegationTransition == nil {
		cfg.DelegationTransition = &config.DelegationTransitionConfig{}
	}

	// Override the values with environment-specific values
	cfg.DelegationTransition.EligibleBeforeBtcHeight = params.EligibleBeforeBtcHeight
	cfg.DelegationTransition.AllowListExpirationHeight = params.AllowListExpirationHeight
}
