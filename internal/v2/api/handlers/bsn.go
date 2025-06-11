package v2handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// GetEventConsumers gets event consumers (BSN-s)
// @Summary Get event consumers
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[[]v2service.EventConsumer]{array} "List of available event consumers"
// @Failure 500 {object} types.Error "Error: Internal Server Error"
// @Router /v2/event-consumers [get]
func (h *V2Handler) GetEventConsumers(request *http.Request) (*handler.Result, *types.Error) {
	items, err := h.Service.GetEventConsumers(request.Context())
	if err != nil {
		return nil, types.NewInternalServiceError(err)
	}

	return handler.NewResultWithPagination(items, ""), nil
}
