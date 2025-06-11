package v2handlers

import (
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2service "github.com/babylonlabs-io/staking-api-service/internal/v2/service"
	"net/http"
)

// GetEventConsumers gets event consumers (BSN-s)
// @Summary Get event consumers
// @Produce json
// @Tags v2
// @Success 200 {object} handler.PublicResponse[[]v2service.EventConsumer]{array} "List of available event consumers"
// @Failure 500 {object} types.Error "Error: Internal Server Error"
// @Router /v2/event-consumers [get]
func (h *V2Handler) GetEventConsumers(request *http.Request) (*handler.Result, *types.Error) {
	return handler.NewResultWithPagination([]v2service.EventConsumer{}, ""), nil
}
