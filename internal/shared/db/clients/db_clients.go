package dbclients

import (
	"context"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	v1dbclient "github.com/babylonlabs-io/staking-api-service/internal/v1/db/client"
	v2dbclient "github.com/babylonlabs-io/staking-api-service/internal/v2/db/client"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type DbClients struct {
	MongoClient    *mongo.Client
	SharedDBClient dbclient.DBClient
	V1DBClient     v1dbclient.V1DBClient
	V2DBClient     v2dbclient.V2DBClient
}

func New(ctx context.Context, cfg *config.Config) (*DbClients, error) {
	mongoClient, err := dbclient.NewMongoClient(ctx, cfg.Db)
	if err != nil {
		return nil, err
	}

	dbClient, err := dbclient.New(ctx, mongoClient, cfg.Db)
	if err != nil {
		return nil, err
	}

	v1dbClient, err := v1dbclient.New(ctx, mongoClient, cfg.Db)
	if err != nil {
		log.Ctx(ctx).Fatal().Err(err).Msg("error while creating v1 db client")
		return nil, err
	}
	v2dbClient, err := v2dbclient.New(ctx, mongoClient, cfg.Db)
	if err != nil {
		log.Ctx(ctx).Fatal().Err(err).Msg("error while creating v2 db client")
		return nil, err
	}

	dbClients := DbClients{
		MongoClient:    mongoClient,
		SharedDBClient: dbClient,
		V1DBClient:     v1dbClient,
		V2DBClient:     v2dbClient,
	}

	return &dbClients, nil
}
