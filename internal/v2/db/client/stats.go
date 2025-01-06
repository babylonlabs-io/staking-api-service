package v2dbclient

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetOrCreateStatsLock fetches the lock status for each stats type for the given staking tx hash.
// If the document does not exist, it will create a new document with the default values
func (db *V2Database) GetOrCreateStatsLock(
	ctx context.Context, stakingTxHashHex string, txType string,
) (*v2dbmodel.V2StatsLockDocument, error) {
	client := db.Client.Database(db.DbName).Collection(dbmodel.V2StatsLockCollection)
	id := constructStatsLockId(stakingTxHashHex, txType)
	filter := bson.M{"_id": id}
	// Define the default document to be inserted if not found
	// This setOnInsert will only be applied if the document is not found
	update := bson.M{
		"$setOnInsert": v2dbmodel.NewV2StatsLockDocument(
			id,
			false,
			false,
			false,
		),
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result v2dbmodel.V2StatsLockDocument
	err := client.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// IncrementOverallStats increments the overall stats for the given staking tx hash.
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v2dbclient *V2Database) IncrementOverallStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	overallStatsClient := v2dbclient.Client.Database(v2dbclient.DbName).Collection(dbmodel.V2OverallStatsCollection)

	// Start a session
	session, sessionErr := v2dbclient.Client.StartSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         int64(amount),
			"active_delegations": 1,
		},
	}
	// Define the work to be done in the transaction
	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := v2dbclient.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, types.Active.ToString(), "overall_stats")
		if err != nil {
			return nil, err
		}

		shardId, err := v2dbclient.generateOverallStatsId()
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
func (v2dbclient *V2Database) SubtractOverallStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         -int64(amount),
			"active_delegations": -1,
		},
	}
	overallStatsClient := v2dbclient.Client.Database(v2dbclient.DbName).Collection(dbmodel.V2OverallStatsCollection)

	// Start a session
	session, sessionErr := v2dbclient.Client.StartSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	// Define the work to be done in the transaction
	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := v2dbclient.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, types.Unbonding.ToString(), "overall_stats")
		if err != nil {
			return nil, err
		}
		shardId, err := v2dbclient.generateOverallStatsId()
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
func (v2dbclient *V2Database) GetOverallStats(ctx context.Context) (*v2dbmodel.V2OverallStatsDocument, error) {
	// The collection is sharded by the _id field, so we need to query all the shards
	var shardsId []string
	for i := 0; i < int(*v2dbclient.Cfg.LogicalShardCount); i++ {
		shardsId = append(shardsId, fmt.Sprintf("%d", i))
	}

	client := v2dbclient.Client.Database(v2dbclient.DbName).Collection(dbmodel.V2OverallStatsCollection)
	filter := bson.M{"_id": bson.M{"$in": shardsId}}
	cursor, err := client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var overallStats []v2dbmodel.V2OverallStatsDocument
	if err = cursor.All(ctx, &overallStats); err != nil {
		return nil, err
	}

	// Sum up the stats for the overall stats
	var result v2dbmodel.V2OverallStatsDocument
	for _, stats := range overallStats {
		result.ActiveTvl += stats.ActiveTvl
		result.ActiveDelegations += stats.ActiveDelegations
	}

	return &result, nil
}

// Generate the id for the overall stats document. Id is a random number ranged from 0-LogicalShardCount-1
// It's a logical shard to avoid locking the same field during concurrent writes
// The sharding number should never be reduced after roll out
func (v2dbclient *V2Database) generateOverallStatsId() (string, error) {
	max := big.NewInt(*v2dbclient.Cfg.LogicalShardCount)
	// Generate a secure random number within the range [0, max)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(n), nil
}

func (v2dbclient *V2Database) updateStatsLockByFieldName(ctx context.Context, stakingTxHashHex, state string, fieldName string) error {
	statsLockClient := v2dbclient.Client.Database(v2dbclient.DbName).Collection(dbmodel.V2StatsLockCollection)
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

// IncrementStakerStats increments the staker stats for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v2dbclient *V2Database) IncrementStakerStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":               int64(amount),
			"active_delegations":       1,
			"withdrawable_tvl":         0,
			"withdrawable_delegations": 0,
			"withdrawn_tvl":            0,
			"withdrawn_delegations":    0,
		},
	}
	err := v2dbclient.updateStakerStats(ctx, types.Active.ToString(), stakingTxHashHex, stakerPkHex, upsertUpdate)
	if err != nil {
		return err
	}

	// Log the final stats
	stakerStats, err := v2dbclient.GetStakerStats(ctx, stakerPkHex)
	if err != nil {
		return err
	}
	log.Debug().
		Str("stakerPkHex", stakerPkHex).
		Str("stakingTxHashHex", stakingTxHashHex).
		Interface("stakerStats", stakerStats).
		Msg("Staker stats after handling active")
	return nil
}

