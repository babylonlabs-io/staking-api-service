package v2dbclient

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	"go.mongodb.org/mongo-driver/mongo"
)

type V2Database struct {
	*dbclient.Database
}

func New(ctx context.Context, client *mongo.Client, cfg *config.DbConfig) (*V2Database, error) {
	return &V2Database{
		Database: &dbclient.Database{
			DbName: cfg.DbName,
			Client: client,
			Cfg:    cfg,
		},
	}, nil
}
