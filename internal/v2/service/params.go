package v2service

import (
	"context"
	"net/http"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type GlobalParamsPublic struct {
	Bbn []*indexertypes.BbnStakingParams    `json:"bbn"`
	Btc []*indexertypes.BtcCheckpointParams `json:"btc"`
}

func (s *V2Service) GetGlobalParams(ctx context.Context) (*GlobalParamsPublic, *types.Error) {
	babylonParams, err := s.getBbnStakingParams(ctx)
	if err != nil {
		return nil, err
	}

	btcParams, err := s.getBtcCheckpointParams(ctx)
	if err != nil {
		return nil, err
	}
	return &GlobalParamsPublic{
		Bbn: babylonParams,
		Btc: btcParams,
	}, nil
}

func (s *V2Service) getBbnStakingParams(ctx context.Context) ([]*indexertypes.BbnStakingParams, *types.Error) {
	params, err := s.DbClients.IndexerDBClient.GetBbnStakingParams(ctx)
	if err != nil {
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError, types.InternalServiceError,
			"failed to get babylon params",
		)
	}

	return params, nil
}

func (s *V2Service) getBtcCheckpointParams(ctx context.Context) ([]*indexertypes.BtcCheckpointParams, *types.Error) {
	params, err := s.DbClients.IndexerDBClient.GetBtcCheckpointParams(ctx)
	if err != nil {
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError, types.InternalServiceError,
			"failed to get btc params",
		)
	}
	return params, nil
}
