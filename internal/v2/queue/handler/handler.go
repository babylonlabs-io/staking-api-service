package v2queuehandler

import (
	"context"

	queuehandler "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2service "github.com/babylonlabs-io/staking-api-service/internal/v2/service"
)

type V2QueueHandler struct {
	*queuehandler.QueueHandler
	Service v2service.V2ServiceProvider
}

func New(queueHandler *queuehandler.QueueHandler, service v2service.V2ServiceProvider) *V2QueueHandler {
	return &V2QueueHandler{
		QueueHandler: queueHandler,
		Service:      service,
	}
}

func (qh *V2QueueHandler) HandleUnprocessedMessage(ctx context.Context, messageBody, receipt string) *types.Error {
	return qh.Service.SaveUnprocessableMessages(ctx, messageBody, receipt)
}
