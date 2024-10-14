package db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient(ctx context.Context, cfg *config.DbConfig) (*mongo.Client, error) {
	credential := options.Credential{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	clientOps := options.Client().ApplyURI(cfg.Address).SetAuth(credential)
	return mongo.Connect(ctx, clientOps)
}

type Database struct {
	DbName string
	Client *mongo.Client
	Cfg    *config.DbConfig
}

func (db *Database) Ping(ctx context.Context) error {
	err := db.Client.Ping(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
