package db

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/babylonchain/staking-api-service/internal/db/model"
	"github.com/babylonchain/staking-api-service/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetOrCreateStatsLock fetches the lock status for each stats type for the given staking tx hash.
// If the document does not exist, it will create a new document with the default values
// Refer to the README.md in this directory for more information on the stats lock
func (db *Database) GetOrCreateStatsLock(
	ctx context.Context, stakingTxHashHex string, txType string,
) (*model.StatsLockDocument, error) {
	client := db.Client.Database(db.DbName).Collection(model.StatsLockCollection)
	id := constructStatsLockId(stakingTxHashHex, txType)
	filter := bson.M{"_id": id}
	// Define the default document to be inserted if not found
	// This setOnInsert will only be applied if the document is not found
	update := bson.M{
		"$setOnInsert": model.NewStatsLockDocument(
			id,
			false,
			false,
			false,
		),
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result model.StatsLockDocument
	err := client.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// IncrementOverallStats increments the overall stats for the given staking tx hash.
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
// Refer to the README.md in this directory for more information on the sharding logic
func (db *Database) IncrementOverallStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	// Define the work to be done in the transaction
  transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
 	overallStatsClient := db.Client.Database(db.DbName).Collection(model.OverallStatsCollection)
  stakerStatsClient := db.Client.Database(db.DbName).Collection(model.StakerStatsCollection)

	err := db.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, types.Active.ToString(), "overall_stats")
	if err != nil {
		return nil, err
	}

	// The order of the overall stats and staker stats update is important.
	// The staker stats colleciton will need to be processed first to determine if the staker is new
	// If the staker stats is the first delegation for the staker, we need to increment the total stakers
	var stakerStats model.StakerStatsDocument
	stakerStatsFilter := bson.M{"_id": stakerPkHex}
	stakerErr := stakerStatsClient.FindOne(ctx, stakerStatsFilter).Decode(&stakerStats)
	if stakerErr != nil {
		return nil, stakerErr
	}

	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         int64(amount),
			"total_tvl":          int64(amount),
			"active_delegations": 1,
			"total_delegations":  1,
		},
	}

	if stakerStats.TotalDelegations == 1 {
			upsertUpdate["$inc"].(bson.M)["total_stakers"] = 1
	}

	upsertFilter := bson.M{"_id": db.generateOverallStatsId()}

	_, err = overallStatsClient.UpdateOne(sessCtx, upsertFilter, upsertUpdate, options.Update().SetUpsert(true))
	if err != nil {
		return nil, err
	}
	return nil, nil
	}

	// Execute the transaction with retries
	_, txErr := TxWithRetries(
		ctx, 
		&dbTransactionClient{db.Client},
		transactionWork,
	)
	if txErr != nil {
		return txErr
	}

	return nil
}

// SubtractOverallStats decrements the overall stats for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
// Refer to the README.md in this directory for more information on the sharding logic
func (db *Database) SubtractOverallStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	// Define the work to be done in the transaction
	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		upsertUpdate := bson.M{
			"$inc": bson.M{
				"active_tvl":         -int64(amount),
				"active_delegations": -1,
			},
		}
		overallStatsClient := db.Client.Database(db.DbName).Collection(model.OverallStatsCollection)
		err := db.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, types.Unbonded.ToString(), "overall_stats")
		if err != nil {
			return nil, err
		}

		upsertFilter := bson.M{"_id": db.generateOverallStatsId()}

		_, err = overallStatsClient.UpdateOne(sessCtx, upsertFilter, upsertUpdate, options.Update().SetUpsert(true))
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	// Execute the transaction with retries
	_, txErr := TxWithRetries(
		ctx, 
		&dbTransactionClient{db.Client},
		transactionWork,
	)
	if txErr != nil {
		return txErr
	}

	return nil
}

