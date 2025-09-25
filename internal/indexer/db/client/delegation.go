package indexerdbclient

import (
	"context"
	"errors"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DelegationsQueryFilter func(options bson.M)

func WithStakerPKHex(pkHex string) DelegationsQueryFilter {
	return func(options bson.M) {
		options["staker_btc_pk_hex"] = pkHex
	}
}

func WithBabylonAddress(address string) DelegationsQueryFilter {
	return func(options bson.M) {
		options["staker_babylon_address"] = address
	}
}

func WithState(state indexertypes.DelegationState) DelegationsQueryFilter {
	return func(options bson.M) {
		options["state"] = state
	}
}

// DumpFilters iterates over filteres and record all the changes in a map
// this function should be used only for logging purposes
func DumpFilters(filters ...DelegationsQueryFilter) map[string]any {
	filtersOptions := bson.M{}
	for _, filter := range filters {
		filter(filtersOptions)
	}

	return filtersOptions
}

func (indexerdbclient *IndexerDatabase) GetDelegation(
	ctx context.Context, stakingTxHashHex string,
) (*indexerdbmodel.IndexerDelegationDetails, error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).
		Collection(indexerdbmodel.BTCDelegationDetailsCollection)
	filter := bson.M{"_id": stakingTxHashHex}
	var delegation indexerdbmodel.IndexerDelegationDetails
	err := client.FindOne(ctx, filter).Decode(&delegation)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &db.NotFoundError{
				Key:     stakingTxHashHex,
				Message: "Delegation not found",
			}
		}
		return nil, err
	}
	return &delegation, nil
}

func (indexerdbclient *IndexerDatabase) GetDelegations(
	ctx context.Context,
	paginationToken string,
	filters ...DelegationsQueryFilter,
) (*db.DbResultMap[indexerdbmodel.IndexerDelegationDetails], error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).
		Collection(indexerdbmodel.BTCDelegationDetailsCollection)

	filterMap := bson.M{}
	for _, filter := range filters {
		filter(filterMap)
	}

	// Default sort by start_height for stable sorting
	opts := options.Find().SetSort(bson.D{
		{Key: "btc_delegation_created_bbn_block.height", Value: -1},
		{Key: "_id", Value: 1},
	})

	// Decode the pagination token if it exists
	if paginationToken != "" {
		decodedToken, err := dbmodel.DecodePaginationToken[indexerdbmodel.IndexerDelegationPagination](paginationToken)
		if err != nil {
			return nil, &db.InvalidPaginationTokenError{
				Message: "Invalid pagination token",
			}
		}

		orConditions := []bson.M{
			{
				"btc_delegation_created_bbn_block.height": bson.M{"$lt": decodedToken.StartHeight},
			},
			{
				"btc_delegation_created_bbn_block.height": decodedToken.StartHeight,
				"_id": bson.M{"$gt": decodedToken.StakingTxHashHex},
			},
		}

		for _, filter := range filters {
			filter(orConditions[0])
			filter(orConditions[1])
		}

		filterMap = bson.M{
			"$or": orConditions,
		}
	}

	return db.FindWithPagination(
		ctx, client, filterMap, opts, indexerdbclient.Cfg.MaxPaginationLimit,
		indexerdbmodel.BuildDelegationPaginationToken,
	)
}

func (indexerdbclient *IndexerDatabase) GetDelegationsInStates(
	ctx context.Context,
	stakerPKHex string,
	stakerBabylonAddress *string,
	states []indexertypes.DelegationState,
) ([]indexerdbmodel.IndexerDelegationDetails, error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).
		Collection(indexerdbmodel.BTCDelegationDetailsCollection)

	// Base filter with staker_btc_pk_hex
	filter := bson.M{
		"staker_btc_pk_hex": stakerPKHex,
		"state":             bson.M{"$in": states},
	}

	if stakerBabylonAddress != nil {
		filter["staker_babylon_address"] = *stakerBabylonAddress
	}

	cursor, err := client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []indexerdbmodel.IndexerDelegationDetails
	if err = cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// CheckDelegationExistByStakerPk checks if a staker has any
// delegation in the specified states by the staker's public key
func (indexerdbclient *IndexerDatabase) CheckDelegationExistByStakerPk(
	ctx context.Context, stakerPk string, extraFilter *DelegationFilter,
) (bool, error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).Collection(indexerdbmodel.BTCDelegationDetailsCollection)
	filter := buildAdditionalDelegationFilter(
		bson.M{"staker_btc_pk_hex": stakerPk}, extraFilter,
	)
	var delegation indexerdbmodel.IndexerDelegationDetails
	err := client.FindOne(ctx, filter).Decode(&delegation)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func buildAdditionalDelegationFilter(
	baseFilter primitive.M,
	filters *DelegationFilter,
) primitive.M {
	if filters == nil {
		return baseFilter
	}

	if filters.States != nil {
		baseFilter["state"] = bson.M{"$in": filters.States}
	}
	if filters.AfterTimestamp != 0 {
		baseFilter["staking_btc_timestamp"] = bson.M{"$gte": filters.AfterTimestamp}
	}
	return baseFilter
}
