//go:build integration

// todo comment
package indexerdbclient_test

import (
	"context"
	"fmt"
	"time"

	model "github.com/babylonlabs-io/staking-api-service/internal/indexer/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type index struct {
	Indexes map[string]int
	Unique  bool
}

var collections = map[string][]index{
	model.FinalityProviderDetailsCollection: {{Indexes: map[string]int{}}},
	model.BTCDelegationDetailsCollection: {
		{
			Indexes: map[string]int{
				"staker_btc_pk_hex":                       1,
				"btc_delegation_created_bbn_block.height": -1,
				"_id": 1,
			},
			Unique: false,
		},
	},
	model.TimeLockCollection: {
		{Indexes: map[string]int{"expire_height": 1}, Unique: false},
	},
	model.GlobalParamsCollection:        {{Indexes: map[string]int{}}},
	model.LastProcessedHeightCollection: {{Indexes: map[string]int{}}},
}

func Setup(ctx context.Context, cfg *config.DbConfig) error {
	credential := options.Credential{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	clientOps := options.Client().ApplyURI(cfg.Address).SetAuth(credential)
	client, err := mongo.Connect(ctx, clientOps)
	if err != nil {
		return err
	}

	// Create a context with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd
	defer cancel()

	// Access a database and create collections.
	database := client.Database(cfg.DbName)

	// Create collections.
	for collection := range collections {
		err := createCollection(ctx, database, collection)
		if err != nil {
			return fmt.Errorf("failed to create %q collection: %w", collection, err)
		}
	}

	for name, idxs := range collections {
		for _, idx := range idxs {
			err = createIndex(ctx, database, name, idx)
			if err != nil {
				return fmt.Errorf("failed to create index %q: %w", name, err)
			}
		}
	}

	return nil
}

func createCollection(ctx context.Context, database *mongo.Database, collectionName string) error {
	return database.CreateCollection(ctx, collectionName)
}

func createIndex(ctx context.Context, database *mongo.Database, collectionName string, idx index) error {
	if len(idx.Indexes) == 0 {
		return nil
	}

	indexKeys := bson.D{}
	for k, v := range idx.Indexes {
		indexKeys = append(indexKeys, bson.E{Key: k, Value: v})
	}

	index := mongo.IndexModel{
		Keys:    indexKeys,
		Options: options.Index().SetUnique(idx.Unique),
	}

	_, err := database.Collection(collectionName).Indexes().CreateOne(ctx, index)
	return err
}
