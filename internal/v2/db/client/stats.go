package v2dbclient

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v2dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v2/db/model"
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
	stakingTxHashHex = strings.ToLower(stakingTxHashHex)

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
	ctx context.Context, stakingTxHashHex string, amount uint64,
) error {
	stakingTxHashHex = strings.ToLower(stakingTxHashHex)

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
	return txErr
}

// SubtractOverallStats decrements the overall stats for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v2dbclient *V2Database) SubtractOverallStats(
	ctx context.Context, stakingTxHashHex string, amount uint64,
) error {
	stakingTxHashHex = strings.ToLower(stakingTxHashHex)

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

// HandleActiveStakerStats handles the active event for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v2dbclient *V2Database) HandleActiveStakerStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64,
) error {
	stakingTxHashHex = strings.ToLower(stakingTxHashHex)
	stakerPkHex = strings.ToLower(stakerPkHex)

	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":         int64(amount),
			"active_delegations": 1,
		},
	}
	return v2dbclient.updateStakerStats(ctx, types.Active.ToString(), stakingTxHashHex, stakerPkHex, upsertUpdate)
}

// HandleUnbondingStakerStats handles the unbonding event for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v2dbclient *V2Database) HandleUnbondingStakerStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64, stateHistory []string,
) error {
	stakingTxHashHex = strings.ToLower(stakingTxHashHex)
	stakerPkHex = strings.ToLower(stakerPkHex)

	// Check if we should process this state change
	for _, state := range stateHistory {
		stateLower := strings.ToLower(state)
		if stateLower == types.Withdrawn.ToString() || stateLower == types.Withdrawable.ToString() {
			// This may happen when Active -> Withdrawn -> Slashed or Active -> Withdrawable -> Slashed
			// Stats already handled by ProcessWithdrawnDelegationStats or ProcessWithdrawableDelegationStats
			return nil
		}
	}

	// It is certain the active event is emitted by the indexer
	// so we need to decrement the active stats
	upsertUpdate := bson.M{
		"$inc": bson.M{
			"active_tvl":            -int64(amount),
			"active_delegations":    -1,
			"unbonding_tvl":         int64(amount),
			"unbonding_delegations": 1,
		},
	}
	return v2dbclient.updateStakerStats(ctx, types.Unbonding.ToString(), stakingTxHashHex, stakerPkHex, upsertUpdate)
}

// HandleWithdrawableStakerStats handles the withdrawable event for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v2dbclient *V2Database) HandleWithdrawableStakerStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64, stateHistory []string,
) error {
	if len(stateHistory) < 1 {
		return fmt.Errorf("state history should have at least 1 state")
	}

	stakingTxHashHex = strings.ToLower(stakingTxHashHex)
	stakerPkHex = strings.ToLower(stakerPkHex)

	statsUpdates := bson.M{
		"withdrawable_tvl":         int64(amount),
		"withdrawable_delegations": 1,
	}

	var hasUnbondingState bool
	for _, state := range stateHistory {
		if strings.ToLower(state) == types.Unbonding.ToString() || strings.ToLower(state) == types.Slashed.ToString() {
			// Both slashed and unbonding events are pushed into the unbonding queue since they affect
			// the same stats. We use the same stats lock key to prevent double counting when both events
			// occur.
			// TODO: Consider using a separate queue for slashed events to avoid confusion, in that case
			// if we have separate lock key for slashed we need to ensure we don't double count.
			hasUnbondingState = true
			break
		}
	}

	if hasUnbondingState {
		statsUpdates["unbonding_tvl"] = -int64(amount)
		statsUpdates["unbonding_delegations"] = -1
	} else {
		statsUpdates["active_tvl"] = -int64(amount)
		statsUpdates["active_delegations"] = -1
	}

	// Apply the stats updates atomically
	upsertUpdate := bson.M{
		"$inc": statsUpdates,
	}

	return v2dbclient.updateStakerStats(ctx, types.Withdrawable.ToString(), stakingTxHashHex, stakerPkHex, upsertUpdate)
}

