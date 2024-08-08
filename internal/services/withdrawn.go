package services

import (
	"context"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/db"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"github.com/rs/zerolog/log"
)

func (s *Services) TransitionToWithdrawnState(
	ctx context.Context, stakingTxHashHex string,
) *types.Error {
	err := s.DbClient.TransitionToWithdrawnState(ctx, stakingTxHashHex)
	if err != nil {
		if ok := db.IsNotFoundError(err); ok {
			log.Ctx(ctx).Warn().Str("stakingTxHashHex", stakingTxHashHex).Err(err).Msg("delegation not found or no longer eligible for withdraw")
			return types.NewErrorWithMsg(http.StatusForbidden, types.NotFound, "delegation not found or no longer eligible for withdraw")
		}
		log.Ctx(ctx).Error().Str("stakingTxHashHex", stakingTxHashHex).Err(err).Msg("failed to transition to withdrawn state")
		return types.NewError(http.StatusInternalServerError, types.InternalServiceError, err)
	}
	return nil
}
