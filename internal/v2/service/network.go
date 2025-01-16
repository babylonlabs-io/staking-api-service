package v2service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
	"slices"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"cmp"
)

type StakingStatusPublic struct {
	IsStakingOpen bool `json:"is_staking_open"`
}

type NetworkInfoPublic struct {
	StakingStatus StakingStatusPublic `json:"staking_status,omitempty"`
	Params        ParamsPublic        `json:"params"`
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
	if s.Cfg.DelegationTransition != nil {
		bbnHeight, dbError := s.DbClients.IndexerDBClient.GetLastProcessedBbnHeight(ctx)
		if dbError != nil {
			log.Ctx(ctx).Error().Err(dbError).Msg("Failed to get last processed BBN height")
			return nil, types.NewInternalServiceError(err)
		}
		status = bbnHeight >= s.Cfg.DelegationTransition.AllowListExpirationHeight
	}

	// sort (asc) babylon params according to their version
	slices.SortFunc(babylonParams, func(a, b *indexertypes.BbnStakingParams) int {
		return cmp.Compare(a.Version, b.Version)
	})

	return &NetworkInfoPublic{
		StakingStatus: StakingStatusPublic{
			IsStakingOpen: status,
		},
		Params: ParamsPublic{
			Bbn: babylonParams,
			Btc: btcParams,
		},
	}, nil
}
