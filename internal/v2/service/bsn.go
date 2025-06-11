package v2service

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/pkg"
)

type EventConsumer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *V2Service) GetEventConsumers(ctx context.Context) ([]EventConsumer, error) {
	items, err := s.dbClients.IndexerDBClient.GetEventConsumers(ctx)
	if err != nil {
		return nil, err
	}

	return pkg.Map(items, mapEventConsumer), nil
}

func mapEventConsumer(consumer indexerdbmodel.EventConsumer) EventConsumer {
	return EventConsumer{
		ID:          consumer.ID,
		Name:        consumer.Name,
		Description: consumer.Description,
	}
}
