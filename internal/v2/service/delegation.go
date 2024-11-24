package v2service

import (
	"context"
	"net/http"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

type DelegationStaking struct {
	StakingTxHashHex   string `json:"staking_tx_hash_hex"`
	StakingTxHex       string `json:"staking_tx_hex"`
	StakingTime        uint32 `json:"staking_time"`
	StakingAmount      uint64 `json:"staking_amount"`
	StartHeight        uint32 `json:"start_height,omitempty"`
	EndHeight          uint32 `json:"end_height,omitempty"`
	BbnInceptionHeight int64  `json:"bbn_inception_height"`
	BbnInceptionTime   int64  `json:"bbn_inception_time"`
}

type CovenantSignature struct {
	CovenantBtcPkHex string `json:"covenant_btc_pk_hex"`
	SignatureHex     string `json:"signature_hex"`
}

type DelegationUnbonding struct {
	UnbondingTime               uint32              `json:"unbonding_time"`
	UnbondingTx                 string              `json:"unbonding_tx"`
	CovenantUnbondingSignatures []CovenantSignature `json:"covenant_unbonding_signatures"`
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
		if db.IsNotFoundError(err) {
			log.Ctx(ctx).Warn().Err(err).Str("stakingTxHashHex", stakingTxHashHex).Msg("Staking delegation not found")
			return nil, types.NewErrorWithMsg(http.StatusNotFound, types.NotFound, "staking delegation not found, please retry")
		}
		return nil, types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "failed to get staker delegation")
	}

	delegationPublic := &StakerDelegationPublic{
		ParamsVersion:             delegation.ParamsVersion,
		FinalityProviderBtcPksHex: delegation.FinalityProviderBtcPksHex,
		StakerBtcPkHex:            delegation.StakerBtcPkHex,
		DelegationStaking: DelegationStaking{
			StakingTxHashHex:   delegation.StakingTxHashHex,
			StakingTxHex:       delegation.StakingTxHex,
			StakingTime:        delegation.StakingTime,
			StakingAmount:      delegation.StakingAmount,
			StartHeight:        delegation.StartHeight,
			EndHeight:          delegation.EndHeight,
			BbnInceptionHeight: delegation.BTCDelegationCreatedBbnBlock.Height,
			BbnInceptionTime:   delegation.BTCDelegationCreatedBbnBlock.Timestamp,
		},
		DelegationUnbonding: DelegationUnbonding{
			UnbondingTime: delegation.UnbondingTime,
			UnbondingTx:   delegation.UnbondingTx,
			CovenantUnbondingSignatures: getUnbondingSignatures(
				delegation.CovenantUnbondingSignatures,
			),
		},
		State: delegation.State,
	}
	return delegationPublic, nil
}

func (s *V2Service) GetDelegations(ctx context.Context, stakerPkHex string, paginationKey string) ([]*StakerDelegationPublic, string, *types.Error) {
	resultMap, err := s.DbClients.IndexerDBClient.GetDelegations(ctx, stakerPkHex, paginationKey)
	if err != nil {
		if db.IsNotFoundError(err) {
			log.Ctx(ctx).Warn().Err(err).Str("stakingTxHashHex", stakerPkHex).Msg("Staking delegations not found")
			return nil, "", types.NewErrorWithMsg(http.StatusNotFound, types.NotFound, "staking delegation not found, please retry")
		}
		return nil, "", types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "failed to get staker delegations")
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
				StakingTxHex:     delegation.StakingTxHex,
				StakingTime:      delegation.StakingTime,
				StakingAmount:    delegation.StakingAmount,
				StartHeight:      delegation.StartHeight,
				EndHeight:        delegation.EndHeight,
			},
			DelegationUnbonding: DelegationUnbonding{
				UnbondingTime: delegation.UnbondingTime,
				UnbondingTx:   delegation.UnbondingTx,
				CovenantUnbondingSignatures: getUnbondingSignatures(
					delegation.CovenantUnbondingSignatures,
				),
			},
			State: delegation.State,
		}
		delegationsPublic = append(delegationsPublic, delegationPublic)
	}

	return delegationsPublic, resultMap.PaginationToken, nil
}

func getUnbondingSignatures(covenantSignatures []indexerdbmodel.CovenantSignature) []CovenantSignature {
	covenantSignaturesPublic := make([]CovenantSignature, 0, len(covenantSignatures))
	for _, covenantSignature := range covenantSignatures {
		covenantSignaturesPublic = append(covenantSignaturesPublic, CovenantSignature{CovenantBtcPkHex: covenantSignature.CovenantBtcPkHex, SignatureHex: covenantSignature.SignatureHex})
	}
	return covenantSignaturesPublic
}
