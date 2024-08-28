package db

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

func (db *Database) SaveActiveStakingDelegation(
	ctx context.Context, stakingTxHashHex, stakerPkHex, fpPkHex string,
	stakingTxHex string, amount, startHeight, timelock, outputIndex uint64,
	startTimestamp int64, isOverflow bool, stakerTaprootAddress string,
) error {
	client := db.Client.Database(db.DbName).Collection(model.DelegationCollection)
	document := model.DelegationDocument{
		StakingTxHashHex:      stakingTxHashHex, // Primary key of db collection
		StakerPkHex:           stakerPkHex,
		FinalityProviderPkHex: fpPkHex,
		StakingValue:          amount,
		State:                 types.Active,
		StakingTx: &model.TimelockTransaction{
			TxHex:          stakingTxHex,
			OutputIndex:    outputIndex,
			StartTimestamp: startTimestamp,
			StartHeight:    startHeight,
			TimeLock:       timelock,
		},
		IsOverflow: isOverflow,
		StakerBtcAddress: &model.StakerBtcAddress{
			TaprootAddress: stakerTaprootAddress,
		},
	}
	_, err := client.InsertOne(ctx, document)
	if err != nil {
		var writeErr mongo.WriteException
		if errors.As(err, &writeErr) {
			for _, e := range writeErr.WriteErrors {
				if mongo.IsDuplicateKeyError(e) {
					// Return the custom error type so that we can return 4xx errors to client
					return &DuplicateKeyError{
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

// CheckDelegationExistByStakerTaprootAddress checks if a staker has any
// delegation in the specified states by the staker's BTC address in taproot format.
func (db *Database) CheckDelegationExistByStakerTaprootAddress(
	ctx context.Context, address string, extraFilter *DelegationFilter,
) (bool, error) {
	client := db.Client.Database(db.DbName).Collection(model.DelegationCollection)
	filter := buildAdditionalDelegationFilter(
		bson.M{"staker_btc_address.taproot_address": address}, extraFilter,
	)
	var delegation model.DelegationDocument
	err := client.FindOne(ctx, filter).Decode(&delegation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (db *Database) FindDelegationsByStakerPk(ctx context.Context, stakerPk string, paginationToken string) (*DbResultMap[model.DelegationDocument], error) {
	client := db.Client.Database(db.DbName).Collection(model.DelegationCollection)

	filter := bson.M{"staker_pk_hex": stakerPk}
	options := options.Find().SetSort(bson.D{
		{Key: "staking_tx.start_height", Value: -1},
		{Key: "_id", Value: 1},
	})

	// Decode the pagination token first if it exist
	if paginationToken != "" {
		decodedToken, err := model.DecodePaginationToken[model.DelegationByStakerPagination](paginationToken)
		if err != nil {
			return nil, &InvalidPaginationTokenError{
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

	return findWithPagination(
		ctx, client, filter, options, db.cfg.MaxPaginationLimit,
		model.BuildDelegationByStakerPaginationToken,
	)
}

// SaveUnbondingTx saves the unbonding transaction details for a staking transaction
// It returns an NotFoundError if the staking transaction is not found
func (db *Database) FindDelegationByTxHashHex(ctx context.Context, stakingTxHashHex string) (*model.DelegationDocument, error) {
	client := db.Client.Database(db.DbName).Collection(model.DelegationCollection)
	filter := bson.M{"_id": stakingTxHashHex}
	var delegation model.DelegationDocument
	err := client.FindOne(ctx, filter).Decode(&delegation)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &NotFoundError{
				Key:     stakingTxHashHex,
				Message: "Delegation not found",
			}
		}
		return nil, err
	}
	return &delegation, nil
}

func (db *Database) ScanDelegationsPaginated(
	ctx context.Context,
	paginationToken string,
) (*DbResultMap[model.DelegationDocument], error) {
	client := db.Client.Database(db.DbName).Collection(model.DelegationCollection)
	filter := bson.M{}
	options := options.Find()
	options.SetSort(bson.M{"_id": 1})
	// Decode the pagination token if it exists
	if paginationToken != "" {
		decodedToken, err :=
			model.DecodePaginationToken[model.DelegationScanPagination](paginationToken)
		if err != nil {
			return nil, &InvalidPaginationTokenError{
				Message: "Invalid pagination token",
			}
		}
		filter["_id"] = bson.M{"$gt": decodedToken.StakingTxHashHex}
	}

	// Perform the paginated query and return the results
	return findWithPagination(
		ctx, client, filter, options, db.cfg.MaxPaginationLimit,
		model.BuildDelegationScanPaginationToken,
	)
}

// TransitionState updates the state of a staking transaction to a new state
// It returns an NotFoundError if the staking transaction is not found or not in the eligible state to transition
func (db *Database) transitionState(
	ctx context.Context, stakingTxHashHex, newState string,
	eligiblePreviousState []types.DelegationState, additionalUpdates map[string]interface{},
) error {
	client := db.Client.Database(db.DbName).Collection(model.DelegationCollection)
	filter := bson.M{"_id": stakingTxHashHex, "state": bson.M{"$in": eligiblePreviousState}}
	update := bson.M{"$set": bson.M{"state": newState}}
	for field, value := range additionalUpdates {
		// Add additional fields to the $set operation
		update["$set"].(bson.M)[field] = value
	}
	_, err := client.UpdateOne(ctx, filter, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &NotFoundError{
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
	if filters.States != nil {
		baseFilter["state"] = bson.M{"$in": filters.States}
	}
	if filters.AfterTimestamp != 0 {
		baseFilter["staking_tx.start_timestamp"] = bson.M{"$gte": filters.AfterTimestamp}
	}
	return baseFilter
}
