package v1db

import (
	"context"
	"errors"

	"github.com/babylonlabs-io/staking-api-service/internal/db"
	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	v1model "github.com/babylonlabs-io/staking-api-service/internal/db/model/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (v1db *V1Database) UpsertLatestBtcInfo(
	ctx context.Context, height uint64, confirmedTvl, unconfirmedTvl uint64,
) error {
	client := v1db.Client.Database(v1db.DbName).Collection(model.V1BtcInfoCollection)
	// Start a session
	session, sessionErr := v1db.Client.StartSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Check for existing document
		var existingInfo v1model.BtcInfo
		findErr := client.FindOne(sessCtx, bson.M{"_id": v1model.LatestBtcInfoId}).Decode(&existingInfo)
		if findErr != nil && findErr != mongo.ErrNoDocuments {
			return nil, findErr
		}

		btcInfo := &v1model.BtcInfo{
			ID:             v1model.LatestBtcInfoId,
			BtcHeight:      height,
			ConfirmedTvl:   confirmedTvl,
			UnconfirmedTvl: unconfirmedTvl,
		}
		if findErr == mongo.ErrNoDocuments {
			// If no document exists, insert a new one
			_, insertErr := client.InsertOne(sessCtx, btcInfo)
			if insertErr != nil {
				return nil, insertErr
			}
			return nil, nil
		}

		// If document exists and the incoming height is greater, update the document
		if existingInfo.BtcHeight < height {
			_, updateErr := client.UpdateOne(
				sessCtx, bson.M{"_id": v1model.LatestBtcInfoId},
				bson.M{"$set": btcInfo},
			)
			if updateErr != nil {
				return nil, updateErr
			}
		}
		return nil, nil
	}

	// Execute the transaction
	_, txErr := session.WithTransaction(ctx, transactionWork)
	return txErr
}

func (v1db *V1Database) GetLatestBtcInfo(ctx context.Context) (*v1model.BtcInfo, error) {
	client := v1db.Client.Database(v1db.DbName).Collection(model.V1BtcInfoCollection)
	var btcInfo v1model.BtcInfo
	err := client.FindOne(ctx, bson.M{"_id": v1model.LatestBtcInfoId}).Decode(&btcInfo)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &db.NotFoundError{
				Key:     v1model.LatestBtcInfoId,
				Message: "Latest Btc info not found",
			}
		}
		return nil, err
	}

	return &btcInfo, nil
}
