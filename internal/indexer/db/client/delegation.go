package indexerdbclient

import (
	"context"
	"errors"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (indexerdbclient *IndexerDatabase) GetDelegation(ctx context.Context, stakingTxHashHex string) (*indexerdbmodel.IndexerDelegationDetails, error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).Collection(indexerdbmodel.BTCDelegationDetailsCollection)
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
	ctx context.Context, stakerPKHex string, paginationToken string,
) (*db.DbResultMap[indexerdbmodel.IndexerDelegationDetails], error) {
	client := indexerdbclient.Client.Database(indexerdbclient.DbName).Collection(indexerdbmodel.BTCDelegationDetailsCollection)

	// Base filter with stakingTxHashHex
	filter := bson.M{"staker_btc_pk_hex": stakerPKHex}

	// Default sort by start_height for stable sorting
	options := options.Find().SetSort(bson.D{
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

		filter = bson.M{
			"$or": []bson.M{
				{
					"staker_btc_pk_hex":                       stakerPKHex,
					"btc_delegation_created_bbn_block.height": bson.M{"$lt": decodedToken.StartHeight},
				},
				{
					"staker_btc_pk_hex":                       stakerPKHex,
					"btc_delegation_created_bbn_block.height": decodedToken.StartHeight,
					"_id": bson.M{"$gt": decodedToken.StakingTxHashHex},
				},
			},
		}
	}

	return db.FindWithPagination(
		ctx, client, filter, options, indexerdbclient.Cfg.MaxPaginationLimit,
		indexerdbmodel.BuildDelegationPaginationToken,
	)
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
