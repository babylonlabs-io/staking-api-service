package v2service

import (
	"context"
	"net/http"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	v2types "github.com/babylonlabs-io/staking-api-service/internal/v2/types"
	"github.com/rs/zerolog/log"
)

type DelegationStaking struct {
	StakingTxHashHex   string          `json:"staking_tx_hash_hex"`
	StakingTxHex       string          `json:"staking_tx_hex"`
	StakingTimelock    uint32          `json:"staking_timelock"`
	StakingAmount      uint64          `json:"staking_amount"`
	StartHeight        uint32          `json:"start_height,omitempty"`
	EndHeight          uint32          `json:"end_height,omitempty"`
	BbnInceptionHeight int64           `json:"bbn_inception_height"`
	BbnInceptionTime   string          `json:"bbn_inception_time"`
	Slashing           StakingSlashing `json:"slashing"`
}

type StakingSlashing struct {
	SlashingTxHex  string `json:"slashing_tx_hex"`
	SpendingHeight uint32 `json:"spending_height"`
}

type UnbondingSlashing struct {
	UnbondingSlashingTxHex string `json:"unbonding_slashing_tx_hex"`
	SpendingHeight         uint32 `json:"spending_height"`
}

type CovenantSignature struct {
	CovenantBtcPkHex string `json:"covenant_btc_pk_hex"`
	SignatureHex     string `json:"signature_hex"`
}

type DelegationUnbonding struct {
	UnbondingTimelock           uint32              `json:"unbonding_timelock"`
	UnbondingTx                 string              `json:"unbonding_tx"`
	CovenantUnbondingSignatures []CovenantSignature `json:"covenant_unbonding_signatures"`
	Slashing                    UnbondingSlashing   `json:"slashing"`
}

type DelegationPublic struct {
	ParamsVersion             uint32                  `json:"params_version"`
	StakerBtcPkHex            string                  `json:"staker_btc_pk_hex"`
	FinalityProviderBtcPksHex []string                `json:"finality_provider_btc_pks_hex"`
	DelegationStaking         DelegationStaking       `json:"delegation_staking"`
	DelegationUnbonding       DelegationUnbonding     `json:"delegation_unbonding"`
	State                     v2types.DelegationState `json:"state"`
}

func FromDelegationDocument(delegation indexerdbmodel.IndexerDelegationDetails) (*DelegationPublic, *types.Error) {
	state, err := v2types.MapDelegationState(delegation.State, delegation.SubState)
	if err != nil {
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			"failed to get delegation state",
		)
	}

	delegationPublic := &DelegationPublic{
		ParamsVersion:             delegation.ParamsVersion,
		FinalityProviderBtcPksHex: delegation.FinalityProviderBtcPksHex,
		StakerBtcPkHex:            delegation.StakerBtcPkHex,
		DelegationStaking: DelegationStaking{
			StakingTxHashHex:   delegation.StakingTxHashHex,
			StakingTxHex:       delegation.StakingTxHex,
			StakingTimelock:    delegation.StakingTimeLock,
			StakingAmount:      delegation.StakingAmount,
			StartHeight:        delegation.StartHeight,
			EndHeight:          delegation.EndHeight,
			BbnInceptionHeight: delegation.BTCDelegationCreatedBbnBlock.Height,
			BbnInceptionTime: utils.ParseTimestampToIsoFormat(
				delegation.BTCDelegationCreatedBbnBlock.Timestamp,
			),
			Slashing: StakingSlashing{
				SlashingTxHex:  delegation.SlashingTx.SlashingTxHex,
				SpendingHeight: delegation.SlashingTx.SpendingHeight,
			},
		},
		DelegationUnbonding: DelegationUnbonding{
			UnbondingTimelock: delegation.UnbondingTimeLock,
			UnbondingTx:       delegation.UnbondingTx,
			CovenantUnbondingSignatures: getUnbondingSignatures(
				delegation.CovenantUnbondingSignatures,
			),
			Slashing: UnbondingSlashing{
				UnbondingSlashingTxHex: delegation.SlashingTx.UnbondingSlashingTxHex,
				SpendingHeight:         delegation.SlashingTx.SpendingHeight,
			},
		},
		State: state,
	}

	return delegationPublic, nil
}

func (s *V2Service) GetDelegation(ctx context.Context, stakingTxHashHex string) (*DelegationPublic, *types.Error) {
	delegation, err := s.DbClients.IndexerDBClient.GetDelegation(ctx, stakingTxHashHex)
	if err != nil {
		if db.IsNotFoundError(err) {
			log.Ctx(ctx).Warn().Err(err).Str("stakingTxHashHex", stakingTxHashHex).Msg("Staking delegation not found")
			return nil, types.NewErrorWithMsg(http.StatusNotFound, types.NotFound, "staking delegation not found, please retry")
		}
		return nil, types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "failed to get staker delegation")
	}

	return FromDelegationDocument(*delegation)
}

func (s *V2Service) GetDelegations(ctx context.Context, stakerPkHex string, paginationKey string) ([]*DelegationPublic, string, *types.Error) {
	resultMap, err := s.DbClients.IndexerDBClient.GetDelegations(ctx, stakerPkHex, paginationKey)
	if err != nil {
		if db.IsNotFoundError(err) {
			log.Ctx(ctx).Warn().Err(err).Str("stakingTxHashHex", stakerPkHex).Msg("Staking delegations not found")
			return nil, "", types.NewErrorWithMsg(http.StatusNotFound, types.NotFound, "staking delegation not found, please retry")
		}
		return nil, "", types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "failed to get staker delegations")
	}

	// Initialize result structure
	delegationsPublic := make([]*DelegationPublic, 0, len(resultMap.Data))

	// Group delegations by state
	for _, delegation := range resultMap.Data {
		delegationPublic, delErr := FromDelegationDocument(delegation)
		if delErr != nil {
			return nil, "", delErr
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

func (s *V2Service) SaveUnprocessableMessages(ctx context.Context, messageBody, receipt string) *types.Error {
	err := s.DbClients.V2DBClient.SaveUnprocessableMessage(ctx, messageBody, receipt)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while saving unprocessable message")
		return types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "error while saving unprocessable message")
	}
	return nil
}

// MarkV1DelegationAsTransitioned marks a v1 delegation as transitioned
func (s *V2Service) MarkV1DelegationAsTransitioned(
	ctx context.Context, stakingTxHashHex string,
) *types.Error {
	err := s.DbClients.V1DBClient.TransitionToTransitionedState(ctx, stakingTxHashHex)
	if err != nil {
		if db.IsNotFoundError(err) {
			// If the delegation is not found, it means it has already been transitioned
			// or not relevant to phase-1 at all.
			return nil
		}
		log.Ctx(ctx).Error().Err(err).Msg("Failed to transition v1 delegation to transitioned state")
		return types.NewInternalServiceError(err)
	}
	return nil
}
