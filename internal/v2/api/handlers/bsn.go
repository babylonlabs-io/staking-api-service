package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetAllBSN gets event consumers (BSN-s)
// @Summary Get event consumers
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[[]v2service.BSN]{array} "List of available event consumers"
// @Failure 500 {object} types.Error "Error: Internal Server Error"
// @Router /v2/bsn [get]
func (h *V2Handler) GetAllBSN(request *http.Request) (*handler.Result, *types.Error) {
	items, err := h.Service.GetAllBSN(request.Context())
	if err != nil {
		return nil, types.NewInternalServiceError(err)
	}

	return handler.NewResultWithPagination(nonNilSlice(items), ""), nil
}

func nonNilSlice[T any](sl []T) []T {
	if sl == nil {
		return []T{}
	}

	return sl
}
