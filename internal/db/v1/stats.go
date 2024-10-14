package v1db

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/babylonlabs-io/staking-api-service/internal/db"
	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	v1model "github.com/babylonlabs-io/staking-api-service/internal/db/model/v1"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetOrCreateStatsLock fetches the lock status for each stats type for the given staking tx hash.
// If the document does not exist, it will create a new document with the default values
// Refer to the README.md in this directory for more information on the stats lock
func (db *V1Database) GetOrCreateStatsLock(
	ctx context.Context, stakingTxHashHex string, txType string,
) (*v1model.StatsLockDocument, error) {
	client := db.Client.Database(db.DbName).Collection(model.StatsLockCollection)
	id := constructStatsLockId(stakingTxHashHex, txType)
	filter := bson.M{"_id": id}
	// Define the default document to be inserted if not found
	// This setOnInsert will only be applied if the document is not found
	update := bson.M{
		"$setOnInsert": v1model.NewStatsLockDocument(
			id,
			false,
			false,
			false,
		),
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result v1model.StatsLockDocument
	err := client.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// IncrementOverallStats increments the overall stats for the given staking tx hash.
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
// Refer to the README.md in this directory for more information on the sharding logic
func (v1db *V1Database) IncrementOverallStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	overallStatsClient := v1db.Client.Database(v1db.DbName).Collection(model.OverallStatsCollection)
	stakerStatsClient := v1db.Client.Database(v1db.DbName).Collection(model.StakerStatsCollection)

	// Start a session
	session, sessionErr := v1db.Client.StartSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         int64(amount),
			"total_tvl":          int64(amount),
			"active_delegations": 1,
			"total_delegations":  1,
		},
	}
	// Define the work to be done in the transaction
	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := v1db.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, types.Active.ToString(), "overall_stats")
		if err != nil {
			return nil, err
		}

		// The order of the overall stats and staker stats update is important.
		// The staker stats colleciton will need to be processed first to determine if the staker is new
		// If the staker stats is the first delegation for the staker, we need to increment the total stakers
		var stakerStats v1model.StakerStatsDocument
		stakerStatsFilter := bson.M{"_id": stakerPkHex}
		stakerErr := stakerStatsClient.FindOne(ctx, stakerStatsFilter).Decode(&stakerStats)
		if stakerErr != nil {
			return nil, stakerErr
		}
		if stakerStats.TotalDelegations == 1 {
			upsertUpdate["$inc"].(bson.M)["total_stakers"] = 1
		}
		shardId, err := v1db.generateOverallStatsId()
		if err != nil {
			return nil, err
		}

		upsertFilter := bson.M{"_id": shardId}

		_, err = overallStatsClient.UpdateOne(sessCtx, upsertFilter, upsertUpdate, options.Update().SetUpsert(true))
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	// Execute the transaction
	_, txErr := session.WithTransaction(ctx, transactionWork)
	if txErr != nil {
		return txErr
	}

	return nil
}

// SubtractOverallStats decrements the overall stats for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
// Refer to the README.md in this directory for more information on the sharding logic
func (v1db *V1Database) SubtractOverallStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         -int64(amount),
			"active_delegations": -1,
		},
	}
	overallStatsClient := v1db.Client.Database(v1db.DbName).Collection(model.OverallStatsCollection)

	// Start a session
	session, sessionErr := v1db.Client.StartSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	// Define the work to be done in the transaction
	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := v1db.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, types.Unbonded.ToString(), "overall_stats")
		if err != nil {
			return nil, err
		}
		shardId, err := v1db.generateOverallStatsId()
		if err != nil {
			return nil, err
		}

		upsertFilter := bson.M{"_id": shardId}

		_, err = overallStatsClient.UpdateOne(sessCtx, upsertFilter, upsertUpdate, options.Update().SetUpsert(true))
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	// Execute the transaction
	_, txErr := session.WithTransaction(ctx, transactionWork)
	if txErr != nil {
		return txErr
	}

	return nil
}

