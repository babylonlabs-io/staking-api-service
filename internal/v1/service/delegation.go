package v1service

import (
	"context"
	"net/http"

	indexerdbclient "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/client"
	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	v1dbclient "github.com/babylonlabs-io/staking-api-service/internal/v1/db/client"
	v1model "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
	"github.com/rs/zerolog/log"
)

type TransactionPublic struct {
	TxHex          string `json:"tx_hex"`
	OutputIndex    uint64 `json:"output_index"`
	StartTimestamp string `json:"start_timestamp"`
	StartHeight    uint64 `json:"start_height"`
	TimeLock       uint64 `json:"timelock"`
}

type DelegationPublic struct {
	StakingTxHashHex        string             `json:"staking_tx_hash_hex"`
	StakerPkHex             string             `json:"staker_pk_hex"`
	FinalityProviderPkHex   string             `json:"finality_provider_pk_hex"`
	State                   string             `json:"state"`
	StakingValue            uint64             `json:"staking_value"`
	StakingTx               *TransactionPublic `json:"staking_tx"`
	UnbondingTx             *TransactionPublic `json:"unbonding_tx,omitempty"`
	IsOverflow              bool               `json:"is_overflow"`
	IsEligibleForTransition bool               `json:"is_eligible_for_transition"`
	IsSlashed               bool               `json:"is_slashed"`
}

func (s *V1Service) DelegationsByStakerPk(
	ctx context.Context, stakerPk string,
	states []types.DelegationState, pageToken string,
) ([]*DelegationPublic, string, *types.Error) {
	filter := &v1dbclient.DelegationFilter{}
	if len(states) > 0 {
		filter = &v1dbclient.DelegationFilter{
			States: states,
		}
	}

	resultMap, err := s.Service.DbClients.V1DBClient.FindDelegationsByStakerPk(ctx, stakerPk, filter, pageToken)
	if err != nil {
		if db.IsInvalidPaginationTokenError(err) {
			log.Ctx(ctx).Warn().Err(err).Msg("Invalid pagination token when fetching delegations by staker pk")
			return nil, "", types.NewError(http.StatusBadRequest, types.BadRequest, err)
		}
		log.Ctx(ctx).Error().Err(err).Msg("Failed to find delegations by staker pk")
		return nil, "", types.NewInternalServiceError(err)
	}
	var delegations []*DelegationPublic = make([]*DelegationPublic, 0, len(resultMap.Data))
	bbnHeight, err := s.Service.DbClients.IndexerDBClient.GetLastProcessedBbnHeight(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to get last processed BBN height")
		return nil, "", types.NewInternalServiceError(err)
	}

	// Get list of all finality providers in phase-2
	transitionedFps, err := s.Service.DbClients.IndexerDBClient.GetFinalityProviders(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to get finality providers")
		return nil, "", types.NewInternalServiceError(err)
	}

	for _, d := range resultMap.Data {
		delegations = append(delegations, s.FromDelegationDocument(&d, bbnHeight, transitionedFps))
	}
	return delegations, resultMap.PaginationToken, nil
}

// SaveActiveStakingDelegation saves the active staking delegation to the database.
func (s *V1Service) SaveActiveStakingDelegation(
	ctx context.Context, txHashHex, stakerPkHex, finalityProviderPkHex string,
	value, startHeight uint64, stakingTimestamp int64, timeLock, stakingOutputIndex uint64,
	stakingTxHex string, isOverflow bool,
) *types.Error {
	err := s.Service.DbClients.V1DBClient.SaveActiveStakingDelegation(
		ctx, txHashHex, stakerPkHex, finalityProviderPkHex, stakingTxHex,
		value, startHeight, timeLock, stakingOutputIndex, stakingTimestamp, isOverflow,
	)
	if err != nil {
		if ok := db.IsDuplicateKeyError(err); ok {
			log.Ctx(ctx).Warn().Err(err).Msg("Skip the active staking event as it already exists in the database")
			return nil
		}
		log.Ctx(ctx).Error().Err(err).Msg("Failed to save active staking delegation")
		return types.NewInternalServiceError(err)
	}
	return nil
}

func (s *V1Service) IsDelegationPresent(ctx context.Context, txHashHex string) (bool, *types.Error) {
	delegation, err := s.Service.DbClients.V1DBClient.FindDelegationByTxHashHex(ctx, txHashHex)
	if err != nil {
		if db.IsNotFoundError(err) {
			return false, nil
		}
		log.Ctx(ctx).Error().Err(err).Msg("Failed to find delegation by tx hash hex")
		return false, types.NewInternalServiceError(err)
	}
	if delegation != nil {
		return true, nil
	}

	return false, nil
}

