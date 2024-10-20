package v1dbclient

import (
	"context"
	"errors"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/utils"
	v1dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (v1dbclient *V1Database) SaveUnbondingTx(
	ctx context.Context, stakingTxHashHex, txHashHex, txHex, signatureHex string,
) error {
	delegationClient := v1dbclient.Client.Database(v1dbclient.DbName).Collection(dbmodel.V1DelegationCollection)
	unbondingClient := v1dbclient.Client.Database(v1dbclient.DbName).Collection(dbmodel.V1UnbondingCollection)

	// Start a session
	session, err := v1dbclient.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Define the work to be done in the transaction
	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Find the existing delegation document first, it will be used later in the transaction
		delegationFilter := bson.M{
			"_id":   stakingTxHashHex,
			"state": types.Active,
		}
		var delegationDocument v1dbmodel.DelegationDocument
		err = delegationClient.FindOne(sessCtx, delegationFilter).Decode(&delegationDocument)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, &db.NotFoundError{
					Key:     stakingTxHashHex,
					Message: "no active delegation found for unbonding request",
				}
			}
			return nil, err
		}
		// Update the state to UnbondingRequested
		delegationUpdate := bson.M{"$set": bson.M{"state": types.UnbondingRequested}}
		result, err := delegationClient.UpdateOne(sessCtx, delegationFilter, delegationUpdate)
		if err != nil {
			return nil, err
		}

		if result.MatchedCount == 0 {
			return nil, &db.NotFoundError{
				Key:     stakingTxHashHex,
				Message: "delegation not found or not eligible for unbonding",
			}
		}

		// Insert the unbonding transaction document
		unbondingDocument := v1dbmodel.UnbondingDocument{
			StakerPkHex:        delegationDocument.StakerPkHex,
			FinalityPkHex:      delegationDocument.FinalityProviderPkHex,
			UnbondingTxSigHex:  signatureHex,
			State:              v1dbmodel.UnbondingInitialState,
			UnbondingTxHashHex: txHashHex,
			UnbondingTxHex:     txHex,
			StakingTxHex:       delegationDocument.StakingTx.TxHex,
			StakingOutputIndex: delegationDocument.StakingTx.OutputIndex,
			StakingTimelock:    delegationDocument.StakingTx.TimeLock,
			StakingTxHashHex:   stakingTxHashHex,
			StakingAmount:      delegationDocument.StakingValue,
		}
		_, err = unbondingClient.InsertOne(sessCtx, unbondingDocument)
		if err != nil {
			var writeErr mongo.WriteException
			if errors.As(err, &writeErr) {
				for _, e := range writeErr.WriteErrors {
					if mongo.IsDuplicateKeyError(e) {
						return nil, &db.DuplicateKeyError{
							Key:     txHashHex,
							Message: "unbonding transaction already exists",
						}
					}
				}
			}
			return nil, err
		}

		return nil, nil
	}

	// Execute the transaction
	_, err = session.WithTransaction(ctx, transactionWork)
	if err != nil {
		return err
	}

	return nil
}

// Change the state to `unbonding` and save the unbondingTx data
// Return not found error if the stakingTxHashHex is not found or the existing state is not eligible for unbonding
func (v1dbclient *V1Database) TransitionToUnbondingState(
	ctx context.Context, txHashHex string, startHeight, timelock, outputIndex uint64, txHex string, startTimestamp int64,
) error {
	unbondingTxMap := make(map[string]interface{})
	unbondingTxMap["unbonding_tx"] = v1dbmodel.TimelockTransaction{
		TxHex:          txHex,
		OutputIndex:    outputIndex,
		StartTimestamp: startTimestamp,
		StartHeight:    startHeight,
		TimeLock:       timelock,
	}

	err := v1dbclient.transitionState(
		ctx, txHashHex, types.Unbonding.ToString(),
		utils.QualifiedStatesToUnbonding(), unbondingTxMap,
	)
	if err != nil {
		return err
	}
	return nil
}
