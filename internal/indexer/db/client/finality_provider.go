package indexerdbclient

import (
	"context"
	"fmt"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (indexerdbclient *IndexerDatabase) FindFinalityProviders(
	ctx context.Context, fpPk string, name string, searchQuery string, state types.FinalityProviderState, paginationToken string,
) (*db.DbResultMap[indexerdbmodel.IndexerFinalityProviderDetails], error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).Collection(indexerdbmodel.FinalityProviderDetailsCollection)

	filter := bson.M{}
	filter = indexerdbclient.applyExactMatchFilters(filter, fpPk, name)
	filter = indexerdbclient.applyStateFilter(filter, state)
	filter = indexerdbclient.applySearchFilter(filter, searchQuery)

	// Default sort by commission, then by btc_pk for stable sorting
	options := options.Find().SetSort(bson.D{
		{Key: "commission", Value: 1},
		{Key: "_id", Value: 1},
	})

	filter = indexerdbclient.applyPaginationFilter(filter, paginationToken)

	return db.FindWithPagination(
		ctx, client, filter, options, indexerdbclient.Cfg.MaxPaginationLimit,
		indexerdbmodel.BuildFinalityProviderPaginationToken,
	)
}

func (indexerdbclient *IndexerDatabase) applyExactMatchFilters(filter bson.M, fpPk string, name string) bson.M {
	if fpPk != "" {
		filter["_id"] = fpPk
	}
	if name != "" {
		filter["description.moniker"] = name
	}
	return filter
}

func (indexerdbclient *IndexerDatabase) applyStateFilter(filter bson.M, state types.FinalityProviderState) bson.M {
	if state == types.FinalityProviderStateActive {
		filter["state"] = indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_ACTIVE
	} else if state == types.FinalityProviderStateStandby {
		filter["state"] = bson.M{
			"$in": []indexerdbmodel.IndexerFinalityProviderState{
				indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_INACTIVE,
				indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_JAILED,
				indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_SLASHED,
			},
		}
	}
	fmt.Println("filter", filter)
	return filter
}

func (indexerdbclient *IndexerDatabase) applySearchFilter(filter bson.M, searchQuery string) bson.M {
	if searchQuery == "" {
		return filter
	}

	searchFilter := bson.M{
		"$or": []bson.M{
			{
				"_id": bson.M{
					"$regex":   searchQuery,
					"$options": "i", // case-insensitive
				},
			},
			{
				"description.moniker": bson.M{
					"$regex":   searchQuery,
					"$options": "i", // case-insensitive
				},
			},
		},
	}

	if len(filter) > 0 {
		return bson.M{
			"$and": []bson.M{
				filter,
				searchFilter,
			},
		}
	}
	return searchFilter
}

func (indexerdbclient *IndexerDatabase) applyPaginationFilter(filter bson.M, paginationToken string) bson.M {
	if paginationToken == "" {
		return filter
	}

	decodedToken, err := dbmodel.DecodePaginationToken[indexerdbmodel.IndexerFinalityProviderPagination](paginationToken)
	if err != nil {
		return filter
	}

	paginationFilter := bson.M{
		"$or": []bson.M{
			{
				"commission": decodedToken.Commission,
				"_id":        bson.M{"$gt": decodedToken.BtcPk},
			},
			{
				"commission": bson.M{"$gt": decodedToken.Commission},
			},
		},
	}

	if len(filter) > 0 {
		return bson.M{
			"$and": []bson.M{
				filter,
				paginationFilter,
			},
		}
	}
	return paginationFilter
}
