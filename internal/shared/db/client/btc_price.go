package dbclient

import (
	"context"
	model "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

func (db *Database) GetLatestPrice(ctx context.Context, symbol string) (float64, error) {
	symbol = strings.ToLower(symbol)

	client := db.Client.Database(db.DbName).Collection(model.PriceCollection)
	var doc model.CoinPrice
	err := client.FindOne(ctx, bson.M{"_id": symbol}).Decode(&doc)
	if err != nil {
		return 0, err
	}
	return doc.Price, nil
}

func (db *Database) SetLatestPrice(ctx context.Context, symbol string, price float64) error {
	symbol = strings.ToLower(symbol)

	doc := model.CoinPrice{
		ID:        symbol,
		Price:     price,
		CreatedAt: time.Now(), // For TTL index
	}
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": symbol}
	update := bson.M{"$set": doc}

	client := db.Client.Database(db.DbName).Collection(model.PriceCollection)
	_, err := client.UpdateOne(ctx, filter, update, opts)
	return err
}
