package dbclient

import (
	"context"

	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	filter := bson.M{}
	options := options.FindOptions{}

	cursor, err := client.Find(ctx, filter, &options)
	if err != nil {
		metrics.RecordDbError("find_unprocessable_messages")
		return nil, err
	}
	defer cursor.Close(ctx)

	var unprocessableMessages []dbmodel.UnprocessableMessageDocument
	if err = cursor.All(ctx, &unprocessableMessages); err != nil {
		metrics.RecordDbError("find_unprocessable_messages")
		return nil, err
	}

	return unprocessableMessages, nil
}

func (db *Database) DeleteUnprocessableMessage(ctx context.Context, Receipt interface{}) error {
	unprocessableMsgClient := db.Client.Database(db.DbName).Collection(dbmodel.V1UnprocessableMsgCollection)
	filter := bson.M{"receipt": Receipt}
	_, err := unprocessableMsgClient.DeleteOne(ctx, filter)
	if err != nil {
		metrics.RecordDbError("delete_unprocessable_message")
	}
	return err
}
