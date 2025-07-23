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

func (s *V2Service) GetAllBSN(ctx context.Context) ([]BSN, error) {
	items, err := s.dbClients.IndexerDBClient.GetAllBSN(ctx)
	if err != nil {
		return nil, err
	}

	networkInfo, err := s.dbClients.IndexerDBClient.GetNetworkInfo(ctx)
	if err != nil {
		return nil, err
	}

	result := []BSN{
		{
			ID:          networkInfo.ChainID,
			Name:        "Babylon network",
			Description: "",
		},
	}
	result = append(result, pkg.Map(items, mapBSN)...)

	return result, nil
}

func mapBSN(consumer indexerdbmodel.BSN) BSN {
	return BSN{
		ID:          consumer.ID,
		Name:        consumer.Name,
		Description: consumer.Description,
	}
}
