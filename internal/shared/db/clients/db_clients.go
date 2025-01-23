package dbclients

import (
	"context"

	indexerdbclient "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/client"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	v1dbclient "github.com/babylonlabs-io/staking-api-service/internal/v1/db/client"
	v2dbclient "github.com/babylonlabs-io/staking-api-service/internal/v2/db/client"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type DbClients struct {
	StakingMongoClient *mongo.Client
	IndexerMongoClient *mongo.Client
	SharedDBClient     dbclient.DBClient
	V1DBClient         v1dbclient.V1DBClient
	V2DBClient         v2dbclient.V2DBClient
	IndexerDBClient    indexerdbclient.IndexerDBClient
}

func New(ctx context.Context, cfg *config.Config) (*DbClients, error) {
	stakingMongoClient, err := dbclient.NewMongoClient(ctx, cfg.StakingDb)
	if err != nil {
		return nil, err
	}

	dbClient, err := dbclient.New(ctx, stakingMongoClient, cfg.StakingDb)
	if err != nil {
		return nil, err
	}

	v1dbClient, err := v1dbclient.New(ctx, stakingMongoClient, cfg.StakingDb)
	if err != nil {
		metrics.RecordServiceCrash("db.New")
		log.Ctx(ctx).Fatal().Err(err).Msg("error while creating v1 db client")
		return nil, err
	}
	v2dbClient, err := v2dbclient.New(ctx, stakingMongoClient, cfg.StakingDb)
	if err != nil {
		metrics.RecordServiceCrash("db.New")
		log.Ctx(ctx).Fatal().Err(err).Msg("error while creating v2 db client")
		return nil, err
	}

	indexerMongoClient, err := dbclient.NewMongoClient(ctx, cfg.IndexerDb)
	if err != nil {
		return nil, err
	}

	indexerDbClient, err := indexerdbclient.New(ctx, indexerMongoClient, cfg.IndexerDb)
	if err != nil {
		metrics.RecordServiceCrash("db.New")
		log.Ctx(ctx).Fatal().Err(err).Msg("error while creating indexer db client")
		return nil, err
	}

	dbClients := DbClients{
		StakingMongoClient: stakingMongoClient,
		IndexerMongoClient: indexerMongoClient,
		SharedDBClient:     dbClient,
		V1DBClient:         v1dbClient,
		V2DBClient:         v2dbClient,
		IndexerDBClient:    indexerDbClient,
	}

	return &dbClients, nil
}
