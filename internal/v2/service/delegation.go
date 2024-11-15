package v2service

import (
	"context"
	"fmt"
	"net/http"

	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type DelegationStaking struct {
	StakingTxHashHex string `json:"staking_tx_hash_hex"`
	StakingTime      uint32 `json:"staking_time"`
	StakingAmount    uint64 `json:"staking_amount"`
	StartHeight      uint32 `json:"start_height,omitempty"`
	EndHeight        uint32 `json:"end_height,omitempty"`
}

type DelegationUnbonding struct {
	UnbondingTime uint32 `json:"unbonding_time"`
	UnbondingTx   string `json:"unbonding_tx"`
}

type StakerDelegationPublic struct {
	ParamsVersion             uint32                       `json:"params_version"`
	StakerBtcPkHex            string                       `json:"staker_btc_pk_hex"`
	FinalityProviderBtcPksHex []string                     `json:"finality_provider_btc_pks_hex"`
	DelegationStaking         DelegationStaking            `json:"delegation_staking"`
	DelegationUnbonding       DelegationUnbonding          `json:"delegation_unbonding"`
	State                     indexertypes.DelegationState `json:"state"`
}

func (s *V2Service) GetDelegation(ctx context.Context, stakingTxHashHex string) (*StakerDelegationPublic, *types.Error) {
	delegation, err := s.DbClients.IndexerDBClient.GetDelegation(ctx, stakingTxHashHex)
	if err != nil {
		return nil, types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "failed to get staker delegation")
	}

	delegationPublic := &StakerDelegationPublic{
		ParamsVersion:             delegation.ParamsVersion,
		FinalityProviderBtcPksHex: delegation.FinalityProviderBtcPksHex,
		StakerBtcPkHex:            delegation.StakerBtcPkHex,
		DelegationStaking: DelegationStaking{
			StakingTxHashHex: delegation.StakingTxHashHex,
			StakingTime:      delegation.StakingTime,
			StakingAmount:    delegation.StakingAmount,
			StartHeight:      delegation.StartHeight,
			EndHeight:        delegation.EndHeight,
		},
		DelegationUnbonding: DelegationUnbonding{
			UnbondingTime: delegation.UnbondingTime,
			UnbondingTx:   delegation.UnbondingTx,
		},
		State: delegation.State,
	}
	return delegationPublic, nil
}

func (s *V2Service) GetDelegations(ctx context.Context, stakerPkHex string, paginationKey string) ([]*StakerDelegationPublic, string, *types.Error) {
	resultMap, err := s.DbClients.IndexerDBClient.GetDelegations(ctx, stakerPkHex, paginationKey)
	if err != nil {
		return nil, "", types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			fmt.Sprintf("failed to get v2 staker delegations: %v", err),
		)
	}

	// Initialize result structure
	delegationsPublic := make([]*StakerDelegationPublic, 0, len(resultMap.Data))

	// Group delegations by state
	for _, delegation := range resultMap.Data {
		delegationPublic := &StakerDelegationPublic{
			ParamsVersion:             delegation.ParamsVersion,
			FinalityProviderBtcPksHex: delegation.FinalityProviderBtcPksHex,
			StakerBtcPkHex:            delegation.StakerBtcPkHex,
			DelegationStaking: DelegationStaking{
				StakingTxHashHex: delegation.StakingTxHashHex,
				StakingTime:      delegation.StakingTime,
				StakingAmount:    delegation.StakingAmount,
				StartHeight:      delegation.StartHeight,
				EndHeight:        delegation.EndHeight,
			},
			DelegationUnbonding: DelegationUnbonding{
				UnbondingTime: delegation.UnbondingTime,
				UnbondingTx:   delegation.UnbondingTx,
			},
			State: delegation.State,
		}
		delegationsPublic = append(delegationsPublic, delegationPublic)
	}

	return delegationsPublic, resultMap.PaginationToken, nil
}
