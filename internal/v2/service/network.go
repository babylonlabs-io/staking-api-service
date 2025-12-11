package v2service

import (
	"cmp"
	"context"
	"slices"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type POPUpgradePublic struct {
	Height  uint64 `json:"height"`
	Version uint64 `json:"version"`
}

type NetworkUpgradePublic struct {
	POP []POPUpgradePublic `json:"pop,omitempty"`
}

type NetworkInfoPublic struct {
	Params         ParamsPublic          `json:"params"`
	NetworkUpgrade *NetworkUpgradePublic `json:"network_upgrade,omitempty"`
}

func (s *V2Service) GetNetworkInfo(ctx context.Context) (*NetworkInfoPublic, *types.Error) {
	// TODO: The get params call should be a single call
	babylonParams, err := s.getBbnStakingParams(ctx)
	if err != nil {
		return nil, err
	}

	btcParams, err := s.getBtcCheckpointParams(ctx)
	if err != nil {
		return nil, err
	}

	// sort (asc) babylon params according to their version
	slices.SortFunc(babylonParams, func(a, b *indexertypes.BbnStakingParams) int {
		return cmp.Compare(a.Version, b.Version)
	})

	result := &NetworkInfoPublic{
		Params: ParamsPublic{
			Bbn: babylonParams,
			Btc: btcParams,
		},
	}

	// Only include NetworkUpgrade if it exists and POP is configured
	if networkUpgrade := s.cfg.NetworkUpgrade; networkUpgrade != nil {
		if len(networkUpgrade.POP) > 0 {
			popUpgrades := make([]POPUpgradePublic, len(networkUpgrade.POP))
			for i, pop := range networkUpgrade.POP {
				popUpgrades[i] = POPUpgradePublic{
					Height:  pop.Height,
					Version: pop.Version,
				}
			}

			result.NetworkUpgrade = &NetworkUpgradePublic{
				POP: popUpgrades,
			}
		}
	}

	return result, nil
}
