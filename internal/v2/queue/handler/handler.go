package v2queuehandler

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/services"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
)

type V2QueueHandler struct {
	Services *services.Services
}

type MessageHandler func(ctx context.Context, messageBody string) *types.Error
type UnprocessableMessageHandler func(ctx context.Context, messageBody, receipt string) *types.Error

func NewV2QueueHandler(services *services.Services) *V2QueueHandler {
	return &V2QueueHandler{
		Services: services,
	}
}

func (qh *V2QueueHandler) HandleUnprocessedMessage(ctx context.Context, messageBody, receipt string) *types.Error {
	return qh.Services.SharedService.SaveUnprocessableMessages(ctx, messageBody, receipt)
}
