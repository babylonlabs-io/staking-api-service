package db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (db *Database) SaveUnprocessableMessage(ctx context.Context, messageBody, receipt string) error {
	unprocessableMsgClient := db.Client.Database(db.DbName).Collection(model.V1UnprocessableMsgCollection)

	_, err := unprocessableMsgClient.InsertOne(ctx, model.NewUnprocessableMessageDocument(messageBody, receipt))
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) FindUnprocessableMessages(ctx context.Context) ([]model.UnprocessableMessageDocument, error) {
	client := db.Client.Database(db.DbName).Collection(model.V1UnprocessableMsgCollection)
	filter := bson.M{}
	options := options.FindOptions{}

	cursor, err := client.Find(ctx, filter, &options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var unprocessableMessages []model.UnprocessableMessageDocument
	if err = cursor.All(ctx, &unprocessableMessages); err != nil {
		return nil, err
	}

	return unprocessableMessages, nil
}

func (db *Database) DeleteUnprocessableMessage(ctx context.Context, Receipt interface{}) error {
	unprocessableMsgClient := db.Client.Database(db.DbName).Collection(model.V1UnprocessableMsgCollection)
	filter := bson.M{"receipt": Receipt}
	_, err := unprocessableMsgClient.DeleteOne(ctx, filter)
	return err
}
