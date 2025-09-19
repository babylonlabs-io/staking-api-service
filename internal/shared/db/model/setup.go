package dbmodel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/rs/zerolog/log"

	"github.com/babylonlabs-io/staking-api-service/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// Shared
	PkAddressMappingsCollection = "pk_address_mappings"
	TermsAcceptanceCollection   = "terms_acceptance"
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
	V2StatsLockCollection                 = "v2_stats_lock"
	V2OverallStatsCollection              = "v2_overall_stats"
	V2FinalityProviderStatsCollection     = "v2_finality_providers_stats"
	V2StakerStatsCollection               = "v2_staker_stats"
	V2FinalityProvidersMetadataCollection = "v2_finality_providers_metadata"
	// V1 Overall Stats
	V1OverallStatsSimplifiedCollection = "v1_overall_stats"
	// v3
	BsnStatsCollection = "bsn_stats"
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
	TermsAcceptanceCollection: {{Indexes: map[string]int{}}},
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
	V2StakerStatsCollection: {
		{Indexes: map[string]int{"active_tvl": -1}, Unique: false},
		{Indexes: map[string]int{"active_delegations": 1}, Unique: false},
	},
	V2FinalityProvidersMetadataCollection: {{Indexes: map[string]int{}}},
	// V1 Simplified Stats (for cron job recalculation)
	V1OverallStatsSimplifiedCollection: {{Indexes: map[string]int{}}},
	BsnStatsCollection:                 {{Indexes: map[string]int{}}},
}

func Setup(ctx context.Context, stakingDB *config.DbConfig, externalConfig *config.ExternalAPIsConfig) error {
	credential := options.Credential{
		Username: stakingDB.Username,
		Password: stakingDB.Password,
	}
	clientOps := options.Client().ApplyURI(stakingDB.Address).SetAuth(credential)
	client, err := mongo.Connect(ctx, clientOps)
	if err != nil {
		return err
	}

	// Create a context with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Access a database and create collections.
	database := client.Database(stakingDB.DbName)

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
	if externalConfig != nil {
		if err := createTTLIndexes(ctx, database, PriceCollection, externalConfig.CoinMarketCap.CacheTTL); err != nil {
			log.Error().Err(err).Msg("Failed to create TTL index for BTC price")
			return err
		}
	}

	err = createTTLIndexes(ctx, database, V2FinalityProvidersMetadataCollection, pkg.Day)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create TTL index for v2 finality providers logos")
		return err
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

func createTTLIndexes(ctx context.Context, database *mongo.Database, collectionName string, cacheTTL time.Duration) error {
	const (
		oldIndexName = "created_at_1" // todo remove later
		indexName    = "created_at_ttl"
	)

	collection := database.Collection(collectionName)
	// while we transitioning to other index name we need to keep this one
	collection.Indexes().DropOne(ctx, oldIndexName) //nolint:errcheck
	// First, drop the existing TTL index if it exists
	_, err := collection.Indexes().DropOne(ctx, indexName)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to drop existing TTL index: %w", err)
	}
	// Create new TTL index
	model := mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: 1}},
		Options: options.Index().
			SetExpireAfterSeconds(int32(cacheTTL.Seconds())).
			SetName(indexName),
	}
	_, err = collection.Indexes().CreateOne(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to create TTL index: %w", err)
	}
	return nil
}
