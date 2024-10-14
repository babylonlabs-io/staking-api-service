package v2db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db"
	"go.mongodb.org/mongo-driver/mongo"
)

type V2Database struct {
	*db.Database
}

func New(ctx context.Context, client *mongo.Client, cfg *config.DbConfig) (*V2Database, error) {
	return &V2Database{
		Database: &db.Database{
			DbName: cfg.DbName,
			Client: client,
			Cfg:    cfg,
		},
	}, nil
}