// GetOverallStats fetches the overall stats from all the shards and sums them up
// Refer to the README.md in this directory for more information on the sharding logic
func (db *Database) GetOverallStats(ctx context.Context) (*model.OverallStatsDocument, error) {
	// The collection is sharded by the _id field, so we need to query all the shards
	var shardsId []string
	for i := 0; i < int(db.cfg.LogicalShardCount); i++ {
		shardsId = append(shardsId, fmt.Sprintf("%d", i))
	}

	client := db.Client.Database(db.DbName).Collection(model.OverallStatsCollection)
	filter := bson.M{"_id": bson.M{"$in": shardsId}}
	cursor, err := client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var overallStats []model.OverallStatsDocument
	if err = cursor.All(ctx, &overallStats); err != nil {
		return nil, err
	}

	// Sum up the stats for the overall stats
	var result model.OverallStatsDocument
	for _, stats := range overallStats {
		result.ActiveTvl += stats.ActiveTvl
		result.TotalTvl += stats.TotalTvl
		result.ActiveDelegations += stats.ActiveDelegations
		result.TotalDelegations += stats.TotalDelegations
		result.TotalStakers += stats.TotalStakers
	}

	return &result, nil
}

// Generate the id for the overall stats document. Id is a random number ranged from 0-LogicalShardCount-1
// It's a logical shard to avoid locking the same field during concurrent writes
// The sharding number should never be reduced after roll out
func (db *Database) generateOverallStatsId() string {
	return fmt.Sprint(rand.Intn(int(db.cfg.LogicalShardCount)))
}

func (db *Database) updateStatsLockByFieldName(ctx context.Context, stakingTxHashHex, state string, fieldName string) error {
	statsLockClient := db.Client.Database(db.DbName).Collection(model.StatsLockCollection)
	filter := bson.M{"_id": constructStatsLockId(stakingTxHashHex, state), fieldName: false}
	update := bson.M{"$set": bson.M{fieldName: true}}
	result, err := statsLockClient.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return &NotFoundError{
			Key:     stakingTxHashHex,
			Message: "document already processed or does not exist",
		}
	}
	return nil
}

func constructStatsLockId(stakingTxHashHex, state string) string {
	return stakingTxHashHex + ":" + state
}

// IncrementFinalityProviderStats increments the finality provider stats for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
// Refer to the README.md in this directory for more information on the sharding logic
func (db *Database) IncrementFinalityProviderStats(
	ctx context.Context, stakingTxHashHex, fpPkHex string, amount uint64,
) error {
	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         int64(amount),
			"total_tvl":          int64(amount),
			"active_delegations": 1,
			"total_delegations":  1,
		},
	}
	return db.updateFinalityProviderStats(ctx, types.Active.ToString(), stakingTxHashHex, fpPkHex, upsertUpdate)
}

// SubtractFinalityProviderStats decrements the finality provider stats for the given provider pk hex
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
// Refer to the README.md in this directory for more information on the sharding logic
func (db *Database) SubtractFinalityProviderStats(
	ctx context.Context, stakingTxHashHex, fpPkHex string, amount uint64,
) error {
	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         -int64(amount),
			"active_delegations": -1,
		},
	}
	return db.updateFinalityProviderStats(ctx, types.Unbonded.ToString(), stakingTxHashHex, fpPkHex, upsertUpdate)
}

// FindFinalityProviderStats fetches the finality provider stats from the database
func (db *Database) FindFinalityProviderStats(ctx context.Context, paginationToken string) (*DbResultMap[*model.FinalityProviderStatsDocument], error) {
	client := db.Client.Database(db.DbName).Collection(model.FinalityProviderStatsCollection)
	options := options.Find().SetSort(bson.D{{Key: "active_tvl", Value: -1}}) // Sorting in descending order
	var filter bson.M

	// Decode the pagination token first if it exist
	if paginationToken != "" {
		decodedToken, err := model.DecodePaginationToken[model.FinalityProviderStatsPagination](paginationToken)
		if err != nil {
			return nil, &InvalidPaginationTokenError{
				Message: "Invalid pagination token",
			}
		}
		filter = bson.M{
			"$or": []bson.M{
				{"active_tvl": bson.M{"$lt": decodedToken.ActiveTvl}},
				{"active_tvl": decodedToken.ActiveTvl, "_id": bson.M{"$lt": decodedToken.FinalityProviderPkHex}},
			},
		}
	}

	return findWithPagination(
		ctx, client, filter, options, db.cfg.MaxPaginationLimit,
		model.BuildFinalityProviderStatsPaginationToken,
	)
}

