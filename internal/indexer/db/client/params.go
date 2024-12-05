package indexerdbclient

import (
	"context"
	"fmt"

	indexerdbmodel "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	indexertypes "github.com/babylonlabs-io/staking-api-service/internal/indexer/types"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *IndexerDatabase) GetBbnStakingParams(ctx context.Context) ([]*indexertypes.BbnStakingParams, error) {
	cursor, err := db.Client.Database(db.DbName).Collection(indexerdbmodel.GlobalParamsCollection).Find(ctx, bson.M{
		"type": indexertypes.STAKING_PARAMS_TYPE,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query global params: %w", err)
	}
	defer cursor.Close(ctx)

	var params []*indexertypes.BbnStakingParams

	for cursor.Next(ctx) {
		var model indexerdbmodel.IndexerGlobalParamsDocument
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		var stakingParams indexerdbmodel.IndexerBbnStakingParamsDocument
		bsonBytes, err := bson.Marshal(model.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		if err := bson.Unmarshal(bsonBytes, &stakingParams); err != nil {
			return nil, fmt.Errorf("failed to unmarshal into staking params: %w", err)
		}

		bbnParams := &indexertypes.BbnStakingParams{
			Version:                      model.Version,
			CovenantPks:                  stakingParams.CovenantPks,
			CovenantQuorum:               stakingParams.CovenantQuorum,
			MinStakingValueSat:           stakingParams.MinStakingValueSat,
			MaxStakingValueSat:           stakingParams.MaxStakingValueSat,
			MinStakingTimeBlocks:         stakingParams.MinStakingTimeBlocks,
			MaxStakingTimeBlocks:         stakingParams.MaxStakingTimeBlocks,
			SlashingPkScript:             stakingParams.SlashingPkScript,
			MinSlashingTxFeeSat:          stakingParams.MinSlashingTxFeeSat,
			SlashingRate:                 stakingParams.SlashingRate,
			UnbondingTimeBlocks:          stakingParams.UnbondingTimeBlocks,
			UnbondingFeeSat:              stakingParams.UnbondingFeeSat,
			MinCommissionRate:            stakingParams.MinCommissionRate,
			MaxActiveFinalityProviders:   stakingParams.MaxActiveFinalityProviders,
			DelegationCreationBaseGasFee: stakingParams.DelegationCreationBaseGasFee,
			AllowListExpirationHeight:    stakingParams.AllowListExpirationHeight,
			BtcActivationHeight:          stakingParams.BtcActivationHeight,
		}

		params = append(params, bbnParams)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor iteration error: %w", err)
	}

	return params, nil
}

func (db *IndexerDatabase) GetBtcCheckpointParams(ctx context.Context) ([]*indexertypes.BtcCheckpointParams, error) {
	cursor, err := db.Client.Database(db.DbName).Collection(indexerdbmodel.GlobalParamsCollection).Find(ctx, bson.M{
		"type": indexertypes.CHECKPOINT_PARAMS_TYPE,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query global params: %w", err)
	}
	defer cursor.Close(ctx)

	var params []*indexertypes.BtcCheckpointParams

	for cursor.Next(ctx) {
		var model indexerdbmodel.IndexerGlobalParamsDocument
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		var btcParamsDoc indexerdbmodel.IndexerBtcCheckpointParamsDocument
		bsonBytes, err := bson.Marshal(model.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		if err := bson.Unmarshal(bsonBytes, &btcParamsDoc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal into checkpoint params: %w", err)
		}

		btcParams := &indexertypes.BtcCheckpointParams{
			Version:              model.Version,
			BtcConfirmationDepth: btcParamsDoc.BtcConfirmationDepth,
		}

		params = append(params, btcParams)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor iteration error: %w", err)
	}

	return params, nil
}
