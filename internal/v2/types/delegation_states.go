package v2types

import (
	"fmt"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
)

// DelegationState represents the flattened state for frontend consumption
type DelegationState string

const (
	// Basic states
	StatePending  DelegationState = "PENDING"
	StateVerified DelegationState = "VERIFIED"
	StateActive   DelegationState = "ACTIVE"
	StateSlashed  DelegationState = "SLASHED"

	// Unbonding states
	StateTimelockUnbonding DelegationState = "TIMELOCK_UNBONDING"
	StateEarlyUnbonding    DelegationState = "EARLY_UNBONDING"

	// Withdrawable states
	StateTimelockWithdrawable               DelegationState = "TIMELOCK_WITHDRAWABLE"
	StateEarlyUnbondingWithdrawable         DelegationState = "EARLY_UNBONDING_WITHDRAWABLE"
	StateTimelockSlashingWithdrawable       DelegationState = "TIMELOCK_SLASHING_WITHDRAWABLE"
	StateEarlyUnbondingSlashingWithdrawable DelegationState = "EARLY_UNBONDING_SLASHING_WITHDRAWABLE"

	// Withdrawn states
	StateTimelockWithdrawn               DelegationState = "TIMELOCK_WITHDRAWN"
	StateEarlyUnbondingWithdrawn         DelegationState = "EARLY_UNBONDING_WITHDRAWN"
	StateTimelockSlashingWithdrawn       DelegationState = "TIMELOCK_SLASHING_WITHDRAWN"
	StateEarlyUnbondingSlashingWithdrawn DelegationState = "EARLY_UNBONDING_SLASHING_WITHDRAWN"

	// Expanded states
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
