package v2dbclient

import (
	"context"
	"errors"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/db"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (v2dbclient *V2Database) MarkV1DelegationAsTransitioned(ctx context.Context, stakingTxHashHex string) error {
	session, err := v2dbclient.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	transactionWork := func(sessCtx mongo.SessionContext) (interface{}, error) {
		client := v2dbclient.Client.Database(v2dbclient.DbName).Collection(dbmodel.V1DelegationCollection)
		filter := bson.M{"_id": stakingTxHashHex}
		
		var delegation interface{}
		err := client.FindOne(sessCtx, filter).Decode(&delegation)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, &db.NotFoundError{
					Key:     stakingTxHashHex,
					Message: "Delegation not found",
				}
			}
			return nil, err
		}

		update := bson.M{"$set": bson.M{"transitioned": true}}
		result, err := client.UpdateOne(sessCtx, filter, update)
		if err != nil {
			return nil, err
		}

		if result.MatchedCount == 0 {
			return nil, &db.NotFoundError{
				Key:     stakingTxHashHex,
				Message: "Delegation not found",
			}
		}

		return nil, nil
	}

	_, err = session.WithTransaction(ctx, transactionWork)
	return err
}

