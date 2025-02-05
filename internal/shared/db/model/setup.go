package dbmodel

import (
	"context"
	"fmt"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/rs/zerolog/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

const (
	// Shared
	PkAddressMappingsCollection = "pk_address_mappings"
	// V1
	V1StatsLockCollection             = "stats_lock"
	V1OverallStatsCollection          = "overall_stats"
	V1FinalityProviderStatsCollection = "finality_providers_stats"
	V1StakerStatsCollection           = "staker_stats"
	V1DelegationCollection            = "delegations"
	V1TimeLockCollection              = "timelock_queue"
	V1UnbondingCollection             = "unbonding_queue"
	V1BtcInfoCollection               = "btc_info"
	V1UnprocessableMsgCollection      = "unprocessable_messages"
	PriceCollection                   = "prices"
	// V2
	V2StatsLockCollection             = "v2_stats_lock"
	V2OverallStatsCollection          = "v2_overall_stats"
	V2FinalityProviderStatsCollection = "v2_finality_providers_stats"
	V2StakerStatsCollection           = "v2_staker_stats"
)

type index struct {
	Indexes map[string]int
	Unique  bool
}

var collections = map[string][]index{
	// Shared
	PkAddressMappingsCollection: {
		{Indexes: map[string]int{"taproot": 1}, Unique: true},
		{Indexes: map[string]int{"native_segwit_odd": 1}, Unique: true},
		{Indexes: map[string]int{"native_segwit_even": 1}, Unique: true},
	},
	// V1
	V1StatsLockCollection:             {{Indexes: map[string]int{}}},
	V1OverallStatsCollection:          {{Indexes: map[string]int{}}},
	V1FinalityProviderStatsCollection: {{Indexes: map[string]int{"active_tvl": -1}, Unique: false}},
	V1StakerStatsCollection:           {{Indexes: map[string]int{"active_tvl": -1}, Unique: false}},
	V1DelegationCollection: {
		{Indexes: map[string]int{"staker_pk_hex": 1, "staking_tx.start_height": -1, "_id": 1}, Unique: false},
	},
	V1TimeLockCollection:         {{Indexes: map[string]int{"expire_height": 1}, Unique: false}},
	V1UnbondingCollection:        {{Indexes: map[string]int{"unbonding_tx_hash_hex": 1}, Unique: true}},
	V1UnprocessableMsgCollection: {{Indexes: map[string]int{}}},
	V1BtcInfoCollection:          {{Indexes: map[string]int{}}},
	// V2
	V2StatsLockCollection:             {{Indexes: map[string]int{}}},
	V2OverallStatsCollection:          {{Indexes: map[string]int{}}},
	V2FinalityProviderStatsCollection: {{Indexes: map[string]int{"active_tvl": -1}, Unique: false}},
	V2StakerStatsCollection:           {{Indexes: map[string]int{"active_tvl": -1}, Unique: false}},
}

func Setup(ctx context.Context, cfg *config.Config) error {
	credential := options.Credential{
		Username: cfg.StakingDb.Username,
		Password: cfg.StakingDb.Password,
	}
	clientOps := options.Client().ApplyURI(cfg.StakingDb.Address).SetAuth(credential)
	client, err := mongo.Connect(ctx, clientOps)
	if err != nil {
		return err
	}

	// Create a context with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Access a database and create collections.
	database := client.Database(cfg.StakingDb.DbName)

	// Create collections.
	for collection := range collections {
		createCollection(ctx, database, collection)
	}

	for name, idxs := range collections {
		for _, idx := range idxs {
			createIndex(ctx, database, name, idx)
		}
	}

	// If external APIs are configured, create TTL index for BTC price collection
	if cfg.ExternalAPIs != nil {
		if err := createTTLIndexes(ctx, database, cfg.ExternalAPIs.CoinMarketCap.CacheTTL); err != nil {
			log.Error().Err(err).Msg("Failed to create TTL index for BTC price")
			return err
		}
	}

	log.Info().Msg("Collections and Indexes created successfully.")
	return nil
}

func createCollection(ctx context.Context, database *mongo.Database, collectionName string) {
	// Check if the collection already exists.
	if _, err := database.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{}); err != nil {
		log.Debug().Msg(fmt.Sprintf("Collection maybe already exists: %s, skip the rest. info: %s", collectionName, err))
		return
	}

	// Create the collection.
	if err := database.CreateCollection(ctx, collectionName); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to create collection: " + collectionName)
		return
	}

	log.Debug().Msg("Collection created successfully: " + collectionName)
}

func createIndex(ctx context.Context, database *mongo.Database, collectionName string, idx index) {
	if len(idx.Indexes) == 0 {
		return
	}

	indexKeys := bson.D{}
	for k, v := range idx.Indexes {
		indexKeys = append(indexKeys, bson.E{Key: k, Value: v})
	}

	index := mongo.IndexModel{
		Keys:    indexKeys,
		Options: options.Index().SetUnique(idx.Unique),
	}

	if _, err := database.Collection(collectionName).Indexes().CreateOne(ctx, index); err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to create index on collection '%s': %v", collectionName, err))
		return
	}

	log.Debug().Msg("Index created successfully on collection: " + collectionName)
}

func createTTLIndexes(ctx context.Context, database *mongo.Database, cacheTTL time.Duration) error {
	collection := database.Collection(PriceCollection)
	// First, drop the existing TTL index if it exists
	_, err := collection.Indexes().DropOne(ctx, "created_at_1")
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to drop existing TTL index: %w", err)
	}
	// Create new TTL index
	model := mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: 1}},
		Options: options.Index().
			SetExpireAfterSeconds(int32(cacheTTL.Seconds())).
			SetName("created_at_1"),
	}
	_, err = collection.Indexes().CreateOne(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to create TTL index: %w", err)
	}
	return nil
}
