package types

import "fmt"

type DelegationState string

const (
	Active             DelegationState = "active"
	UnbondingRequested DelegationState = "unbonding_requested"
	Unbonding          DelegationState = "unbonding"
	Unbonded           DelegationState = "unbonded"
	Withdrawable       DelegationState = "withdrawable"
	Withdrawn          DelegationState = "withdrawn"
	Transitioned       DelegationState = "transitioned"
	Slashed            DelegationState = "slashed"
)

func (s DelegationState) ToString() string {
	return string(s)
}

func FromStringToDelegationState(s string) (DelegationState, error) {
	switch s {
	case "active":
		return Active, nil
	case "unbonding_requested":
		return UnbondingRequested, nil
	case "unbonding":
		return Unbonding, nil
	case "unbonded":
		return Unbonded, nil
	case "withdrawable":
		return Withdrawable, nil
	case "withdrawn":
		return Withdrawn, nil
	case "transitioned":
		return Transitioned, nil
	default:
		return "", fmt.Errorf("invalid delegation state: %s", s)
	}
}