// HandleWithdrawnStakerStats handles the withdrawn event for the given staking tx hash
// This method is idempotent, only the first call will be processed. Otherwise it will return a notFoundError for duplicates
func (v2dbclient *V2Database) HandleWithdrawnStakerStats(
	ctx context.Context, stakingTxHashHex, stakerPkHex string, amount uint64, stateHistory []string,
) error {
	if len(stateHistory) < 1 {
		return fmt.Errorf("state history should have at least 1 state")
	}

	stakingTxHashHex = strings.ToLower(stakingTxHashHex)
	stakerPkHex = strings.ToLower(stakerPkHex)

	// Initialize empty stats updates map
	statsUpdates := bson.M{}

	var (
		hasWithdrawableState bool
		hasUnbondingState    bool
	)

	for _, state := range stateHistory {
		switch strings.ToLower(state) {
		case types.Withdrawable.ToString():
			hasWithdrawableState = true
		case types.Unbonding.ToString(), types.Slashed.ToString():
			// Both slashed and unbonding events are pushed into the unbonding queue since they affect
			// the same stats. We use the same stats lock key to prevent double counting when both events
			// occur.
			// TODO: Consider using a separate queue for slashed events to avoid confusion, in that case
			// if we have separate lock key for slashed we need to ensure we don't double count.
			hasUnbondingState = true
		}
	}

	// Handle stats updates based on state transition history:
	//
	// 1. If both withdrawable and unbonding occurred:
	//    - Decrement withdrawable stats since it was the final state before withdrawn
	// 2. If neither withdrawable nor unbonding occurred:
	//    - Decrement active stats since delegation was withdrawn directly from active state
	// 3. If only unbonding occurred (no withdrawable):
	//    - Decrement unbonding stats since delegation was withdrawn from unbonding state
	// 4. If only withdrawable occurred (no unbonding):
	//    - Decrement both withdrawable and active stats since delegation transitioned
	//      directly from active to withdrawn
	switch {
	case hasWithdrawableState && hasUnbondingState:
		statsUpdates["withdrawable_tvl"] = -int64(amount)
		statsUpdates["withdrawable_delegations"] = -1
	case !hasWithdrawableState && !hasUnbondingState:
		statsUpdates["active_tvl"] = -int64(amount)
		statsUpdates["active_delegations"] = -1
	case !hasWithdrawableState && hasUnbondingState:
		statsUpdates["unbonding_tvl"] = -int64(amount)
		statsUpdates["unbonding_delegations"] = -1
	case hasWithdrawableState && !hasUnbondingState:
		statsUpdates["withdrawable_tvl"] = -int64(amount)
		statsUpdates["withdrawable_delegations"] = -1
		statsUpdates["active_tvl"] = -int64(amount)
		statsUpdates["active_delegations"] = -1
	}

	// Apply the stats updates atomically
	upsertUpdate := bson.M{
		"$inc": statsUpdates,
	}

	return v2dbclient.updateStakerStats(ctx, types.Withdrawn.ToString(), stakingTxHashHex, stakerPkHex, upsertUpdate)
}

func (v2dbclient *V2Database) updateStakerStats(ctx context.Context, state, stakingTxHashHex, stakerPkHex string, upsertUpdate primitive.M) error {
	stakingTxHashHex = strings.ToLower(stakingTxHashHex)
	stakerPkHex = strings.ToLower(stakerPkHex)

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
	stakerPkHex = strings.ToLower(stakerPkHex)

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
	stakingTxHashHex = strings.ToLower(stakingTxHashHex)

	// Create bulk write operations for each FP
	var operations []mongo.WriteModel
	for _, fpPkHex := range fpPkHexes {
		fpPkHex = strings.ToLower(fpPkHex)
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
	stakingTxHashHex = strings.ToLower(stakingTxHashHex)

	// Create bulk write operations for each FP
	var operations []mongo.WriteModel
	for _, fpPkHex := range fpPkHexes {
		fpPkHex = strings.ToLower(fpPkHex)
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
	stakingTxHashHex = strings.ToLower(stakingTxHashHex)

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
