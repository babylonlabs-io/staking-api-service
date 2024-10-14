package v2db

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db"
)

type V2Database struct {
	*db.Database
}

func New(ctx context.Context, cfg *config.DbConfig) (*V2Database, error) {
	client, err := db.NewMongoClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &V2Database{
		Database: &db.Database{
			DbName: cfg.DbName,
			Client: client,
			Cfg:    cfg,
		},
	}, nil
}
