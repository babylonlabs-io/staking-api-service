package config

import "errors"

// DelegationTransitionConfig represents the transition cutoff height for the phase-1 delegation to phase-2.
// A delegation can transition to phase-2 if either:
// 1. The delegation's BTC staking height is less than BeforeBtcHeight, or
// 2. The current BBN height is greater than AfterBbnHeight (allowing all delegations to transition)
type DelegationTransitionConfig struct {
	BeforeBtcHeight uint64 `mapstructure:"before_btc_height"`
	AfterBbnHeight  uint64 `mapstructure:"after_bbn_height"`
}

func (cfg *DelegationTransitionConfig) Validate() error {
	if cfg.BeforeBtcHeight == 0 {
		return errors.New("before_btc_height cannot be 0")
	}

	if cfg.AfterBbnHeight == 0 {
		return errors.New("after_bbn_height cannot be 0")
	}

	return nil
}
