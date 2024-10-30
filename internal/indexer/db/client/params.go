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

		var stakingParams indexertypes.BbnStakingParams
		bsonBytes, err := bson.Marshal(model.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		if err := bson.Unmarshal(bsonBytes, &stakingParams); err != nil {
			return nil, fmt.Errorf("failed to unmarshal into staking params: %w", err)
		}

		stakingParams.Version = model.Version

		params = append(params, &stakingParams)
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

		var btcParams indexertypes.BtcCheckpointParams

		bsonBytes, err := bson.Marshal(model.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		if err := bson.Unmarshal(bsonBytes, &btcParams); err != nil {
			return nil, fmt.Errorf("failed to unmarshal into staking params: %w", err)
		}

		btcParams.Version = model.Version

		params = append(params, &btcParams)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor iteration error: %w", err)
	}

	return params, nil
}
