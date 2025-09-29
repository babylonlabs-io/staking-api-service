package dbclient

import (
	"context"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/babylonlabs-io/staking-api-service/pkg"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *Database) SaveUnprocessableMessage(ctx context.Context, messageBody, receipt string) error {
	unprocessableMsgClient := db.Client.Database(db.DbName).Collection(dbmodel.V1UnprocessableMsgCollection)

	_, err := unprocessableMsgClient.InsertOne(ctx, dbmodel.NewUnprocessableMessageDocument(messageBody, receipt))
	if err != nil {
		metrics.RecordDbError("save_unprocessable_message")
	}

	return err
}

func (db *Database) FindUnprocessableMessages(ctx context.Context) ([]dbmodel.UnprocessableMessageDocument, error) {
	client := db.Client.Database(db.DbName).Collection(dbmodel.V1UnprocessableMsgCollection)

	unprocessableMessages, err := pkg.FetchAll[dbmodel.UnprocessableMessageDocument](ctx, client, bson.M{})
	if err != nil {
		metrics.RecordDbError("find_unprocessable_messages")
		return nil, err
	}

	return unprocessableMessages, nil
}

func (db *Database) DeleteUnprocessableMessage(ctx context.Context, receipt interface{}) error {
	unprocessableMsgClient := db.Client.Database(db.DbName).Collection(dbmodel.V1UnprocessableMsgCollection)
	filter := bson.M{"receipt": receipt}
	_, err := unprocessableMsgClient.DeleteOne(ctx, filter)
	if err != nil {
		metrics.RecordDbError("delete_unprocessable_message")
	}
	return err
}