func (db *Database) FindFinalityProviderStatsByFinalityProviderPkHex(
	ctx context.Context, finalityProviderPkHex []string,
) ([]*model.FinalityProviderStatsDocument, error) {
	client := db.Client.Database(db.DbName).Collection(model.FinalityProviderStatsCollection)
	filter := bson.M{"_id": bson.M{"$in": finalityProviderPkHex}}
	cursor, err := client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var finalityProviders []*model.FinalityProviderStatsDocument
	if err = cursor.All(ctx, &finalityProviders); err != nil {
		return nil, err
	}

	return finalityProviders, nil
}

func (db *Database) updateFinalityProviderStats(ctx context.Context, state, stakingTxHashHex, fpPkHex string, upsertUpdate primitive.M) error {
	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		client := db.Client.Database(db.DbName).Collection(model.FinalityProviderStatsCollection)

		err := db.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, state, "finality_provider_stats")
		if err != nil {
			return nil, err
		}

		upsertFilter := bson.M{"_id": fpPkHex}

		_, err = client.UpdateOne(sessCtx, upsertFilter, upsertUpdate, options.Update().SetUpsert(true))
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	// Execute the transaction with retries
	_, txErr := TxWithRetries(
		ctx, 
		&dbTransactionClient{db.Client},
		transactionWork,
	)
	if txErr != nil {
		return txErr
	}

	return nil
}

// IncrementStakerStats increments the staker stats for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (db *Database) IncrementStakerStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         int64(amount),
			"total_tvl":          int64(amount),
			"active_delegations": 1,
			"total_delegations":  1,
		},
	}
	return db.updateStakerStats(ctx, types.Active.ToString(), stakingTxHashHex, stakerPkHex, upsertUpdate)
}

// SubtractStakerStats decrements the staker stats for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (db *Database) SubtractStakerStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         -int64(amount),
			"active_delegations": -1,
		},
	}
	return db.updateStakerStats(ctx, types.Unbonded.ToString(), stakingTxHashHex, stakerPkHex, upsertUpdate)
}

func (db *Database) updateStakerStats(ctx context.Context, state, stakingTxHashHex, stakerPkHex string, upsertUpdate primitive.M) error {	
	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		client := db.Client.Database(db.DbName).Collection(model.StakerStatsCollection)

		err := db.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, state, "staker_stats")
		if err != nil {
			return nil, err
		}

		upsertFilter := bson.M{"_id": stakerPkHex}

		_, err = client.UpdateOne(sessCtx, upsertFilter, upsertUpdate, options.Update().SetUpsert(true))
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	// Execute the transaction with retries
	_, txErr := TxWithRetries(
		ctx, 
		&dbTransactionClient{db.Client},
		transactionWork,
	)
	return txErr
}

func (db *Database) FindTopStakersByTvl(ctx context.Context, paginationToken string) (*DbResultMap[*model.StakerStatsDocument], error) {
	client := db.Client.Database(db.DbName).Collection(model.StakerStatsCollection)

	opts := options.Find().SetSort(bson.D{{Key: "active_tvl", Value: -1}})
	var filter bson.M
	// Decode the pagination token first if it exist
	if paginationToken != "" {
		decodedToken, err := model.DecodePaginationToken[model.StakerStatsByStakerPagination](paginationToken)
		if err != nil {
			return nil, &InvalidPaginationTokenError{
				Message: "Invalid pagination token",
			}
		}
		filter = bson.M{
			"$or": []bson.M{
				{"active_tvl": bson.M{"$lt": decodedToken.ActiveTvl}},
				{"active_tvl": decodedToken.ActiveTvl, "_id": bson.M{"$lt": decodedToken.StakerPkHex}},
			},
		}
	}

	return findWithPagination(
		ctx, client, filter, opts, db.cfg.MaxPaginationLimit,
		model.BuildStakerStatsByStakerPaginationToken,
	)
}
