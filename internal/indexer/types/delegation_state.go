package indexertypes

// Enum values for Delegation State
type DelegationState string

const (
	StatePending      DelegationState = "PENDING"
	StateVerified     DelegationState = "VERIFIED"
	StateActive       DelegationState = "ACTIVE"
	StateUnbonding    DelegationState = "UNBONDING"
	StateWithdrawable DelegationState = "WITHDRAWABLE"
	StateWithdrawn    DelegationState = "WITHDRAWN"
	StateSlashed      DelegationState = "SLASHED"
)

func (s DelegationState) String() string {
	return string(s)
}

type DelegationSubState string

const (
	SubStateTimelock               DelegationSubState = "TIMELOCK"
	SubStateEarlyUnbonding         DelegationSubState = "EARLY_UNBONDING"
	SubStateTimelockSlashing       DelegationSubState = "TIMELOCK_SLASHING"
	SubStateEarlyUnbondingSlashing DelegationSubState = "EARLY_UNBONDING_SLASHING"
)
