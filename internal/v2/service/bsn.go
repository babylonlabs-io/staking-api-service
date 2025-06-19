package v2service

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/pkg"
)

type BSN struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ActiveTvl   int64  `json:"active_tvl"`
}

var babylonBSN = BSN{
	ID:          "", // empty string corresponds to babylon network
	Name:        "Babylon",
	Description: "Babylon network",
	ActiveTvl:   0,
}

func (s *V2Service) GetAllBSN(ctx context.Context) ([]BSN, error) {
	items, err := s.dbClients.IndexerDBClient.GetAllBSN(ctx)
	if err != nil {
		return nil, err
	}

	// because Babylon BSN is not stored in the db we add it in service layer
	allItems := []BSN{babylonBSN}
	allItems = append(allItems, pkg.Map(items, mapBSN)...)
	return allItems, nil
}

func mapBSN(consumer indexerdbmodel.BSN) BSN {
	return BSN{
		ID:          consumer.ID,
		Name:        consumer.Name,
		Description: consumer.Description,
	}
}
