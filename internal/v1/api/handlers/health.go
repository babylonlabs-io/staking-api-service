package v1handlers

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Health check the service, including ping database connection
// @Produce json
// @Success 200 {string} PublicResponse[string] "Server is up and running"
// @Router /healthcheck [get]
func (h *V1Handler) HealthCheck(request *http.Request) (*handler.Result, *types.Error) {
	err := h.Service.DoHealthCheck(request.Context())
	if err != nil {
		return nil, types.NewInternalServiceError(err)
	}

	return handler.NewResult("Server is up and running"), nil
}
