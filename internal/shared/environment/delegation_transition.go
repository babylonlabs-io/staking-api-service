package environment

// DelegationTransitionParams contains the environment-specific delegation transition parameters
type DelegationTransitionParams struct {
	EligibleBeforeBtcHeight   uint64
	AllowListExpirationHeight uint64
}

// GetDelegationTransitionParams returns the delegation transition parameters for the current environment
func GetDelegationTransitionParams() DelegationTransitionParams {
	env := GetEnvironment()
	return GetDelegationTransitionParamsForEnvironment(env)
}

// GetDelegationTransitionParamsForEnvironment returns the delegation transition parameters for the given environment
func GetDelegationTransitionParamsForEnvironment(env Environment) DelegationTransitionParams {
	switch env {
	case MockMainnet:
		return DelegationTransitionParams{
			EligibleBeforeBtcHeight:   882251,
			AllowListExpirationHeight: 8844,
		}
	case Phase2Devnet:
		return DelegationTransitionParams{
			EligibleBeforeBtcHeight:   227490,
			AllowListExpirationHeight: 1440,
		}
	case Phase2Testnet:
		return DelegationTransitionParams{
			EligibleBeforeBtcHeight:   198663,
			AllowListExpirationHeight: 26124,
		}
	case Phase2PrivateMainnet:
		return DelegationTransitionParams{
			EligibleBeforeBtcHeight:   882251,
			AllowListExpirationHeight: 8844,
		}
	default:
		// Default to some safe values
		return DelegationTransitionParams{
			EligibleBeforeBtcHeight:   10,
			AllowListExpirationHeight: 10,
		}
	}
}