func (s *V1Service) GetDelegation(ctx context.Context, txHashHex string) (*DelegationPublic, *types.Error) {
	delegation, err := s.Service.DbClients.V1DBClient.FindDelegationByTxHashHex(ctx, txHashHex)
	if err != nil {
		if db.IsNotFoundError(err) {
			log.Ctx(ctx).Warn().Err(err).Str("stakingTxHash", txHashHex).Msg("Staking delegation not found")
			return nil, types.NewErrorWithMsg(http.StatusNotFound, types.NotFound, "staking delegation not found, please retry")
		}
		log.Ctx(ctx).Error().Err(err).Msg("Failed to find delegation by tx hash hex")
		return nil, types.NewInternalServiceError(err)
	}
	bbnHeight, err := s.Service.DbClients.IndexerDBClient.GetLastProcessedBbnHeight(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to get last processed BBN height")
		return nil, types.NewInternalServiceError(err)
	}

	// Get list of all finality providers in phase-2
	transitionedFps, err := s.Service.DbClients.IndexerDBClient.GetFinalityProviders(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to get finality providers")
		return nil, types.NewInternalServiceError(err)
	}

	return s.FromDelegationDocument(delegation, bbnHeight, transitionedFps), nil
}

func (s *V1Service) CheckStakerHasActiveDelegationByPk(
	ctx context.Context, stakerPk string, afterTimestamp int64,
) (bool, *types.Error) {
	filter := &indexerdbclient.DelegationFilter{
		States:         []indexertypes.DelegationState{indexertypes.StateActive},
		AfterTimestamp: afterTimestamp,
	}
	hasDelegation, err := s.Service.DbClients.IndexerDBClient.CheckDelegationExistByStakerPk(
		ctx, stakerPk, filter,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to check if staker has active delegation")
		return false, types.NewInternalServiceError(err)
	}
	return hasDelegation, nil
}

// This method checks if the finality provider is slashed and whether it is in the transitioned list
func (s *V1Service) checkFpStatus(
	fpPk string, transitionedFps []*indexerdbmodel.IndexerFinalityProviderDetails,
) (bool, bool) {
	for _, fp := range transitionedFps {
		if fp.BtcPk == fpPk {
			return true, fp.State == indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_SLASHED
		}
	}
	return false, false
}

func (s *V1Service) isEligibleForTransition(
	delegation *v1model.DelegationDocument, bbnHeight uint64,
) bool {
	if s.Cfg.DelegationTransition == nil {
		return false
	}

	// Check the delegation state, only active delegations are eligible for transition
	if delegation.State != types.Active {
		return false
	}

	// Check the delegation staking height
	stakingHeight := delegation.StakingTx.StartHeight
	// Only not overflow delegations are eligible for transition before the Btc height
	if !delegation.IsOverflow && stakingHeight < s.Cfg.DelegationTransition.EligibleBeforeBtcHeight {
		return true
	}
	if bbnHeight >= s.Cfg.DelegationTransition.AllowListExpirationHeight {
		return true
	}

	return false
}

func (s *V1Service) FromDelegationDocument(
	d *v1model.DelegationDocument, bbnHeight uint64,
	transitionedFps []*indexerdbmodel.IndexerFinalityProviderDetails,
) *DelegationPublic {
	isFpTransitioned, isSlashed := s.checkFpStatus(d.FinalityProviderPkHex, transitionedFps)
	delPublic := &DelegationPublic{
		StakingTxHashHex:      d.StakingTxHashHex,
		StakerPkHex:           d.StakerPkHex,
		FinalityProviderPkHex: d.FinalityProviderPkHex,
		StakingValue:          d.StakingValue,
		State:                 d.State.ToString(),
		StakingTx: &TransactionPublic{
			TxHex:          d.StakingTx.TxHex,
			OutputIndex:    d.StakingTx.OutputIndex,
			StartTimestamp: utils.ParseTimestampToIsoFormat(d.StakingTx.StartTimestamp),
			StartHeight:    d.StakingTx.StartHeight,
			TimeLock:       d.StakingTx.TimeLock,
		},
		IsOverflow:              d.IsOverflow,
		IsEligibleForTransition: isFpTransitioned && !isSlashed && s.isEligibleForTransition(d, bbnHeight),
		IsSlashed:               isSlashed,
	}

	// Add unbonding transaction if it exists
	if d.UnbondingTx != nil && d.UnbondingTx.TxHex != "" {
		delPublic.UnbondingTx = &TransactionPublic{
			TxHex:          d.UnbondingTx.TxHex,
			OutputIndex:    d.UnbondingTx.OutputIndex,
			StartTimestamp: utils.ParseTimestampToIsoFormat(d.UnbondingTx.StartTimestamp),
			StartHeight:    d.UnbondingTx.StartHeight,
			TimeLock:       d.UnbondingTx.TimeLock,
		}
	}
	return delPublic
}
