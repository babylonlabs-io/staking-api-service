package services

import (
	"context"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/db"
	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"github.com/babylonlabs-io/staking-api-service/internal/utils"
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
	StakingTxHashHex      string             `json:"staking_tx_hash_hex"`
	StakerPkHex           string             `json:"staker_pk_hex"`
	FinalityProviderPkHex string             `json:"finality_provider_pk_hex"`
	State                 string             `json:"state"`
	StakingValue          uint64             `json:"staking_value"`
	StakingTx             *TransactionPublic `json:"staking_tx"`
	UnbondingTx           *TransactionPublic `json:"unbonding_tx,omitempty"`
	IsOverflow            bool               `json:"is_overflow"`
}

func FromDelegationDocument(d *model.DelegationDocument) DelegationPublic {
	delPublic := DelegationPublic{
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
		IsOverflow: d.IsOverflow,
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

func (s *Services) DelegationsByStakerPk(ctx context.Context, stakerPk string, pageToken string) ([]DelegationPublic, string, *types.Error) {
	resultMap, err := s.DbClient.FindDelegationsByStakerPk(ctx, stakerPk, pageToken)
	if err != nil {
		if db.IsInvalidPaginationTokenError(err) {
			log.Ctx(ctx).Warn().Err(err).Msg("Invalid pagination token when fetching delegations by staker pk")
			return nil, "", types.NewError(http.StatusBadRequest, types.BadRequest, err)
		}
		log.Ctx(ctx).Error().Err(err).Msg("Failed to find delegations by staker pk")
		return nil, "", types.NewInternalServiceError(err)
	}
	var delegations []DelegationPublic = make([]DelegationPublic, 0, len(resultMap.Data))
	for _, d := range resultMap.Data {
		delegations = append(delegations, FromDelegationDocument(&d))
	}
	return delegations, resultMap.PaginationToken, nil
}

// SaveActiveStakingDelegation saves the active staking delegation to the database.
func (s *Services) SaveActiveStakingDelegation(
	ctx context.Context, txHashHex, stakerPkHex, finalityProviderPkHex string,
	value, startHeight uint64, stakingTimestamp int64, timeLock, stakingOutputIndex uint64,
	stakingTxHex string, isOverflow bool,
) *types.Error {
	addresses, err := utils.DeriveAddressesFromNoCoordPk(stakerPkHex, s.cfg.Server.BTCNetParam)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to get taproot address from staker pk")
		return types.NewErrorWithMsg(
			http.StatusBadRequest, types.BadRequest, "failed to get taproot address from staker pk",
		)
	}
	err = s.DbClient.SaveActiveStakingDelegation(
		ctx, txHashHex, stakerPkHex, finalityProviderPkHex, stakingTxHex,
		value, startHeight, timeLock, stakingOutputIndex, stakingTimestamp, isOverflow, addresses.Taproot,
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

func (s *Services) IsDelegationPresent(ctx context.Context, txHashHex string) (bool, *types.Error) {
	delegation, err := s.DbClient.FindDelegationByTxHashHex(ctx, txHashHex)
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

func (s *Services) GetDelegation(ctx context.Context, txHashHex string) (*model.DelegationDocument, *types.Error) {
	delegation, err := s.DbClient.FindDelegationByTxHashHex(ctx, txHashHex)
	if err != nil {
		if db.IsNotFoundError(err) {
			log.Ctx(ctx).Warn().Err(err).Str("stakingTxHash", txHashHex).Msg("Staking delegation not found")
			return nil, types.NewErrorWithMsg(http.StatusNotFound, types.NotFound, "staking delegation not found, please retry")
		}
		log.Ctx(ctx).Error().Err(err).Msg("Failed to find delegation by tx hash hex")
		return nil, types.NewInternalServiceError(err)
	}
	return delegation, nil
}

func (s *Services) CheckStakerHasActiveDelegationByAddress(
	ctx context.Context, btcAddress string, afterTimestamp int64,
) (bool, *types.Error) {
	filter := &db.DelegationFilter{
		States:         []types.DelegationState{types.Active},
		AfterTimestamp: afterTimestamp,
	}
	hasDelegation, err := s.DbClient.CheckDelegationExistByStakerTaprootAddress(
		ctx, btcAddress, filter,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to check if staker has active delegation")
		return false, types.NewInternalServiceError(err)
	}
	return hasDelegation, nil
}
