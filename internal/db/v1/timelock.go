package v1db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	v1model "github.com/babylonlabs-io/staking-api-service/internal/db/model/v1"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
)

func (v1db *V1Database) SaveTimeLockExpireCheck(
	ctx context.Context, stakingTxHashHex string,
	expireHeight uint64, txType string,
) error {
	client := v1db.Client.Database(v1db.DbName).Collection(model.TimeLockCollection)
	document := v1model.NewTimeLockDocument(stakingTxHashHex, expireHeight, txType)
	_, err := client.InsertOne(ctx, document)
	if err != nil {
		return err
	}
	return nil
}

func (v1db *V1Database) TransitionToUnbondedState(
	ctx context.Context, stakingTxHashHex string, eligiblePreviousState []types.DelegationState,
) error {
	return v1db.transitionState(ctx, stakingTxHashHex, types.Unbonded.ToString(), eligiblePreviousState, nil)
}
