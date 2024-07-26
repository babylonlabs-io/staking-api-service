package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/babylonchain/staking-queue-client/client"
	"github.com/go-chi/chi"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	queueConfig "github.com/babylonchain/staking-queue-client/config"

	"github.com/babylonchain/staking-api-service/internal/api"
	"github.com/babylonchain/staking-api-service/internal/api/middlewares"
	"github.com/babylonchain/staking-api-service/internal/clients"
	"github.com/babylonchain/staking-api-service/internal/config"
	"github.com/babylonchain/staking-api-service/internal/db"
	"github.com/babylonchain/staking-api-service/internal/observability/metrics"
	"github.com/babylonchain/staking-api-service/internal/queue"
	"github.com/babylonchain/staking-api-service/internal/services"
	"github.com/babylonchain/staking-api-service/internal/types"
)

type TestServerDependency struct {
	ConfigOverrides         *config.Config
	MockDbClient            db.DBClient
	PreInjectEventsHandler  func(queueClient client.QueueClient) error
	MockedFinalityProviders []types.FinalityProviderDetails
	MockedGlobalParams      *types.GlobalParams
	MockedClients           *clients.Clients
}

type TestServer struct {
	Server  *httptest.Server
	Queues  *queue.Queues
	Conn    *amqp091.Connection
	channel *amqp091.Channel
	Config  *config.Config
}

func (ts *TestServer) Close() {
	ts.Server.Close()
	ts.Queues.StopReceivingMessages()
	ts.Conn.Close()
	ts.channel.Close()
}

func loadTestConfig(t *testing.T) *config.Config {
	cfg, err := config.New("./config/config-test.yml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}
	return cfg
}

func setupTestServer(t *testing.T, dep *TestServerDependency) *TestServer {
	var err error
	var cfg *config.Config
	if dep != nil && dep.ConfigOverrides != nil {
		cfg = dep.ConfigOverrides
	} else {
		cfg = loadTestConfig(t)
	}
	metricsPort := cfg.Metrics.GetMetricsPort()
	metrics.Init(metricsPort)

	var params *types.GlobalParams
	if dep != nil && dep.MockedGlobalParams != nil {
		params = dep.MockedGlobalParams
	} else {
		params, err = types.NewGlobalParams("./config/global-params-test.json")
		if err != nil {
			t.Fatalf("Failed to load global params: %v", err)
		}
	}

	var fps []types.FinalityProviderDetails
	if dep != nil && dep.MockedFinalityProviders != nil {
		fps = dep.MockedFinalityProviders
	} else {
		fps, err = types.NewFinalityProviders("./config/finality-providers-test.json")
		if err != nil {
			t.Fatalf("Failed to load finality providers: %v", err)
		}
	}

	var c *clients.Clients
	if dep != nil && dep.MockedClients != nil {
		c = dep.MockedClients
	} else {
		c = clients.New(cfg)
	}

	services, err := services.New(context.Background(), cfg, params, fps, c)
	if err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	if dep != nil && dep.MockDbClient != nil {
		services.DbClient = dep.MockDbClient
	} else {
		// This means we are using real database, we not mocking anything
		setupTestDB(*cfg)
	}

	apiServer, err := api.New(context.Background(), cfg, services)
	if err != nil {
		t.Fatalf("Failed to initialize API server: %v", err)
	}

	// Setup routes
	r := chi.NewRouter()

	r.Use(middlewares.CorsMiddleware(cfg))
	r.Use(middlewares.SecurityHeadersMiddleware())
	r.Use(middlewares.ContentLengthMiddleware(cfg))
	apiServer.SetupRoutes(r)

	queues, conn, ch, err := setUpTestQueue(cfg.Queue, services)
	if err != nil {
		t.Fatalf("Failed to setup test queue: %v", err)
	}

	// Create an httptest server
	server := httptest.NewServer(r)

	return &TestServer{
		Server:  server,
		Queues:  queues,
		Conn:    conn,
		channel: ch,
		Config:  cfg,
	}
}

// PurgeAllCollections drops all collections in the specified database.
func PurgeAllCollections(ctx context.Context, client *mongo.Client, databaseName string) error {
	database := client.Database(databaseName)
	collections, err := database.ListCollectionNames(ctx, bson.D{{}})
	if err != nil {
		return err
	}

	for _, collection := range collections {
		// Use DeleteMany with an empty filter to delete all documents
		_, err := database.Collection(collection).DeleteMany(ctx, bson.D{{}})
		if err != nil {
			return err
		}
	}
	return nil
}

// setupTestDB connects to MongoDB and purges all collections.
func setupTestDB(cfg config.Config) *mongo.Client {
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.Db.Address))
	if err != nil {
		log.Fatal(err)
	}

	// Purge all collections in the test database
	if err := PurgeAllCollections(context.TODO(), client, cfg.Db.DbName); err != nil {
		log.Fatal("Failed to purge database:", err)
	}

	return client
}

