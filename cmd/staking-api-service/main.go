package main

import (
	"context"
	"fmt"

	"github.com/babylonlabs-io/staking-api-service/cmd/staking-api-service/cli"
	"github.com/babylonlabs-io/staking-api-service/cmd/staking-api-service/scripts"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	dbclients "github.com/babylonlabs-io/staking-api-service/internal/shared/db/clients"
	dbmodel "github.com/babylonlabs-io/staking-api-service/internal/shared/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/http/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/healthcheck"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	queueclients "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Debug().Msg("failed to load .env file")
	}
}

// @title           Babylon Staking API
// @version         1.0
// @description     The Babylon Staking API offers information about the state of the Phase-1 BTC Staking system.
// @description     Your access and use is governed by the API Access License linked to below.
// @license.name    API Access License
// @license.url     https://docs.babylonlabs.io/assets/files/api-access-license.pdf
// @contact.email   contact@babylonlabs.io
func main() {
	ctx := context.Background()

	// setup cli commands and flags
	if err := cli.Setup(); err != nil {
		log.Fatal().Err(err).Msg("error while setting up cli")
	}

	// load config
	cfgPath := cli.GetConfigPath()
	cfg, err := config.New(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg(fmt.Sprintf("error while loading config file: %s", cfgPath))
	}

	paramsPath := cli.GetGlobalParamsPath()
	params, err := types.NewGlobalParams(paramsPath)
	if err != nil {
		log.Fatal().Err(err).Msg(fmt.Sprintf("error while loading global params file: %s", paramsPath))
	}

	finalityProvidersPath := cli.GetFinalityProvidersPath()
	finalityProviders, err := types.NewFinalityProviders(finalityProvidersPath)
	if err != nil {
		log.Fatal().Err(err).Msg(fmt.Sprintf("error while loading finality providers file: %s", finalityProvidersPath))
	}

	// initialize metrics with the metrics port from config
	metricsPort := cfg.Metrics.GetMetricsPort()
	metrics.Init(metricsPort)

	err = dbmodel.Setup(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("error while setting up staking db model")
	}

	// initialize clients package which is used to interact with external services
	clients := clients.New(cfg)

	dbClients, err := dbclients.New(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("error while setting up staking db clients")
	}

	services, err := services.New(ctx, cfg, params, finalityProviders, clients, dbClients)
	if err != nil {
		log.Fatal().Err(err).Msg("error while setting up staking services layer")
	}

	// Start the event queue processing
	queueClients := queueclients.New(ctx, cfg.Queue, services)

	// Check if the scripts flag is set
	if cli.GetReplayFlag() {
		log.Info().Msg("Replay flag is set. Starting replay of unprocessable messages.")

		err := scripts.ReplayUnprocessableMessages(ctx, cfg, queueClients, dbClients.SharedDBClient)
		if err != nil {
			log.Fatal().Err(err).Msg("error while replaying unprocessable messages")
		}
		return
	} else if cli.GetBackfillPubkeyAddressFlag() {
		log.Info().Msg("Backfill pubkey address flag is set. Starting backfill of pubkey address mappings.")
		err := scripts.BackfillPubkeyAddressesMappings(ctx, cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("error while backfilling pubkey address mappings")
		}
		return
	}

	queueClients.StartReceivingMessages()

	healthcheckErr := healthcheck.StartHealthCheckCron(ctx, queueClients, cfg.Server.HealthCheckInterval)
	if healthcheckErr != nil {
		log.Fatal().Err(healthcheckErr).Msg("error while starting health check cron")
	}

	apiServer, err := api.New(ctx, cfg, services)
	if err != nil {
		log.Fatal().Err(err).Msg("error while setting up staking api service")
	}
	if err = apiServer.Start(); err != nil {
		log.Fatal().Err(err).Msg("error while starting staking api service")
	}
}