// SubtractStakerStats decrements the staker stats for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v2dbclient *V2Database) SubtractStakerStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         -int64(amount),
			"active_delegations": -1,
		},
	}
	err := v2dbclient.updateStakerStats(ctx, types.Unbonding.ToString(), stakingTxHashHex, stakerPkHex, upsertUpdate)
	if err != nil {
		return err
	}

	// Log the final stats
	stakerStats, err := v2dbclient.GetStakerStats(ctx, stakerPkHex)
	if err != nil {
		return err
	}

	log.Debug().
		Str("stakerPkHex", stakerPkHex).
		Str("stakingTxHashHex", stakingTxHashHex).
		Interface("stakerStats", stakerStats).
		Msg("Staker stats after handling unbonding")
	return nil
}

// HandleWithdrawableStakerStats increments the withdrawable delegations count for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v2dbclient *V2Database) HandleWithdrawableStakerStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	stakerStatsClient := v2dbclient.Client.Database(v2dbclient.DbName).Collection(dbmodel.V2StakerStatsCollection)
	session, err := v2dbclient.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// 1. Create lock documents
		_, err := v2dbclient.GetOrCreateStatsLock(sessCtx, stakingTxHashHex, types.Withdrawable.ToString())
		if err != nil {
			return nil, fmt.Errorf("failed to create withdrawable lock: %w", err)
		}
		_, err = v2dbclient.GetOrCreateStatsLock(sessCtx, stakingTxHashHex, types.Unbonding.ToString())
		if err != nil {
			return nil, fmt.Errorf("failed to create unbonding lock: %w", err)
		}

		// 2. First lock the withdrawable state (prevents duplicate processing)
		if err := v2dbclient.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, types.Withdrawable.ToString(), "staker_stats"); err != nil {
			return nil, err
		}

		// 3. Try to lock unbonding state - if we can't lock it, it means it was already unbonding
		err = v2dbclient.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, types.Unbonding.ToString(), "staker_stats")
		shouldUnbond := (err == nil)

		// Common fields that will be incremented in both cases
		updateFields := bson.M{
			"withdrawable_tvl":         int64(amount),
			"withdrawable_delegations": 1,
		}

		// Add unbonding fields if needed
		if shouldUnbond {
			updateFields["active_tvl"] = -int64(amount)
			updateFields["active_delegations"] = -1
		}

		update := bson.M{
			"$inc": updateFields,
		}

		_, err = stakerStatsClient.UpdateOne(
			sessCtx,
			bson.M{"_id": stakerPkHex},
			update,
			options.Update().SetUpsert(true),
		)
		return nil, err
	}

	_, err = session.WithTransaction(ctx, transactionWork)
	if err != nil {
		return err
	}

	// Log the final stats
	stakerStats, err := v2dbclient.GetStakerStats(ctx, stakerPkHex)
	if err != nil {
		return err
	}
	log.Debug().
		Str("stakerPkHex", stakerPkHex).
		Str("stakingTxHashHex", stakingTxHashHex).
		Interface("stakerStats", stakerStats).
		Msg("Staker stats after handling withdrawable")
	return nil
}

// HandleWithdrawntakerStats decrements the withdrawable delegations count and
// increments the withdrawn delegations count for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v2dbclient *V2Database) HandleWithdrawnStakerStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	stakerStatsClient := v2dbclient.Client.Database(v2dbclient.DbName).Collection(dbmodel.V2StakerStatsCollection)
	session, err := v2dbclient.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// 1. Create lock documents
		_, err := v2dbclient.GetOrCreateStatsLock(sessCtx, stakingTxHashHex, types.Withdrawn.ToString())
		if err != nil {
			return nil, fmt.Errorf("failed to create withdrawn lock: %w", err)
		}
		_, err = v2dbclient.GetOrCreateStatsLock(sessCtx, stakingTxHashHex, types.Withdrawable.ToString())
		if err != nil {
			return nil, fmt.Errorf("failed to create withdrawable lock: %w", err)
		}
		_, err = v2dbclient.GetOrCreateStatsLock(sessCtx, stakingTxHashHex, types.Unbonding.ToString())
		if err != nil {
			return nil, fmt.Errorf("failed to create unbonding lock: %w", err)
		}

		// 2. First lock the withdrawn state (prevents duplicate processing)
		if err := v2dbclient.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, types.Withdrawn.ToString(), "staker_stats"); err != nil {
			return nil, err
		}

		// 3. Try to lock withdrawable state - if we can't lock it, it means it's already true
		err = v2dbclient.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, types.Withdrawable.ToString(), "staker_stats")
		shouldWithdrawable := (err == nil)

		// 4. Try to lock unbonding state - if we can't lock it, it means it's already true
		err = v2dbclient.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, types.Unbonding.ToString(), "staker_stats")
		shouldUnbond := (err == nil)

		// Common fields that will be incremented in both cases
		updateFields := bson.M{
			"withdrawn_tvl":         int64(amount),
			"withdrawn_delegations": 1,
		}

		// Add withdrawable fields if needed
		if shouldWithdrawable {
			updateFields["withdrawable_tvl"] = -int64(amount)
			updateFields["withdrawable_delegations"] = -1
		}

		// Add unbonding fields if needed
		if shouldUnbond {
			updateFields["active_tvl"] = -int64(amount)
			updateFields["active_delegations"] = -1
		}

		update := bson.M{
			"$inc": updateFields,
		}

		_, err = stakerStatsClient.UpdateOne(
			sessCtx,
			bson.M{"_id": stakerPkHex},
			update,
			options.Update().SetUpsert(true),
		)
		return nil, err
	}

	_, err = session.WithTransaction(ctx, transactionWork)
	if err != nil {
		return err
	}

	stakerStats, err := v2dbclient.GetStakerStats(ctx, stakerPkHex)
	if err != nil {
		return err
	}
	log.Debug().
		Str("stakerPkHex", stakerPkHex).
		Str("stakingTxHashHex", stakingTxHashHex).
		Interface("stakerStats", stakerStats).
		Msg("Staker stats after handling withdrawn")
	return nil
}