// GetOverallStats fetches the overall stats from all the shards and sums them up
// Refer to the README.md in this directory for more information on the sharding logic
func (v1db *V1Database) GetOverallStats(ctx context.Context) (*v1model.OverallStatsDocument, error) {
	// The collection is sharded by the _id field, so we need to query all the shards
	var shardsId []string
	for i := 0; i < int(v1db.Cfg.LogicalShardCount); i++ {
		shardsId = append(shardsId, fmt.Sprintf("%d", i))
	}

	client := v1db.Client.Database(v1db.DbName).Collection(model.OverallStatsCollection)
	filter := bson.M{"_id": bson.M{"$in": shardsId}}
	cursor, err := client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var overallStats []v1model.OverallStatsDocument
	if err = cursor.All(ctx, &overallStats); err != nil {
		return nil, err
	}

	// Sum up the stats for the overall stats
	var result v1model.OverallStatsDocument
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
func (v1db *V1Database) generateOverallStatsId() (string, error) {
	max := big.NewInt(int64(v1db.Cfg.LogicalShardCount))
	// Generate a secure random number within the range [0, max)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(n), nil
}

func (v1db *V1Database) updateStatsLockByFieldName(ctx context.Context, stakingTxHashHex, state string, fieldName string) error {
	statsLockClient := v1db.Client.Database(v1db.DbName).Collection(model.StatsLockCollection)
	filter := bson.M{"_id": constructStatsLockId(stakingTxHashHex, state), fieldName: false}
	update := bson.M{"$set": bson.M{fieldName: true}}
	result, err := statsLockClient.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return &db.NotFoundError{
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
func (v1db *V1Database) IncrementFinalityProviderStats(
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
	return v1db.updateFinalityProviderStats(ctx, types.Active.ToString(), stakingTxHashHex, fpPkHex, upsertUpdate)
}

// SubtractFinalityProviderStats decrements the finality provider stats for the given provider pk hex
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
// Refer to the README.md in this directory for more information on the sharding logic
func (v1db *V1Database) SubtractFinalityProviderStats(
	ctx context.Context, stakingTxHashHex, fpPkHex string, amount uint64,
) error {
	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         -int64(amount),
			"active_delegations": -1,
		},
	}
	return v1db.updateFinalityProviderStats(ctx, types.Unbonded.ToString(), stakingTxHashHex, fpPkHex, upsertUpdate)
}

// FindFinalityProviderStats fetches the finality provider stats from the database
func (v1db *V1Database) FindFinalityProviderStats(ctx context.Context, paginationToken string) (*db.DbResultMap[*v1model.FinalityProviderStatsDocument], error) {
	client := v1db.Client.Database(v1db.DbName).Collection(model.FinalityProviderStatsCollection)
	options := options.Find().SetSort(bson.D{{Key: "active_tvl", Value: -1}}) // Sorting in descending order
	var filter bson.M

	// Decode the pagination token first if it exist
	if paginationToken != "" {
		decodedToken, err := model.DecodePaginationToken[v1model.FinalityProviderStatsPagination](paginationToken)
		if err != nil {
			return nil, &db.InvalidPaginationTokenError{
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

	return db.FindWithPagination(
		ctx, client, filter, options, v1db.Cfg.MaxPaginationLimit,
		v1model.BuildFinalityProviderStatsPaginationToken,
	)
}

func (v1db *V1Database) FindFinalityProviderStatsByFinalityProviderPkHex(
	ctx context.Context, finalityProviderPkHex []string,
) ([]*v1model.FinalityProviderStatsDocument, error) {
	client := v1db.Client.Database(v1db.DbName).Collection(model.FinalityProviderStatsCollection)
	filter := bson.M{"_id": bson.M{"$in": finalityProviderPkHex}}
	cursor, err := client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var finalityProviders []*v1model.FinalityProviderStatsDocument
	if err = cursor.All(ctx, &finalityProviders); err != nil {
		return nil, err
	}

	return finalityProviders, nil
}

func (v1db *V1Database) updateFinalityProviderStats(ctx context.Context, state, stakingTxHashHex, fpPkHex string, upsertUpdate primitive.M) error {
	client := v1db.Client.Database(v1db.DbName).Collection(model.FinalityProviderStatsCollection)

	// Start a session
	session, sessionErr := v1db.Client.StartSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := v1db.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, state, "finality_provider_stats")
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

	// Execute the transaction
	_, txErr := session.WithTransaction(ctx, transactionWork)
	if txErr != nil {
		return txErr
	}

	return nil
}

// IncrementStakerStats increments the staker stats for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v1db *V1Database) IncrementStakerStats(
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
	return v1db.updateStakerStats(ctx, types.Active.ToString(), stakingTxHashHex, stakerPkHex, upsertUpdate)
}

// SubtractStakerStats decrements the staker stats for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v1db *V1Database) SubtractStakerStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         -int64(amount),
			"active_delegations": -1,
		},
	}
	return v1db.updateStakerStats(ctx, types.Unbonded.ToString(), stakingTxHashHex, stakerPkHex, upsertUpdate)
}

func (v1db *V1Database) updateStakerStats(ctx context.Context, state, stakingTxHashHex, stakerPkHex string, upsertUpdate primitive.M) error {
	client := v1db.Client.Database(v1db.DbName).Collection(model.StakerStatsCollection)

	// Start a session
	session, sessionErr := v1db.Client.StartSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := v1db.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, state, "staker_stats")
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

	// Execute the transaction
	_, txErr := session.WithTransaction(ctx, transactionWork)
	return txErr
}

func (v1db *V1Database) FindTopStakersByTvl(ctx context.Context, paginationToken string) (*db.DbResultMap[*v1model.StakerStatsDocument], error) {
	client := v1db.Client.Database(v1db.DbName).Collection(model.StakerStatsCollection)

	opts := options.Find().SetSort(bson.D{{Key: "active_tvl", Value: -1}})
	var filter bson.M
	// Decode the pagination token first if it exist
	if paginationToken != "" {
		decodedToken, err := model.DecodePaginationToken[v1model.StakerStatsByStakerPagination](paginationToken)
		if err != nil {
			return nil, &db.InvalidPaginationTokenError{
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

	return db.FindWithPagination(
		ctx, client, filter, opts, v1db.Cfg.MaxPaginationLimit,
		v1model.BuildStakerStatsByStakerPaginationToken,
	)
}

func (v1db *V1Database) GetStakerStats(
	ctx context.Context, stakerPkHex string,
) (*v1model.StakerStatsDocument, error) {
	client := v1db.Client.Database(v1db.DbName).Collection(model.StakerStatsCollection)
	filter := bson.M{"_id": stakerPkHex}
	var result v1model.StakerStatsDocument
	err := client.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		// If the document is not found, return nil
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &db.NotFoundError{
				Key:     stakerPkHex,
				Message: "Staker stats not found",
			}
		}
		return nil, err
	}
	return &result, nil
}
