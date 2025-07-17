package handler

import (
	"fmt"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
	"net/http"
)

// AllowList godoc
// @Summary Checks that given staking transaction hash is part of allow-list
// @Produce json
// @Tags shared
// @Param staking_tx_hash query string true "Staking transaction hash"
// @Success 200 {string} handler.PublicResponse[bool] "Given stakingTxHash is in allow-list"
// @Router /v1/allow-list [get]
func (h *Handler) AllowList(request *http.Request) (*Result, *types.Error) {
	const key = "staking_tx_hash"

	stakingTxHash := request.URL.Query().Get(key)
	if stakingTxHash == "" {
		return nil, types.NewErrorWithMsg(http.StatusBadRequest, types.BadRequest, fmt.Sprintf("missing %q GET parameter", key))
	}

	ctx := request.Context()
	allowed, err := h.Service.IsTxInAllowList(ctx, stakingTxHash)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to check allowlist")
		return nil, types.NewErrorWithMsg(http.StatusInternalServerError, types.InternalServiceError, "failed to check allow list")
	}

	return NewResult(allowed), nil
}
