package dbclient

import (
	"context"
	model "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (db *Database) GetLatestBtcPrice(ctx context.Context) (*model.BtcPrice, error) {
	client := db.Client.Database(db.DbName).Collection(model.BtcPriceCollection)
	var btcPrice model.BtcPrice
	err := client.FindOne(ctx, bson.M{"_id": model.BtcPriceDocID}).Decode(&btcPrice)
	if err != nil {
		return nil, err
	}
	return &btcPrice, nil
}
func (db *Database) SetBtcPrice(ctx context.Context, price float64) error {
	client := db.Client.Database(db.DbName).Collection(model.BtcPriceCollection)
	btcPrice := model.BtcPrice{
		ID:        model.BtcPriceDocID, // Fixed ID for single document
		Price:     price,
		CreatedAt: time.Now(), // For TTL index
	}
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": model.BtcPriceDocID}
	update := bson.M{"$set": btcPrice}
	_, err := client.UpdateOne(ctx, filter, update, opts)
	return err
}
