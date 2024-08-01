package db

import (
	"context"
	"errors"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (db *Database) UpsertLatestBtcInfo(
	ctx context.Context, height uint64, confirmedTvl, unconfirmedTvl uint64,
) error {
	client := db.Client.Database(db.DbName).Collection(model.BtcInfoCollection)
	// Start a session
	session, sessionErr := db.Client.StartSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Check for existing document
		var existingInfo model.BtcInfo
		findErr := client.FindOne(sessCtx, bson.M{"_id": model.LatestBtcInfoId}).Decode(&existingInfo)
		if findErr != nil && findErr != mongo.ErrNoDocuments {
			return nil, findErr
		}

		btcInfo := &model.BtcInfo{
			ID:             model.LatestBtcInfoId,
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
				sessCtx, bson.M{"_id": model.LatestBtcInfoId},
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

func (db *Database) GetLatestBtcInfo(ctx context.Context) (*model.BtcInfo, error) {
	client := db.Client.Database(db.DbName).Collection(model.BtcInfoCollection)
	var btcInfo model.BtcInfo
	err := client.FindOne(ctx, bson.M{"_id": model.LatestBtcInfoId}).Decode(&btcInfo)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &NotFoundError{
				Key:     model.LatestBtcInfoId,
				Message: "Latest Btc info not found",
			}
		}
		return nil, err
	}

	return &btcInfo, nil
}
