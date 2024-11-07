package v2service

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils/datagen"
)

type StakerDelegationPublic struct {
	StakingTxHashHex          string                       `json:"staking_tx_hash_hex"`
	ParamsVersion             string                       `json:"params_version"`
	FinalityProviderBtcPksHex []string                     `json:"finality_provider_btc_pks_hex"`
	StakerBtcPkHex            string                       `json:"staker_btc_pk_hex"`
	StakingTime               string                       `json:"staking_time"`
	StakingAmount             string                       `json:"staking_amount"`
	UnbondingTime             string                       `json:"unbonding_time"`
	UnbondingTx               string                       `json:"unbonding_tx"`
	State                     indexertypes.DelegationState `json:"state"`
	StartHeight               uint32                       `json:"start_height"`
	EndHeight                 uint32                       `json:"end_height"`
}

type StakerStatsPublic struct {
	StakerPKHex       string `json:"staker_pk_hex"`
	ActiveTVL         int64  `json:"active_tvl"`
	TotalTVL          int64  `json:"total_tvl"`
	ActiveDelegations int64  `json:"active_delegations"`
	TotalDelegations  int64  `json:"total_delegations"`
}

func (s *V2Service) GetStakerDelegations(ctx context.Context, stakingTxHashHex string, paginationKey string) ([]*StakerDelegationPublic, string, *types.Error) {
	resultMap, err := s.DbClients.IndexerDBClient.GetStakerDelegations(ctx, stakingTxHashHex, paginationKey)
	if err != nil {
		return nil, "", types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "failed to get staker delegations")
	}

	delegationsPublic := make([]*StakerDelegationPublic, 0, len(resultMap.Data))
	for _, delegation := range resultMap.Data {
		delegationsPublic = append(delegationsPublic, &StakerDelegationPublic{
			StakingTxHashHex:          delegation.StakingTxHashHex,
			ParamsVersion:             delegation.ParamsVersion,
			FinalityProviderBtcPksHex: delegation.FinalityProviderBtcPksHex,
			StakerBtcPkHex:            delegation.StakerBtcPkHex,
			StakingTime:               delegation.StakingTime,
			StakingAmount:             delegation.StakingAmount,
			UnbondingTime:             delegation.UnbondingTime,
			UnbondingTx:               delegation.UnbondingTx,
			State:                     delegation.State,
			StartHeight:               delegation.StartHeight,
			EndHeight:                 delegation.EndHeight,
		})
	}
	return delegationsPublic, resultMap.PaginationToken, nil
}

func (s *V2Service) GetStakerStats(ctx context.Context, stakerPKHex string) (StakerStatsPublic, *types.Error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	stakerStats := StakerStatsPublic{
		StakerPKHex:       stakerPKHex,
		ActiveTVL:         int64(datagen.RandomPositiveInt(r, 1000000)),
		TotalTVL:          int64(datagen.RandomPositiveInt(r, 1000000)),
		ActiveDelegations: int64(datagen.RandomPositiveInt(r, 100)),
		TotalDelegations:  int64(datagen.RandomPositiveInt(r, 100)),
	}
	return stakerStats, nil
}
