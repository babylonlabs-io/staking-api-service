package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	bbndatagen "github.com/babylonlabs-io/babylon/testutil/datagen"
	"github.com/babylonlabs-io/staking-queue-client/client"
	"github.com/go-chi/chi"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"

	queueConfig "github.com/babylonlabs-io/staking-queue-client/config"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handler"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/middlewares"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/services"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	queueclients "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/tests/testutils"
)

var setUpDbIndex = false

type TestServerDependency struct {
	ConfigOverrides         *config.Config
	MockDbClients           dbclients.DbClients
	PreInjectEventsHandler  func(queueClient client.QueueClient) error
	MockedFinalityProviders []types.FinalityProviderDetails
	MockedGlobalParams      *types.GlobalParams
	MockedClients           *clients.Clients
}

type TestServer struct {
	Server  *httptest.Server
	Queues  *queueclients.QueueClients
	Conn    *amqp091.Connection
	channel *amqp091.Channel
	Config  *config.Config
	Db      *mongo.Client
}

func (ts *TestServer) Close() {
	ts.Server.Close()
	ts.Queues.V1QueueClient.StopReceivingMessages()
	if err := ts.Conn.Close(); err != nil {
		log.Fatalf("failed to close connection in test: %v", err)
	}
	if err := ts.channel.Close(); err != nil {
		log.Fatalf("failed to close channel in test: %v", err)
	}
	if err := ts.Db.Disconnect(context.Background()); err != nil {
		log.Fatalf("failed to close db connection in test: %v", err)
	}
}

func loadTestConfig(t *testing.T) *config.Config {
	cfg, err := config.New("../config/config-test.yml")
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
		cfg = testutils.LoadTestConfig()
	}
	metricsPort := cfg.Metrics.GetMetricsPort()
	metrics.Init(metricsPort)

	var params *types.GlobalParams
	if dep != nil && dep.MockedGlobalParams != nil {
		params = dep.MockedGlobalParams
	} else {
		params, err = types.NewGlobalParams("../config/global-params-test.json")
		if err != nil {
			t.Fatalf("Failed to load global params: %v", err)
		}
	}

	var fps []types.FinalityProviderDetails
	if dep != nil && dep.MockedFinalityProviders != nil {
		fps = dep.MockedFinalityProviders
	} else {
		fps, err = types.NewFinalityProviders("../config/finality-providers-test.json")
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

	// setup test db
	dbClients := testutils.SetupTestDB(*cfg)

	if dep != nil && (dep.MockDbClients.V1DBClient != nil || dep.MockDbClients.V2DBClient != nil || dep.MockDbClients.MongoClient != nil) {
		dbClients = &dep.MockDbClients
	}

	services, err := services.New(context.Background(), cfg, params, fps, c, dbClients)
	if err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
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
		Db:      dbClients.MongoClient,
	}
}

func setUpTestQueue(cfg *queueConfig.QueueConfig, services *services.Services) (*queueclients.QueueClients, *amqp091.Connection, *amqp091.Channel, error) {
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
	queueClients := queueclients.New(context.Background(), cfg, services)
	queueClients.StartReceivingMessages()

	return queueClients, conn, ch, nil
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

// TODO: To be removed and use the method from testutils
func buildActiveStakingEvent(t *testing.T, numOfEvenet int) []*client.ActiveStakingEvent {
	var activeStakingEvents []*client.ActiveStakingEvent
	stakerPk, err := testutils.RandomPk()
	require.NoError(t, err)
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

func attachRandomSeedsToFuzzer(f *testing.F, numOfSeeds int) {
	bbndatagen.AddRandomSeedsToFuzzer(f, uint(numOfSeeds))
}

func fetchSuccessfulResponse[T any](t *testing.T, url string) handler.PublicResponse[T] {
	// Make a GET request to the finality providers endpoint
	resp, err := http.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var responseBody handler.PublicResponse[T]
	err = json.Unmarshal(bodyBytes, &responseBody)
	assert.NoError(t, err, "unmarshalling response body should not fail")
	return responseBody
}
