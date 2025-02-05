package v1service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

type OverallStatsPublic struct {
	ActiveTvl         int64    `json:"active_tvl"`
	TotalTvl          int64    `json:"total_tvl"`
	ActiveDelegations int64    `json:"active_delegations"`
	TotalDelegations  int64    `json:"total_delegations"`
	TotalStakers      uint64   `json:"total_stakers"`
	UnconfirmedTvl    uint64   `json:"unconfirmed_tvl"`
	PendingTvl        uint64   `json:"pending_tvl"`
	BtcPriceUsd       *float64 `json:"btc_price_usd,omitempty"` // Optional field
}

type StakerStatsPublic struct {
	StakerPkHex       string `json:"staker_pk_hex"`
	ActiveTvl         int64  `json:"active_tvl"`
	TotalTvl          int64  `json:"total_tvl"`
	ActiveDelegations int64  `json:"active_delegations"`
	TotalDelegations  int64  `json:"total_delegations"`
}

// ProcessStakingStatsCalculation calculates the staking stats and updates the database.
// This method tolerates duplicated calls, only the first call will be processed.
func (s *V1Service) ProcessStakingStatsCalculation(
	ctx context.Context, stakingTxHashHex, stakerPkHex, fpPkHex string,
	state types.DelegationState, amount uint64,
) *types.Error {
	// Fetch existing or initialize the stats lock document if not exist
	statsLockDocument, err := s.Service.DbClients.V1DBClient.GetOrCreateStatsLock(
		ctx, stakingTxHashHex, state.ToString(),
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
			Msg("error while fetching stats lock document")
		return types.NewInternalServiceError(err)
	}
	switch state {
	case types.Active:
		// Add to the finality stats
		if !statsLockDocument.FinalityProviderStats {
			err = s.Service.DbClients.V1DBClient.IncrementFinalityProviderStats(
				ctx, stakingTxHashHex, fpPkHex, amount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
					Msg("error while incrementing finality stats")
				return types.NewInternalServiceError(err)
			}
		}
		if !statsLockDocument.StakerStats {
			// Convert the staker public key to multiple BTC addresses and save
			// them in the database.
			if addressConversionErr := s.ProcessAndSaveBtcAddresses(
				ctx, stakerPkHex,
			); addressConversionErr != nil {
				log.Ctx(ctx).Error().Err(addressConversionErr).
					Str("stakingTxHashHex", stakingTxHashHex).
					Msg("error while processing and saving btc addresses")
				return types.NewInternalServiceError(addressConversionErr)
			}
			err = s.Service.DbClients.V1DBClient.IncrementStakerStats(
				ctx, stakingTxHashHex, stakerPkHex, amount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
					Msg("error while incrementing staker stats")
				return types.NewInternalServiceError(err)
			}
		}
		// Add to the overall stats
		// The overall stats should be the last to be updated as it has dependency
		// on staker stats.
		if !statsLockDocument.OverallStats {
			err = s.Service.DbClients.V1DBClient.IncrementOverallStats(
				ctx, stakingTxHashHex, stakerPkHex, amount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					// This is a duplicate call, ignore it
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
					Msg("error while incrementing overall stats")
				return types.NewInternalServiceError(err)
			}
		}
	case types.Unbonded:
		// Subtract from the finality stats
		if !statsLockDocument.FinalityProviderStats {
			err = s.Service.DbClients.V1DBClient.SubtractFinalityProviderStats(
				ctx, stakingTxHashHex, fpPkHex, amount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
					Msg("error while subtracting finality stats")
				return types.NewInternalServiceError(err)
			}
		}
		if !statsLockDocument.StakerStats {
			err = s.Service.DbClients.V1DBClient.SubtractStakerStats(
				ctx, stakingTxHashHex, stakerPkHex, amount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
					Msg("error while subtracting staker stats")
				return types.NewInternalServiceError(err)
			}
		}
		// Subtract from the overall stats.
		// The overall stats should be the last to be updated as it has dependency
		// on staker stats.
		if !statsLockDocument.OverallStats {
			err = s.Service.DbClients.V1DBClient.SubtractOverallStats(
				ctx, stakingTxHashHex, stakerPkHex, amount,
			)
			if err != nil {
				if db.IsNotFoundError(err) {
					return nil
				}
				log.Ctx(ctx).Error().Err(err).Str("stakingTxHashHex", stakingTxHashHex).
					Msg("error while subtracting overall stats")
				return types.NewInternalServiceError(err)
			}
		}
	default:
		return types.NewErrorWithMsg(
			http.StatusBadRequest,
			types.BadRequest,
			fmt.Sprintf("invalid delegation state for stats calculation: %s", state),
		)
	}
	return nil
}

