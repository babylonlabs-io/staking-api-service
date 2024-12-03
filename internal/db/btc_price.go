package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
)

func (db *Database) GetLatestBtcPrice(ctx context.Context) (*model.BtcPrice, error) {
	client := db.Client.Database(db.DbName).Collection(model.BtcPriceCollection)

	var btcPrice model.BtcPrice
	err := client.FindOne(ctx, bson.M{}).Decode(&btcPrice)
	if err != nil {
		return nil, err
	}

	return &btcPrice, nil
}

func (db *Database) SetBtcPrice(ctx context.Context, price float64) error {
	client := db.Client.Database(db.DbName).Collection(model.BtcPriceCollection)

	// Always store as a single document
	_, err := client.DeleteMany(ctx, bson.M{})
	if err != nil {
		return err
	}

	btcPrice := model.BtcPrice{
		Price:     price,
		CreatedAt: time.Now(),
	}

	_, err = client.InsertOne(ctx, btcPrice)
	return err
}
