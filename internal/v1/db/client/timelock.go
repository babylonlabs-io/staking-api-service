package v1dbclient

import (
	"context"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	v1dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
)

func (v1dbclient *V1Database) SaveTimeLockExpireCheck(
	ctx context.Context, stakingTxHashHex string,
	expireHeight uint64, txType string,
) error {
	client := v1dbclient.Client.Database(v1dbclient.DbName).Collection(dbmodel.V1TimeLockCollection)
	document := v1dbmodel.NewTimeLockDocument(stakingTxHashHex, expireHeight, txType)
	_, err := client.InsertOne(ctx, document)
	if err != nil {
		return err
	}
	return nil
}

func (v1dbclient *V1Database) TransitionToUnbondedState(
	ctx context.Context, stakingTxHashHex string, eligiblePreviousState []types.DelegationState,
) error {
	return v1dbclient.transitionState(ctx, stakingTxHashHex, types.Unbonded.ToString(), eligiblePreviousState, nil)
}
