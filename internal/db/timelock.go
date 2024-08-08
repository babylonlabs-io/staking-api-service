package db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

func (db *Database) SaveTimeLockExpireCheck(
	ctx context.Context, stakingTxHashHex string,
	expireHeight uint64, txType string,
) error {
	client := db.Client.Database(db.DbName).Collection(model.TimeLockCollection)
	document := model.NewTimeLockDocument(stakingTxHashHex, expireHeight, txType)
	_, err := client.InsertOne(ctx, document)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) TransitionToUnbondedState(
	ctx context.Context, stakingTxHashHex string, eligiblePreviousState []types.DelegationState,
) error {
	return db.transitionState(ctx, stakingTxHashHex, types.Unbonded.ToString(), eligiblePreviousState, nil)
}
