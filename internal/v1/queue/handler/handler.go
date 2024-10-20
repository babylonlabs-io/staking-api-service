package v1queuehandler

import (
	"context"

	queuehandler "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1service "github.com/babylonlabs-io/staking-api-service/internal/v1/api/service"
)

type V1QueueHandler struct {
	*queuehandler.QueueHandler
	Service v1service.V1ServiceInterface
}

func New(queueHandler *queuehandler.QueueHandler, service v1service.V1ServiceInterface) *V1QueueHandler {
	return &V1QueueHandler{
		QueueHandler: queueHandler,
		Service:      service,
	}
}

func (qh *V1QueueHandler) HandleUnprocessedMessage(ctx context.Context, messageBody, receipt string) *types.Error {
	return qh.Service.SaveUnprocessableMessages(ctx, messageBody, receipt)
}
