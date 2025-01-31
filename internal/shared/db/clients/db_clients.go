package dbclients

import (
	"context"

	"fmt"
	indexerdbclient "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/client"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	v1dbclient "github.com/babylonlabs-io/staking-api-service/internal/v1/db/client"
	v2dbclient "github.com/babylonlabs-io/staking-api-service/internal/v2/db/client"
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
		return nil, fmt.Errorf("error while creating v1 db client: %w", err)
	}
	v2dbClient, err := v2dbclient.New(ctx, stakingMongoClient, cfg.StakingDb)
	if err != nil {
		return nil, fmt.Errorf("error while creating v2 db client: %w", err)
	}

	indexerMongoClient, err := dbclient.NewMongoClient(ctx, cfg.IndexerDb)
	if err != nil {
		return nil, fmt.Errorf("error while creating indexer mongo client: %w", err)
	}

	indexerDbClient, err := indexerdbclient.New(ctx, indexerMongoClient, cfg.IndexerDb)
	if err != nil {
		return nil, fmt.Errorf("error while creating indexer db client: %w", err)
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
