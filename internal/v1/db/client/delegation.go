package v1dbclient

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	v1dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
)

func (v1dbclient *V1Database) SaveActiveStakingDelegation(
	ctx context.Context, stakingTxHashHex, stakerPkHex, fpPkHex string,
	stakingTxHex string, amount, startHeight, timelock, outputIndex uint64,
	startTimestamp int64, isOverflow bool,
) error {
	client := v1dbclient.Client.Database(v1dbclient.DbName).Collection(dbmodel.V1DelegationCollection)
	document := v1dbmodel.DelegationDocument{
		StakingTxHashHex:      stakingTxHashHex, // Primary key of db collection
		StakerPkHex:           stakerPkHex,
		FinalityProviderPkHex: fpPkHex,
		StakingValue:          amount,
		State:                 types.Active,
		StakingTx: &v1dbmodel.TimelockTransaction{
			TxHex:          stakingTxHex,
			OutputIndex:    outputIndex,
			StartTimestamp: startTimestamp,
			StartHeight:    startHeight,
			TimeLock:       timelock,
		},
		IsOverflow: isOverflow,
	}
	_, err := client.InsertOne(ctx, document)
	if err != nil {
		var writeErr mongo.WriteException
		if errors.As(err, &writeErr) {
			for _, e := range writeErr.WriteErrors {
				if mongo.IsDuplicateKeyError(e) {
					// Return the custom error type so that we can return 4xx errors to client
					return &db.DuplicateKeyError{
						Key:     stakingTxHashHex,
						Message: "Delegation already exists",
					}
				}
			}
		}
		return err
	}
	return nil
}

// CheckDelegationExistByStakerPk checks if a staker has any
// delegation in the specified states by the staker's public key
func (v1dbclient *V1Database) CheckDelegationExistByStakerPk(
	ctx context.Context, stakerPk string, extraFilter *DelegationFilter,
) (bool, error) {
	client := v1dbclient.Client.Database(v1dbclient.DbName).Collection(dbmodel.V1DelegationCollection)
	filter := buildAdditionalDelegationFilter(
		bson.M{"staker_pk_hex": stakerPk}, extraFilter,
	)
	var delegation v1dbmodel.DelegationDocument
	err := client.FindOne(ctx, filter).Decode(&delegation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (v1dbclient *V1Database) FindDelegationsByStakerPk(
	ctx context.Context, stakerPk string,
	extraFilter *DelegationFilter, paginationToken string,
) (*db.DbResultMap[v1dbmodel.DelegationDocument], error) {
	client := v1dbclient.Client.Database(v1dbclient.DbName).Collection(dbmodel.V1DelegationCollection)

	filter := bson.M{"staker_pk_hex": stakerPk}
	filter = buildAdditionalDelegationFilter(filter, extraFilter)
	options := options.Find().SetSort(bson.D{
		{Key: "staking_tx.start_height", Value: -1},
		{Key: "_id", Value: 1},
	})

	// Decode the pagination token first if it exist
	if paginationToken != "" {
		decodedToken, err := dbmodel.DecodePaginationToken[v1dbmodel.DelegationByStakerPagination](paginationToken)
		if err != nil {
			return nil, &db.InvalidPaginationTokenError{
				Message: "Invalid pagination token",
			}
		}
		filter = bson.M{
			"$or": []bson.M{
				{"staker_pk_hex": stakerPk, "staking_tx.start_height": bson.M{"$lt": decodedToken.StakingStartHeight}},
				{"staker_pk_hex": stakerPk, "staking_tx.start_height": decodedToken.StakingStartHeight, "_id": bson.M{"$gt": decodedToken.StakingTxHashHex}},
			},
		}
	}

	return db.FindWithPagination(
		ctx, client, filter, options, v1dbclient.Cfg.MaxPaginationLimit,
		v1dbmodel.BuildDelegationByStakerPaginationToken,
	)
}

// SaveUnbondingTx saves the unbonding transaction details for a staking transaction
// It returns an NotFoundError if the staking transaction is not found
func (v1dbclient *V1Database) FindDelegationByTxHashHex(ctx context.Context, stakingTxHashHex string) (*v1dbmodel.DelegationDocument, error) {
	client := v1dbclient.Client.Database(v1dbclient.DbName).Collection(dbmodel.V1DelegationCollection)
	filter := bson.M{"_id": stakingTxHashHex}
	var delegation v1dbmodel.DelegationDocument
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

func (v1dbclient *V1Database) ScanDelegationsPaginated(
	ctx context.Context,
	paginationToken string,
) (*db.DbResultMap[v1dbmodel.DelegationDocument], error) {
	client := v1dbclient.Client.Database(v1dbclient.DbName).Collection(dbmodel.V1DelegationCollection)
	filter := bson.M{}
	options := options.Find()
	options.SetSort(bson.M{"_id": 1})
	// Decode the pagination token if it exists
	if paginationToken != "" {
		decodedToken, err :=
			dbmodel.DecodePaginationToken[v1dbmodel.DelegationScanPagination](paginationToken)
		if err != nil {
			return nil, &db.InvalidPaginationTokenError{
				Message: "Invalid pagination token",
			}
		}
		filter["_id"] = bson.M{"$gt": decodedToken.StakingTxHashHex}
	}

	// Perform the paginated query and return the results
	return db.FindWithPagination(
		ctx, client, filter, options, v1dbclient.Cfg.MaxPaginationLimit,
		v1dbmodel.BuildDelegationScanPaginationToken,
	)
}

// TransitionToTransitionedState marks an existing delegation as transitioned
func (v1dbclient *V1Database) TransitionToTransitionedState(
	ctx context.Context, stakingTxHashHex string,
) error {
	return v1dbclient.transitionState(
		ctx, stakingTxHashHex, types.Transitioned.ToString(),
		utils.QualifiedStatesToTransitioned(), nil,
	)
}

// TransitionState updates the state of a staking transaction to a new state
// It returns an NotFoundError if the staking transaction is not found or not in the eligible state to transition
func (v1dbclient *V1Database) transitionState(
	ctx context.Context, stakingTxHashHex, newState string,
	eligiblePreviousState []types.DelegationState, additionalUpdates map[string]interface{},
) error {
	client := v1dbclient.Client.Database(v1dbclient.DbName).Collection(dbmodel.V1DelegationCollection)
	filter := bson.M{"_id": stakingTxHashHex, "state": bson.M{"$in": eligiblePreviousState}}
	update := bson.M{"$set": bson.M{"state": newState}}
	for field, value := range additionalUpdates {
		// Add additional fields to the $set operation
		update["$set"].(bson.M)[field] = value
	}
	_, err := client.UpdateOne(ctx, filter, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &db.NotFoundError{
				Key:     stakingTxHashHex,
				Message: "Delegation not found or not in eligible state to transition",
			}
		}
		return err
	}
	return nil
}

func buildAdditionalDelegationFilter(
	baseFilter primitive.M,
	filters *DelegationFilter,
) primitive.M {
	if filters != nil {
		if filters.States != nil {
			baseFilter["state"] = bson.M{"$in": filters.States}
		}
		if filters.AfterTimestamp != 0 {
			baseFilter["staking_tx.start_timestamp"] = bson.M{"$gte": filters.AfterTimestamp}
		}
	}
	return baseFilter
}
