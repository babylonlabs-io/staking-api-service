package handler

import (
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Health check the service, including ping database connection
// @Produce json
// @Tags shared
// @Success 200 {string} handler.PublicResponse[string] "Server is up and running"
// @Router /healthcheck [get]
func (h *Handler) HealthCheck(request *http.Request) (*Result, *types.Error) {
	err := h.Service.DoHealthCheck(request.Context())
	if err != nil {
		return nil, types.NewInternalServiceError(err)
	}

	return NewResult("Server is up and running"), nil
}
