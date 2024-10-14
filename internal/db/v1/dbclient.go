package v1db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db"
)

type V1Database struct {
	*db.Database
}

func New(ctx context.Context, cfg *config.DbConfig) (*V1Database, error) {
	client, err := db.NewMongoClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &V1Database{
		Database: &db.Database{
			DbName: cfg.DbName,
			Client: client,
			Cfg:    cfg,
		},
	}, nil
}
