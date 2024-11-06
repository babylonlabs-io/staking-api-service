package indexerdbclient

import (
	"context"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetFinalityProviderByPk retrieves a single finality provider by their primary key
func (indexerdbclient *IndexerDatabase) GetFinalityProviderByPk(
	ctx context.Context,
	fpPk string,
) (*indexerdbmodel.IndexerFinalityProviderDetails, error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).Collection(indexerdbmodel.FinalityProviderDetailsCollection)

	filter := bson.M{}
	filter = indexerdbclient.applyFpPkFilter(filter, fpPk)

	var result indexerdbmodel.IndexerFinalityProviderDetails
	err := client.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetFinalityProviders retrieves finality providers filtered by state
func (indexerdbclient *IndexerDatabase) GetFinalityProviders(
	ctx context.Context,
	state types.FinalityProviderState,
	paginationToken string,
) (*db.DbResultMap[indexerdbmodel.IndexerFinalityProviderDetails], error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).Collection(indexerdbmodel.FinalityProviderDetailsCollection)

	filter := bson.M{}
	filter = indexerdbclient.applyStateFilter(filter, state)
	filter = indexerdbclient.applyPaginationFilter(filter, paginationToken)

	options := options.Find().SetSort(bson.D{
		{Key: "commission", Value: 1},
		{Key: "_id", Value: 1},
	})

	return db.FindWithPagination(
		ctx, client, filter, options, indexerdbclient.Cfg.MaxPaginationLimit,
		indexerdbmodel.BuildFinalityProviderPaginationToken,
	)
}

// SearchFinalityProviders performs a text search across finality providers
func (indexerdbclient *IndexerDatabase) SearchFinalityProviders(
	ctx context.Context,
	searchQuery string,
	paginationToken string,
) (*db.DbResultMap[indexerdbmodel.IndexerFinalityProviderDetails], error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).Collection(indexerdbmodel.FinalityProviderDetailsCollection)

	filter := indexerdbclient.applySearchFilter(bson.M{}, searchQuery)
	filter = indexerdbclient.applyPaginationFilter(filter, paginationToken)

	options := options.Find().SetSort(bson.D{
		{Key: "commission", Value: 1},
		{Key: "_id", Value: 1},
	})

	return db.FindWithPagination(
		ctx, client, filter, options, indexerdbclient.Cfg.MaxPaginationLimit,
		indexerdbmodel.BuildFinalityProviderPaginationToken,
	)
}

func (indexerdbclient *IndexerDatabase) applyFpPkFilter(filter bson.M, fpPk string) bson.M {
	if fpPk != "" {
		filter["_id"] = fpPk
	}
	return filter
}

func (indexerdbclient *IndexerDatabase) applyMonikerFilters(filter bson.M, moniker string) bson.M {
	if moniker != "" {
		filter["description.moniker"] = moniker
	}
	return filter
}

func (indexerdbclient *IndexerDatabase) applyStateFilter(filter bson.M, state types.FinalityProviderState) bson.M {
	if state == types.FinalityProviderStateActive {
		filter["state"] = indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_ACTIVE
	} else if state == types.FinalityProviderStateStandby {
		filter["state"] = bson.M{
			"$in": []indexerdbmodel.FinalityProviderState{
				indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_INACTIVE,
				indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_JAILED,
				indexerdbmodel.FinalityProviderStatus_FINALITY_PROVIDER_STATUS_SLASHED,
			},
		}
	}
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
