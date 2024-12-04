package utils

import "github.com/babylonlabs-io/staking-api-service/internal/shared/types"

// QualifiedStatesToUnbondingRequest returns the qualified exisitng states to transition to "unbonding_request"
func QualifiedStatesToUnbondingRequest() []types.DelegationState {
	return []types.DelegationState{types.Active}
}

// QualifiedStatesToUnbonding returns the qualified exisitng states to transition to "unbonding"
// The Active state is allowed to directly transition to Unbonding without the need of UnbondingRequested due to bootstrap usecase
func QualifiedStatesToUnbonding() []types.DelegationState {
	return []types.DelegationState{types.Active, types.UnbondingRequested}
}

// List of states to be ignored for unbonding as it means it's already been processed
func OutdatedStatesForUnbonding() []types.DelegationState {
	return []types.DelegationState{types.Unbonding, types.Unbonded, types.Withdrawn}
}

// QualifiedStatesToUnbonded returns the qualified exisitng states to transition to "unbonded"
func QualifiedStatesToUnbonded(unbondTxType types.StakingTxType) []types.DelegationState {
	switch unbondTxType {
	case types.ActiveTxType:
		return []types.DelegationState{types.Active}
	case types.UnbondingTxType:
		return []types.DelegationState{types.Unbonding}
	default:
		return nil
	}
}

// List of states to be ignored for unbonded(timelock expired) as it means it's already been processed
func OutdatedStatesForUnbonded() []types.DelegationState {
	return []types.DelegationState{types.Unbonded, types.Withdrawn}
}

// QualifiedStatesToWithdrawn returns the qualified exisitng states to transition to "withdrawn"
func QualifiedStatesToWithdraw() []types.DelegationState {
	return []types.DelegationState{types.Unbonded}
}

func OutdatedStatesForWithdraw() []types.DelegationState {
	return []types.DelegationState{types.Withdrawn}
}

// QualifiedStatesToTransitioned returns the qualified exisitng states to transition to "transitioned"
func QualifiedStatesToTransitioned() []types.DelegationState {
	return []types.DelegationState{types.Active, types.UnbondingRequested}
}
