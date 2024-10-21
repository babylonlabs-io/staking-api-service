package v1dbclient

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	"go.mongodb.org/mongo-driver/mongo"
)

type V1Database struct {
	*dbclient.Database
}

func New(ctx context.Context, client *mongo.Client, cfg *config.DbConfig) (*V1Database, error) {
	return &V1Database{
		Database: &dbclient.Database{
			DbName: cfg.DbName,
			Client: client,
			Cfg:    cfg,
		},
	}, nil
}
