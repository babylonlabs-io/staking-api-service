package main

import (
	"context"
	"fmt"

	"github.com/babylonlabs-io/staking-api-service/cmd/staking-api-service/cli"
	"github.com/babylonlabs-io/staking-api-service/cmd/staking-api-service/scripts"
	"github.com/babylonlabs-io/staking-api-service/internal/api"
	"github.com/babylonlabs-io/staking-api-service/internal/clients"
	"github.com/babylonlabs-io/staking-api-service/internal/config"
	"github.com/babylonlabs-io/staking-api-service/internal/db/model"
	"github.com/babylonlabs-io/staking-api-service/internal/observability/healthcheck"
	"github.com/babylonlabs-io/staking-api-service/internal/observability/metrics"
	"github.com/babylonlabs-io/staking-api-service/internal/queue"
	"github.com/babylonlabs-io/staking-api-service/internal/services"
	"github.com/babylonlabs-io/staking-api-service/internal/types"
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
// @description     Your access and use is governed by the Terms of Service listed below.
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

	err = model.Setup(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("error while setting up staking db model")
	}

	// initialize clients package which is used to interact with external services
	clients := clients.New(cfg)
	services, err := services.New(ctx, cfg, params, finalityProviders, clients)
	if err != nil {
		log.Fatal().Err(err).Msg("error while setting up staking services layer")
	}
	// Start the event queue processing
	queues := queue.New(cfg.Queue, services)

	// Check if the replay flag is set
	if cli.GetReplayFlag() {
		log.Info().Msg("Replay flag is set. Starting replay of unprocessable messages.")
		err := scripts.ReplayUnprocessableMessages(ctx, cfg, queues, services.DbClient)
		if err != nil {
			log.Fatal().Err(err).Msg("error while replaying unprocessable messages")
		}
		return
	}

	queues.StartReceivingMessages()

	healthcheck.StartHealthCheckCron(ctx, queues, cfg.Server.HealthCheckInterval)

	apiServer, err := api.New(ctx, cfg, services)
	if err != nil {
		log.Fatal().Err(err).Msg("error while setting up staking api service")
	}
	if err = apiServer.Start(); err != nil {
		log.Fatal().Err(err).Msg("error while starting staking api service")
	}
}
