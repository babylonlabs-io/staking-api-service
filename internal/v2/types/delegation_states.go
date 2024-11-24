package v2types

import (
	"fmt"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
)

// DelegationAPIState represents the flattened state for frontend consumption
type DelegationAPIState string

const (
	// Basic states
	StatePending  DelegationAPIState = "PENDING"
	StateVerified DelegationAPIState = "VERIFIED"
	StateActive   DelegationAPIState = "ACTIVE"

	// Unbonding states
	StateTimelockUnbonding DelegationAPIState = "TIMELOCK_UNBONDING"
	StateEarlyUnbonding    DelegationAPIState = "EARLY_UNBONDING"

	// Withdrawable states
	StateTimelockWithdrawable               DelegationAPIState = "TIMELOCK_WITHDRAWABLE"
	StateEarlyUnbondingWithdrawable         DelegationAPIState = "EARLY_UNBONDING_WITHDRAWABLE"
	StateTimelockSlashingWithdrawable       DelegationAPIState = "TIMELOCK_SLASHING_WITHDRAWABLE"
	StateEarlyUnbondingSlashingWithdrawable DelegationAPIState = "EARLY_UNBONDING_SLASHING_WITHDRAWABLE"

	// Withdrawn states
	StateTimelockWithdrawn               DelegationAPIState = "TIMELOCK_WITHDRAWN"
	StateEarlyUnbondingWithdrawn         DelegationAPIState = "EARLY_UNBONDING_WITHDRAWN"
	StateTimelockSlashingWithdrawn       DelegationAPIState = "TIMELOCK_SLASHING_WITHDRAWN"
	StateEarlyUnbondingSlashingWithdrawn DelegationAPIState = "EARLY_UNBONDING_SLASHING_WITHDRAWN"

	// Slashed states
	StateTimelockSlashed       DelegationAPIState = "TIMELOCK_SLASHED"
	StateEarlyUnbondingSlashed DelegationAPIState = "EARLY_UNBONDING_SLASHED"
)

// DeriveAPIState converts internal states to API states, returns error if combination is invalid
func DeriveDelegationAPIState(state indexertypes.DelegationState, subState indexertypes.DelegationSubState) (DelegationAPIState, error) {
	switch state {
	case indexertypes.StatePending:
		return StatePending, nil
	case indexertypes.StateVerified:
		return StateVerified, nil
	case indexertypes.StateActive:
		return StateActive, nil

	case indexertypes.StateUnbonding:
		switch subState {
		case indexertypes.SubStateTimelock:
			return StateTimelockUnbonding, nil
		case indexertypes.SubStateEarlyUnbonding:
			return StateEarlyUnbonding, nil
		}

	case indexertypes.StateWithdrawable:
		switch subState {
		case indexertypes.SubStateTimelock:
			return StateTimelockWithdrawable, nil
		case indexertypes.SubStateEarlyUnbonding:
			return StateEarlyUnbondingWithdrawable, nil
		case indexertypes.SubStateTimelockSlashing:
			return StateTimelockSlashingWithdrawable, nil
		case indexertypes.SubStateEarlyUnbondingSlashing:
			return StateEarlyUnbondingSlashingWithdrawable, nil
		}

	case indexertypes.StateWithdrawn:
		switch subState {
		case indexertypes.SubStateTimelock:
			return StateTimelockWithdrawn, nil
		case indexertypes.SubStateEarlyUnbonding:
			return StateEarlyUnbondingWithdrawn, nil
		case indexertypes.SubStateTimelockSlashing:
			return StateTimelockSlashingWithdrawn, nil
		case indexertypes.SubStateEarlyUnbondingSlashing:
			return StateEarlyUnbondingSlashingWithdrawn, nil
		}

	case indexertypes.StateSlashed:
		switch subState {
		case indexertypes.SubStateTimelockSlashing:
			return StateTimelockSlashed, nil
		case indexertypes.SubStateEarlyUnbondingSlashing:
			return StateEarlyUnbondingSlashed, nil
		}
	}

	return "", fmt.Errorf("invalid state/subState combination: state=%s, subState=%s", state, subState)
}