func setUpTestQueue(cfg *queueConfig.QueueConfig, service *services.Services) (*queue.Queues, *amqp091.Connection, *amqp091.Channel, error) {
	amqpURI := fmt.Sprintf("amqp://%s:%s@%s", cfg.QueueUser, cfg.QueuePassword, cfg.Url)
	conn, err := amqp091.Dial(amqpURI)
	if err != nil {
		log.Fatal("failed to connect to RabbitMQ in test: ", err)
		return nil, nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open a channel in test: %w", err)
	}
	purgeError := purgeQueues(ch, []string{
		client.ActiveStakingQueueName,
		client.UnbondingStakingQueueName,
		client.WithdrawStakingQueueName,
		client.ExpiredStakingQueueName,
		client.StakingStatsQueueName,
		// purge delay queues as well
		client.ActiveStakingQueueName + "_delay",
		client.UnbondingStakingQueueName + "_delay",
		client.WithdrawStakingQueueName + "_delay",
		client.ExpiredStakingQueueName + "_delay",
		client.StakingStatsQueueName + "_delay",
	})
	if purgeError != nil {
		log.Fatal("failed to purge queues in test: ", purgeError)
		return nil, nil, nil, purgeError
	}

	// Start the actual queue processing in our codebase
	queues := queue.New(cfg, service)
	queues.StartReceivingMessages()

	return queues, conn, ch, nil
}

// inspectQueueMessageCount inspects the number of messages in the given queue.
func inspectQueueMessageCount(t *testing.T, conn *amqp091.Connection, queueName string) (int, error) {
	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("failed to open a channel in test: %v", err)
	}
	q, err := ch.QueueInspect(queueName)
	if err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") || strings.Contains(err.Error(), "channel/connection is not open") {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to inspect queue in test %s: %w", queueName, err)
	}
	return q.Messages, nil
}

// purgeQueues purges all messages from the given list of queues.
func purgeQueues(ch *amqp091.Channel, queues []string) error {
	for _, queue := range queues {
		_, err := ch.QueuePurge(queue, false)
		if err != nil {
			if strings.Contains(err.Error(), "NOT_FOUND") || strings.Contains(err.Error(), "channel/connection is not open") {
				continue
			}
			return fmt.Errorf("failed to purge queue in test %s: %w", queue, err)
		}
	}

	return nil
}

func sendTestMessage[T any](client client.QueueClient, data []T) error {
	for _, d := range data {
		jsonBytes, err := json.Marshal(d)
		if err != nil {
			return err
		}
		messageBody := string(jsonBytes)
		err = client.SendMessage(context.TODO(), messageBody)
		if err != nil {
			return fmt.Errorf("failed to publish a message to queue %s: %w", client.GetQueueName(), err)
		}
	}
	return nil
}

func directDbConnection(t *testing.T) *db.Database {
	cfg, err := config.New("./config/config-test.yml")
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.Db.Address))
	if err != nil {
		log.Fatal(err)
	}
	return &db.Database{
		DbName: cfg.Db.DbName,
		Client: client,
	}
}

func injectDbDocuments[T any](t *testing.T, collectionName string, doc T) {
	connection := directDbConnection(t)
	collection := connection.Client.Database(connection.DbName).Collection(collectionName)

	_, err := collection.InsertOne(context.Background(), doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}
}

// Inspect the items in the real database
func inspectDbDocuments[T any](t *testing.T, collectionName string) ([]T, error) {
	connection := directDbConnection(t)
	collection := connection.Client.Database(connection.DbName).Collection(collectionName)

	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []T
	for cursor.Next(context.Background()) {
		var result T
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func buildActiveStakingEvent(t *testing.T, numOfEvenet int) []*client.ActiveStakingEvent {
	var activeStakingEvents []*client.ActiveStakingEvent
	stakerPk, err := randomPk()
	require.NoError(t, err)
	// To be replaced with https://github.com/babylonchain/staking-api-service/issues/21
	rand.New(rand.NewSource(time.Now().Unix()))

	for i := 0; i < numOfEvenet; i++ {
		activeStakingEvent := &client.ActiveStakingEvent{
			EventType:             client.ActiveStakingEventType,
			StakingTxHashHex:      "0x1234567890abcdef" + fmt.Sprint(i),
			StakerPkHex:           stakerPk,
			FinalityProviderPkHex: "0xabcdef1234567890" + fmt.Sprint(i),
			StakingValue:          uint64(rand.Intn(1000)),
			StakingStartHeight:    uint64(rand.Intn(200)),
			StakingStartTimestamp: time.Now().Unix(),
			StakingTimeLock:       uint64(rand.Intn(100)),
			StakingOutputIndex:    uint64(rand.Intn(100)),
			StakingTxHex:          "0xabcdef1234567890" + fmt.Sprint(i),
			IsOverflow:            false,
		}
		activeStakingEvents = append(activeStakingEvents, activeStakingEvent)
	}
	return activeStakingEvents
}


