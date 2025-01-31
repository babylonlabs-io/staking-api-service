package dbclient

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	DbName string
	Client *mongo.Client
	Cfg    *config.DbConfig
}

func NewMongoClient(ctx context.Context, cfg *config.DbConfig) (*mongo.Client, error) {
	credential := options.Credential{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	clientOps := options.Client().ApplyURI(cfg.Address).SetAuth(credential)
	return mongo.Connect(ctx, clientOps)
}

func (db *Database) Ping(ctx context.Context) error {
	err := db.Client.Ping(ctx, nil)
	if err != nil {
		metrics.RecordDbError("ping")
		return err
	}
	return nil
}

func New(ctx context.Context, client *mongo.Client, cfg *config.DbConfig) (*Database, error) {
	return &Database{
		DbName: cfg.DbName,
		Client: client,
		Cfg:    cfg,
	}, nil
}
