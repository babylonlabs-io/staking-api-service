package v2types

// List of states to be ignored for unbonding as it means it's already been processed
func OutdatedStatesForUnbonding() []DelegationState {
	return []DelegationState{
		StateTimelockUnbonding,
		StateEarlyUnbonding,
		StateTimelockWithdrawable,
		StateEarlyUnbondingWithdrawable,
		StateTimelockSlashingWithdrawable,
		StateEarlyUnbondingSlashingWithdrawable,
		StateTimelockWithdrawn,
		StateEarlyUnbondingWithdrawn,
		StateTimelockSlashingWithdrawn,
		StateEarlyUnbondingSlashingWithdrawn,
		StateTimelockSlashed,
		StateEarlyUnbondingSlashed,
	}
}