func (v2dbclient *V2Database) updateStakerStats(ctx context.Context, state, stakingTxHashHex, stakerPkHex string, upsertUpdate primitive.M) error {
	client := v2dbclient.Client.Database(v2dbclient.DbName).Collection(dbmodel.V2StakerStatsCollection)

	// Start a session
	session, sessionErr := v2dbclient.Client.StartSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := v2dbclient.updateStatsLockByFieldName(sessCtx, stakingTxHashHex, state, "staker_stats")
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

func (v2dbclient *V2Database) GetStakerStats(
	ctx context.Context, stakerPkHex string,
) (*v2dbmodel.V2StakerStatsDocument, error) {
	client := v2dbclient.Client.Database(v2dbclient.DbName).Collection(dbmodel.V2StakerStatsCollection)
	filter := bson.M{"_id": stakerPkHex}
	var result v2dbmodel.V2StakerStatsDocument
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

func (v2dbclient *V2Database) GetActiveStakersCount(ctx context.Context) (int64, error) {
	client := v2dbclient.Client.
		Database(v2dbclient.DbName).
		Collection(dbmodel.V2StakerStatsCollection)

	filter := bson.M{
		"active_delegations": bson.M{
			"$gt": 0,
		},
	}

	count, err := client.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count active stakers: %w", err)
	}

	return count, nil
}

func (v2dbclient *V2Database) IncrementFinalityProviderStats(
	ctx context.Context,
	stakingTxHashHex string,
	fpPkHexes []string,
	amount uint64,
) error {
	// Create bulk write operations for each FP
	var operations []mongo.WriteModel
	for _, fpPkHex := range fpPkHexes {
		operation := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": fpPkHex}).
			SetUpdate(bson.M{
				"$inc": bson.M{
					"active_tvl":         int64(amount),
					"active_delegations": 1,
				},
			}).
			SetUpsert(true)
		operations = append(operations, operation)
	}

	return v2dbclient.updateFinalityProviderStats(
		ctx,
		types.Active.ToString(),
		stakingTxHashHex,
		operations,
	)
}

func (v2dbclient *V2Database) SubtractFinalityProviderStats(
	ctx context.Context,
	stakingTxHashHex string,
	fpPkHexes []string,
	amount uint64,
) error {
	// Create bulk write operations for each FP
	var operations []mongo.WriteModel
	for _, fpPkHex := range fpPkHexes {
		operation := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": fpPkHex}).
			SetUpdate(bson.M{
				"$inc": bson.M{
					"active_tvl":         -int64(amount),
					"active_delegations": -1,
				},
			}).
			SetUpsert(true)
		operations = append(operations, operation)
	}

	return v2dbclient.updateFinalityProviderStats(
		ctx,
		types.Unbonding.ToString(),
		stakingTxHashHex,
		operations,
	)
}

func (v2dbclient *V2Database) updateFinalityProviderStats(
	ctx context.Context,
	state string,
	stakingTxHashHex string,
	operations []mongo.WriteModel,
) error {
	client := v2dbclient.Client.Database(v2dbclient.DbName).Collection(dbmodel.V2FinalityProviderStatsCollection)

	session, sessionErr := v2dbclient.Client.StartSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Single lock for the entire operation
		err := v2dbclient.updateStatsLockByFieldName(
			sessCtx,
			stakingTxHashHex,
			state,
			"finality_provider_stats",
		)
		if err != nil {
			return nil, err
		}

		// Execute all updates in a single bulk write
		opts := options.BulkWrite().SetOrdered(true)
		_, err = client.BulkWrite(sessCtx, operations, opts)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	_, txErr := session.WithTransaction(ctx, transactionWork)
	return txErr
}

func (v2dbclient *V2Database) GetFinalityProviderStats(
	ctx context.Context,
) ([]*v2dbmodel.V2FinalityProviderStatsDocument, error) {
	client := v2dbclient.Client.Database(v2dbclient.DbName).Collection(dbmodel.V2FinalityProviderStatsCollection)
	cursor, err := client.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*v2dbmodel.V2FinalityProviderStatsDocument
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
