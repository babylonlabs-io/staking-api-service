package v2service

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
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
		bbnHeight, dbError := s.Service.DbClients.IndexerDBClient.GetLastProcessedBbnHeight(ctx)
		if dbError != nil {
			log.Ctx(ctx).Error().Err(dbError).Msg("Failed to get last processed BBN height")
			return nil, types.NewInternalServiceError(err)
		}
		status = bbnHeight >= s.Cfg.DelegationTransition.AllowListExpirationHeight
	}

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
