package config

import "errors"

// DelegationTransitionConfig represents the transition cutoff height for
// the phase-1 delegation to phase-2.
// A delegation can transition to phase-2 if either:
// 1. The delegation's BTC staking height is less than EligibleBeforeBtcHeight, or
// 2. The current BBN height is greater than AllowListExpirationHeight
// (allowing all delegations to transition)
type DelegationTransitionConfig struct {
	EligibleBeforeBtcHeight   uint64 `mapstructure:"eligible-before-btc-height"`
	AllowListExpirationHeight uint64 `mapstructure:"allow-list-expiration-height"`
}

func (cfg *DelegationTransitionConfig) Validate() error {
	if cfg.EligibleBeforeBtcHeight == 0 {
		return errors.New("before-btc-height cannot be 0")
	}

	if cfg.AllowListExpirationHeight == 0 {
		return errors.New("allow-list-expiration-height cannot be 0")
	}

	return nil
}
