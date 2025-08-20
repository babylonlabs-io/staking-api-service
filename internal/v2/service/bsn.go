package v2service

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
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

	stats, err := s.dbClients.V2DBClient.GetBsnStats(ctx)
	if err != nil {
		return nil, err
	}
	statsByBSN := pkg.SliceToMap(stats, func(doc *v2dbmodel.BSNStatsDocument) string {
		return doc.BsnID
	})

  // we don't store babylon bsn in mongo, we place it on top so on frontend
	// it's always displayed first
	result := []BSN{
		{
			ID:          s.sharedService.ChainInfo.ChainID,
			Name:        "Babylon Genesis",
			Description: "",
		},
	}
	for _, item := range items {
		var activeTVL int64
		if v, ok := statsByBSN[item.ID]; ok {
			activeTVL = v.ActiveTvl
		}

		resultItem := mapBSN(item, activeTVL)
		result = append(result, resultItem)
	}

	return result, nil
}

func mapBSN(consumer indexerdbmodel.BSN, activeTVL int64) BSN {
	return BSN{
		ID:          consumer.ID,
		Name:        consumer.Name,
		Description: consumer.Description,
		ActiveTvl:   activeTVL,
	}
}
