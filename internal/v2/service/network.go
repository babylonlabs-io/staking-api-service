package v2service

import (
	"cmp"
	"context"
	"slices"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

type StakingStatusPublic struct {
	IsStakingOpen bool `json:"is_staking_open"`
}

type POPUpgradePublic struct {
	Height  uint64 `json:"height"`
	Version uint64 `json:"version"`
}

type NetworkUpgradePublic struct {
	POP []POPUpgradePublic `json:"pop,omitempty"`
}

type NetworkInfoPublic struct {
	StakingStatus  StakingStatusPublic   `json:"staking_status,omitempty"`
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

	// Default to true if there is no rules for delegation transition
	status := true
	if s.cfg.DelegationTransition != nil {
		bbnHeight, dbError := s.dbClients.IndexerDBClient.GetLastProcessedBbnHeight(ctx)
		if dbError != nil {
			log.Ctx(ctx).Error().Err(dbError).Msg("Failed to get last processed BBN height")
			return nil, types.NewInternalServiceError(err)
		}
		status = bbnHeight >= s.cfg.DelegationTransition.AllowListExpirationHeight
	}

	// sort (asc) babylon params according to their version
	slices.SortFunc(babylonParams, func(a, b *indexertypes.BbnStakingParams) int {
		return cmp.Compare(a.Version, b.Version)
	})

	result := &NetworkInfoPublic{
		StakingStatus: StakingStatusPublic{
			IsStakingOpen: status,
		},
		Params: ParamsPublic{
			Bbn: babylonParams,
			Btc: btcParams,
		},
	}

	// Only include NetworkUpgrade if it exists and POP is configured
	if s.cfg.NetworkUpgrade != nil && len(s.cfg.NetworkUpgrade.POP) > 0 {
		popUpgrades := make([]POPUpgradePublic, len(s.cfg.NetworkUpgrade.POP))
		for i, pop := range s.cfg.NetworkUpgrade.POP {
			popUpgrades[i] = POPUpgradePublic{
				Height:  pop.Height,
				Version: pop.Version,
			}
		}

		result.NetworkUpgrade = &NetworkUpgradePublic{
			POP: popUpgrades,
		}
	}

	return result, nil
}
