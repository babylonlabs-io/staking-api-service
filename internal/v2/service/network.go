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
	StakingStatus StakingStatusPublic `json:"staking_status"`
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

	isStakingOpen := false
	bbnHeight, dbError := s.Service.DbClients.IndexerDBClient.GetLastProcessedBbnHeight(ctx)
	if dbError != nil {
		log.Ctx(ctx).Error().Err(dbError).Msg("Failed to get last processed BBN height")
		return nil, types.NewInternalServiceError(err)
	}
	if bbnHeight >= s.Cfg.DelegationTransition.AllowListExpirationHeight {
		isStakingOpen = true
	}

	stakingStatus := StakingStatusPublic{
		IsStakingOpen: isStakingOpen,
	}

	return &NetworkInfoPublic{
		StakingStatus: stakingStatus,
		Params: ParamsPublic{
			Bbn: babylonParams,
			Btc: btcParams,
		},
	}, nil
}
