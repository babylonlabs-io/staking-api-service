package v1service

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type VersionedGlobalParamsPublic struct {
	Version           uint64   `json:"version"`
	ActivationHeight  uint64   `json:"activation_height"`
	StakingCap        uint64   `json:"staking_cap"`
	CapHeight         uint64   `json:"cap_height"`
	Tag               string   `json:"tag"`
	CovenantPks       []string `json:"covenant_pks"`
	CovenantQuorum    uint64   `json:"covenant_quorum"`
	UnbondingTime     uint64   `json:"unbonding_time"`
	UnbondingFee      uint64   `json:"unbonding_fee"`
	MaxStakingAmount  uint64   `json:"max_staking_amount"`
	MinStakingAmount  uint64   `json:"min_staking_amount"`
	MaxStakingTime    uint64   `json:"max_staking_time"`
	MinStakingTime    uint64   `json:"min_staking_time"`
	ConfirmationDepth uint64   `json:"confirmation_depth"`
}

type GlobalParamsPublic struct {
	Versions []VersionedGlobalParamsPublic `json:"versions"`
}

func (s *V1Service) GetGlobalParamsPublic() *GlobalParamsPublic {
	var versionedParams []VersionedGlobalParamsPublic
	for _, version := range s.Service.Params.Versions {
		versionedParams = append(versionedParams, VersionedGlobalParamsPublic{
			Version:           version.Version,
			ActivationHeight:  version.ActivationHeight,
			StakingCap:        version.StakingCap,
			CapHeight:         version.CapHeight,
			Tag:               version.Tag,
			CovenantPks:       version.CovenantPks,
			CovenantQuorum:    version.CovenantQuorum,
			UnbondingTime:     version.UnbondingTime,
			UnbondingFee:      version.UnbondingFee,
			MaxStakingAmount:  version.MaxStakingAmount,
			MinStakingAmount:  version.MinStakingAmount,
			MaxStakingTime:    version.MaxStakingTime,
			MinStakingTime:    version.MinStakingTime,
			ConfirmationDepth: version.ConfirmationDepth,
		})
	}
	return &GlobalParamsPublic{
		Versions: versionedParams,
	}
}

// GetVersionedGlobalParamsByHeight returns the versioned global params
// for a particular bitcoin height
func (s *V1Service) GetVersionedGlobalParamsByHeight(height uint64) *types.VersionedGlobalParams {
	// Iterate the list in reverse (i.e. decreasing ActivationHeight)
	// and identify the first element that has an activation height below
	// the specified BTC height.
	for i := len(s.Service.Params.Versions) - 1; i >= 0; i-- {
		paramsVersion := s.Service.Params.Versions[i]
		if paramsVersion.ActivationHeight <= height {
			return paramsVersion
		}
	}
	return nil
}
