package v1dbclient

import (
	"context"
	"errors"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	v1dbmodel "github.com/babylonlabs-io/staking-api-service/internal/v1/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (v1dbclient *V1Database) UpsertLatestBtcInfo(
	ctx context.Context, height uint64, confirmedTvl, unconfirmedTvl uint64,
) error {
	client := v1dbclient.Client.Database(v1dbclient.DbName).Collection(dbmodel.V1BtcInfoCollection)
	// Start a session
	session, sessionErr := v1dbclient.Client.StartSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Check for existing document
		var existingInfo v1dbmodel.BtcInfo
		findErr := client.FindOne(sessCtx, bson.M{"_id": v1dbmodel.LatestBtcInfoId}).Decode(&existingInfo)
		if findErr != nil && findErr != mongo.ErrNoDocuments {
			return nil, findErr
		}

		btcInfo := &v1dbmodel.BtcInfo{
			ID:             v1dbmodel.LatestBtcInfoId,
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
				sessCtx, bson.M{"_id": v1dbmodel.LatestBtcInfoId},
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

func (v1dbclient *V1Database) GetLatestBtcInfo(ctx context.Context) (*v1dbmodel.BtcInfo, error) {
	client := v1dbclient.Client.Database(v1dbclient.DbName).Collection(dbmodel.V1BtcInfoCollection)
	var btcInfo v1dbmodel.BtcInfo
	err := client.FindOne(ctx, bson.M{"_id": v1dbmodel.LatestBtcInfoId}).Decode(&btcInfo)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &db.NotFoundError{
				Key:     v1dbmodel.LatestBtcInfoId,
				Message: "Latest Btc info not found",
			}
		}
		return nil, err
	}

	return &btcInfo, nil
}
