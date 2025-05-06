//go:build e2e

package api

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/babylonlabs-io/babylon-staking-indexer/testutil"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclient "github.com/babylonlabs-io/staking-api-service/internal/shared/db/client"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/babylonlabs-io/staking-api-service/pkg"
	"github.com/ory/dockertest"
	dc "github.com/ory/dockertest/docker"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

const requestTimeout = 3 * time.Second

var apiURL string

func TestMain(t *testing.M) {
	ctx := context.Background()

	db, err := setupDB(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to setup DB")
	}

	cfg := &config.Config{
		StakingDb: db.stakingConfig,
		IndexerDb: db.indexerConfig,
		Server: &config.ServerConfig{
			LogLevel:         "info",
			MaxContentLength: 4096,
			BTCNet:           "mainnet",
			// these values are not used in tests, but we keep them here so validation pass
			Host:                "127.0.0.1",
			Port:                9999,
			HealthCheckInterval: 1,
		},
		AddressScreeningConfig: &config.AddressScreeningConfig{
			Enabled: true,
		},
	}

	s, err := setupServices(ctx, cfg)
	if err != nil {
		db.cleanup()
		log.Fatal().Err(err).Msg("Failed to setup services")
	}

	srv, err := api.New(ctx, cfg, s)
	if err != nil {
		db.cleanup()
		log.Fatal().Err(err).Msg("Failed to initialize api")
	}

	metrics.Init(7777)

	go srv.Start() //nolint:errcheck
	time.Sleep(time.Second)
	_, port, err := net.SplitHostPort(srv.Addr())
	if err != nil {
		db.cleanup()
		log.Fatal().Err(err).Msg("Failed to parse server address")
	}
	apiURL = "http://localhost:" + port
	defer srv.Stop() //nolint:errcheck

	// running tests
	code := t.Run()
	db.cleanup()
	os.Exit(code)
}

func setupServices(ctx context.Context, cfg *config.Config) (*services.Services, error) {
	err := cfg.Server.Validate()
	if err != nil {
		return nil, err
	}

	dbClients, err := dbclients.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	fp, err := types.NewFinalityProviders("testdata/finality-providers.json")
	if err != nil {
		return nil, err
	}

	globals, err := types.NewGlobalParams("testdata/global-params.json")
	if err != nil {
		return nil, err
	}

	clients := clients.New(cfg)
	return services.New(cfg, globals, fp, clients, dbClients)
}

type db struct {
	stakingConfig *config.DbConfig
	indexerConfig *config.DbConfig
	cleanup       func()
}

func setupDB(ctx context.Context) (*db, error) {
	mongoCfg, cleanup, err := setupMongoContainer()
	if err != nil {
		return nil, err
	}

	stakingConfig := createDbConfig(mongoCfg, "api")
	indexerConfig := createDbConfig(mongoCfg, "indexer")

	err = dbmodel.Setup(ctx, stakingConfig, nil)
	if err != nil {
		cleanup()
		return nil, err
	}

	err = loadTestdata(ctx, stakingConfig, indexerConfig)
	if err != nil {
		cleanup()
		return nil, err
	}

	return &db{
		stakingConfig: stakingConfig,
		indexerConfig: indexerConfig,
		cleanup:       cleanup,
	}, nil
}

func loadTestdata(ctx context.Context, configs ...*config.DbConfig) error {
	loadDb := func(cfg *config.DbConfig) error {
		client, err := dbclient.NewMongoClient(ctx, cfg)
		if err != nil {
			return err
		}
		db := client.Database(cfg.DbName)

		pattern := fmt.Sprintf("testdata/%s/*.json", cfg.DbName)
		files, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}

		for _, file := range files {
			buff, err := os.ReadFile(file)
			if err != nil {
				return err
			}

			var docs []any
			err = bson.UnmarshalExtJSON(buff, true, &docs)
			if err != nil {
				return err
			}

			filename := filepath.Base(file)
			collectionName := strings.TrimSuffix(filename, ".json")
			coll := db.Collection(collectionName)

			_, err = coll.InsertMany(ctx, docs)
			if err != nil {
				return err
			}
		}

		return nil
	}
	for _, cfg := range configs {
		err := loadDb(cfg)
		if err != nil {
			return fmt.Errorf("failed to load %q db: %w", cfg.DbName, err)
		}
	}

	return nil
}

func clientGet(t *testing.T, endpoint string) ([]byte, int) {
	require.True(t, strings.HasPrefix(endpoint, "/"), "endpoint must start with /")
	url := apiURL + endpoint

	ctx, cancel := context.WithTimeout(context.TODO(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := (&http.Client{}).Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	buff, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return buff, resp.StatusCode
}

// setupMongoContainer setups container with mongodb returning db credentials through config.DbConfig, cleanup function
// and an error if any. Cleanup function MUST be called in the end to cleanup docker resources
func setupMongoContainer() (*mongoConfig, func(), error) {
	const (
		mongoUsername = "user"
		mongoPassword = "password"
		// this version corresponds to docker tag for mongodb
		// it should be in sync with mongo version used in production
		mongoVersion = "7.0.5"
	)

	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, err
	}

	// generate random string for container name
	randomString, err := testutil.RandomAlphaNum(5)
	if err != nil {
		return nil, nil, err
	}

	// there can be only 1 container with the same name, so we add
	// random string in the end in case there is still old container running
	containerName := "mongo-integration-tests-db-" + randomString
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       containerName,
		Repository: "mongo",
		Tag:        mongoVersion,
		Env: []string{
			"MONGO_INITDB_ROOT_USERNAME=" + mongoUsername,
			"MONGO_INITDB_ROOT_PASSWORD=" + mongoPassword,
		},
	}, func(config *dc.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = dc.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		err := pool.Purge(resource)
		if err != nil {
			// todo change to log fatal
			panic(err)
		}
	}

	// get host port (randomly chosen) that is mapped to mongo port inside container
	hostPort := resource.GetPort("27017/tcp")

	return &mongoConfig{
		username: mongoUsername,
		password: mongoPassword,
		address:  fmt.Sprintf("mongodb://localhost:%s/", hostPort),
	}, cleanup, nil
}

type mongoConfig struct {
	username string
	password string
	address  string
}

func createDbConfig(cfg *mongoConfig, dbName string) *config.DbConfig {
	return &config.DbConfig{
		Username:           cfg.username,
		Password:           cfg.password,
		DbName:             dbName,
		Address:            cfg.address,
		MaxPaginationLimit: 2,
		LogicalShardCount:  pkg.Ptr[int64](10),
	}
}