func (s *V1Service) GetOverallStats(
	ctx context.Context,
) (*OverallStatsPublic, *types.Error) {
	stats, err := s.Service.DbClients.V1DBClient.GetOverallStats(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching overall stats")
		return nil, types.NewInternalServiceError(err)
	}

	var unconfirmedTvl, confirmedTvl, pendingTvl uint64

	btcInfo, err := s.Service.DbClients.V1DBClient.GetLatestBtcInfo(ctx)
	if err != nil {
		// Handle missing BTC information, which may occur during initial setup.
		// Default the unconfirmed TVL to 0; this will be updated automatically
		// after processing new BTC blocks, all subsequent requests will be served
		// with the correct value.
		if db.IsNotFoundError(err) {
			log.Ctx(ctx).Error().Err(err).Msg("latest btc info not found")
		} else {
			log.Ctx(ctx).Error().Err(err).Msg("error while fetching latest btc info")
			return nil, types.NewInternalServiceError(err)
		}
	} else {
		unconfirmedTvl = btcInfo.UnconfirmedTvl
		confirmedTvl = btcInfo.ConfirmedTvl
		pendingTvl = unconfirmedTvl - confirmedTvl
	}

	return &OverallStatsPublic{
		ActiveTvl:         int64(confirmedTvl),
		TotalTvl:          stats.TotalTvl,
		ActiveDelegations: stats.ActiveDelegations,
		TotalDelegations:  stats.TotalDelegations,
		TotalStakers:      stats.TotalStakers,
		UnconfirmedTvl:    unconfirmedTvl,
		PendingTvl:        pendingTvl,
	}, nil
}

func (s *V1Service) GetStakerStats(
	ctx context.Context, stakerPkHex string,
) (*StakerStatsPublic, *types.Error) {
	stats, err := s.Service.DbClients.V1DBClient.GetStakerStats(ctx, stakerPkHex)
	if err != nil {
		// Not found error is not an error, return nil
		if db.IsNotFoundError(err) {
			return nil, nil
		}
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching staker stats")
		return nil, types.NewInternalServiceError(err)
	}

	return &StakerStatsPublic{
		StakerPkHex:       stakerPkHex,
		ActiveTvl:         stats.ActiveTvl,
		TotalTvl:          stats.TotalTvl,
		ActiveDelegations: stats.ActiveDelegations,
		TotalDelegations:  stats.TotalDelegations,
	}, nil
}

func (s *V1Service) GetTopStakersByActiveTvl(
	ctx context.Context, pageToken string,
) ([]StakerStatsPublic, string, *types.Error) {
	resultMap, err := s.Service.DbClients.V1DBClient.FindTopStakersByTvl(ctx, pageToken)
	if err != nil {
		if db.IsInvalidPaginationTokenError(err) {
			log.Ctx(ctx).Warn().Err(err).
				Msg("invalid pagination token while fetching top stakers by active tvl")
			return nil, "", types.NewError(http.StatusBadRequest, types.BadRequest, err)
		}
		log.Ctx(ctx).Error().Err(err).Msg("error while fetching top stakers by active tvl")
		return nil, "", types.NewInternalServiceError(err)
	}
	var topStakersStats []StakerStatsPublic
	for _, d := range resultMap.Data {
		topStakersStats = append(topStakersStats, StakerStatsPublic{
			StakerPkHex:       d.StakerPkHex,
			ActiveTvl:         d.ActiveTvl,
			TotalTvl:          d.TotalTvl,
			ActiveDelegations: d.ActiveDelegations,
			TotalDelegations:  d.TotalDelegations,
		})
	}

	return topStakersStats, resultMap.PaginationToken, nil
}

func (s *V1Service) ProcessBtcInfoStats(
	ctx context.Context, btcHeight uint64, confirmedTvl uint64, unconfirmedTvl uint64,
) *types.Error {
	err := s.Service.DbClients.V1DBClient.UpsertLatestBtcInfo(ctx, btcHeight, confirmedTvl, unconfirmedTvl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error while upserting latest btc info")
		return types.NewInternalServiceError(err)
	}
	return nil
}
