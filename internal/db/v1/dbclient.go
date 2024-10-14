package v1db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db"
	"go.mongodb.org/mongo-driver/mongo"
)

type V1Database struct {
	*db.Database
}

func New(ctx context.Context, client *mongo.Client, cfg *config.DbConfig) (*V1Database, error) {
	return &V1Database{
		Database: &db.Database{
			DbName: cfg.DbName,
			Client: client,
			Cfg:    cfg,
		},
	}, nil
}
