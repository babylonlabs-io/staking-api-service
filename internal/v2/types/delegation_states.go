package v2types

import (
	"fmt"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
)

// DelegationState represents the current state of a BTC delegation in the staking lifecycle.
//
// States are organized into categories:
// - Setup: PENDING, VERIFIED, ACTIVE
// - Unbonding: TIMELOCK_UNBONDING, EARLY_UNBONDING
// - Withdrawable: TIMELOCK_WITHDRAWABLE, EARLY_UNBONDING_WITHDRAWABLE, TIMELOCK_SLASHING_WITHDRAWABLE, EARLY_UNBONDING_SLASHING_WITHDRAWABLE
// - Withdrawn: TIMELOCK_WITHDRAWN, EARLY_UNBONDING_WITHDRAWN, TIMELOCK_SLASHING_WITHDRAWN, EARLY_UNBONDING_SLASHING_WITHDRAWN
// - Special: SLASHED, EXPANDED
//
// For detailed explanations of each status, see https://github.com/babylonlabs-io/staking-api-service/blob/main/docs/delegation-statuses.md
type DelegationState string

const (
	// Setup States

	// StatePending - Delegation created on Babylon, awaiting covenant committee signatures
	StatePending DelegationState = "PENDING"
	// StateVerified - Has covenant signatures, waiting for BTC confirmation and inclusion proof
	StateVerified DelegationState = "VERIFIED"
	// StateActive - Fully active delegation participating in staking, contributing voting power
	StateActive DelegationState = "ACTIVE"

	// Unbonding States

	// StateTimelockUnbonding - Natural expiration: delegation in unbonding period after reaching end height
	StateTimelockUnbonding DelegationState = "TIMELOCK_UNBONDING"
	// StateEarlyUnbonding - Early unbonding requested: unbonding transaction submitted, in unbonding period
	StateEarlyUnbonding DelegationState = "EARLY_UNBONDING"

	// Withdrawable States

	// StateTimelockWithdrawable - Natural expiration complete: funds ready to withdraw via timelock path
	StateTimelockWithdrawable DelegationState = "TIMELOCK_WITHDRAWABLE"
	// StateEarlyUnbondingWithdrawable - Early unbonding complete: funds ready to withdraw via unbonding tx
	StateEarlyUnbondingWithdrawable DelegationState = "EARLY_UNBONDING_WITHDRAWABLE"
	// StateTimelockSlashingWithdrawable - Staking output slashed: remaining funds ready to withdraw after slashing timelock
	StateTimelockSlashingWithdrawable DelegationState = "TIMELOCK_SLASHING_WITHDRAWABLE"
	// StateEarlyUnbondingSlashingWithdrawable - Unbonding output slashed: remaining funds ready to withdraw after slashing timelock
	StateEarlyUnbondingSlashingWithdrawable DelegationState = "EARLY_UNBONDING_SLASHING_WITHDRAWABLE"

	// Withdrawn States (Terminal)

	// StateTimelockWithdrawn - Funds withdrawn via natural expiration path (terminal state)
	StateTimelockWithdrawn DelegationState = "TIMELOCK_WITHDRAWN"
	// StateEarlyUnbondingWithdrawn - Funds withdrawn via early unbonding path (terminal state)
	StateEarlyUnbondingWithdrawn DelegationState = "EARLY_UNBONDING_WITHDRAWN"
	// StateTimelockSlashingWithdrawn - Remaining funds withdrawn after staking output slashed (terminal state)
	StateTimelockSlashingWithdrawn DelegationState = "TIMELOCK_SLASHING_WITHDRAWN"
	// StateEarlyUnbondingSlashingWithdrawn - Remaining funds withdrawn after unbonding output slashed (terminal state)
	StateEarlyUnbondingSlashingWithdrawn DelegationState = "EARLY_UNBONDING_SLASHING_WITHDRAWN"

	// Special States

	// StateSlashed - Delegation slashed due to finality provider misbehavior, waiting for slashing timelock
	StateSlashed DelegationState = "SLASHED"
	// StateExpanded - Delegation extended to a new delegation with extended timelock (terminal state). Currently supports extending timelock only, not increasing stake amount.
	StateExpanded DelegationState = "EXPANDED"
)

// MapDelegationState consumes internal indexer states and maps them to the frontend-facing states
func MapDelegationState(state indexertypes.DelegationState, subState indexertypes.DelegationSubState) (DelegationState, error) {
	switch state {
	case indexertypes.StatePending:
		return StatePending, nil
	case indexertypes.StateVerified:
		return StateVerified, nil
	case indexertypes.StateActive:
		return StateActive, nil
	case indexertypes.StateSlashed:
		return StateSlashed, nil
	case indexertypes.StateExpanded:
		return StateExpanded, nil
	case indexertypes.StateUnbonding:
		return mapUnbondingState(subState)
	case indexertypes.StateWithdrawable:
		return mapWithdrawableState(subState)
	case indexertypes.StateWithdrawn:
		return mapWithdrawnState(subState)
	}

	return "", fmt.Errorf("invalid state/subState combination: state=%s, subState=%s", state, subState)
}

// mapUnbondingState maps unbonding states based on subState
func mapUnbondingState(subState indexertypes.DelegationSubState) (DelegationState, error) {
	switch subState {
	case indexertypes.SubStateTimelock:
		return StateTimelockUnbonding, nil
	case indexertypes.SubStateEarlyUnbonding:
		return StateEarlyUnbonding, nil
	default:
		return "", fmt.Errorf("invalid subState for StateUnbonding: %s", subState)
	}
}

// mapWithdrawableState maps withdrawable states based on subState
func mapWithdrawableState(subState indexertypes.DelegationSubState) (DelegationState, error) {
	switch subState {
	case indexertypes.SubStateTimelock:
		return StateTimelockWithdrawable, nil
	case indexertypes.SubStateEarlyUnbonding:
		return StateEarlyUnbondingWithdrawable, nil
	case indexertypes.SubStateTimelockSlashing:
		return StateTimelockSlashingWithdrawable, nil
	case indexertypes.SubStateEarlyUnbondingSlashing:
		return StateEarlyUnbondingSlashingWithdrawable, nil
	default:
		return "", fmt.Errorf("invalid subState for StateWithdrawable: %s", subState)
	}
}

// mapWithdrawnState maps withdrawn states based on subState
func mapWithdrawnState(subState indexertypes.DelegationSubState) (DelegationState, error) {
	switch subState {
	case indexertypes.SubStateTimelock:
		return StateTimelockWithdrawn, nil
	case indexertypes.SubStateEarlyUnbonding:
		return StateEarlyUnbondingWithdrawn, nil
	case indexertypes.SubStateTimelockSlashing:
		return StateTimelockSlashingWithdrawn, nil
	case indexertypes.SubStateEarlyUnbondingSlashing:
		return StateEarlyUnbondingSlashingWithdrawn, nil
	default:
		return "", fmt.Errorf("invalid subState for StateWithdrawn: %s", subState)
	}
}
