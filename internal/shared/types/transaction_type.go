package types

import "fmt"

type StakingTxType string

const (
	ActiveTxType    StakingTxType = "active"
	UnbondingTxType StakingTxType = "unbonding"
)

func (s StakingTxType) ToString() string {
	return string(s)
}

func StakingTxTypeFromString(s string) (StakingTxType, error) {
	switch s {
	case ActiveTxType.ToString():
		return ActiveTxType, nil
	case UnbondingTxType.ToString():
		return UnbondingTxType, nil
	default:
		return "", fmt.Errorf("unknown staking tx type: %s", s)
	}
}
