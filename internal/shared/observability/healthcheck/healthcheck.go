package healthcheck

import (
	"context"
	"fmt"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	queueclient "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/client"
	queueclients "github.com/babylonlabs-io/staking-api-service/internal/shared/queue/clients"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger zerolog.Logger = log.Logger

func SetLogger(customLogger zerolog.Logger) {
	logger = customLogger
}

func StartHealthCheckCron(ctx context.Context, queueClients *queueclients.QueueClients, cronTime int) error {
	c := cron.New()
	logger.Info().Msg("Initiated Health Check Cron")

	if cronTime == 0 {
		cronTime = 60
	}

	cronSpec := fmt.Sprintf("@every %ds", cronTime)

	_, err := c.AddFunc(cronSpec, func() {
		queueHealthCheck(queueClients.V1QueueClient)
	})

	if err != nil {
		return err
	}

	c.Start()

	go func() {
		<-ctx.Done()
		logger.Info().Msg("Stopping Health Check Cron")
		c.Stop()
	}()

	return nil
}

func queueHealthCheck(queueClient queueclient.QueueClientInterface) {
	if err := queueClient.IsConnectionHealthy(); err != nil {
		logger.Error().Err(err).Msg("One or more queue connections are not healthy.")
		// Record service unavailable in metrics
		metrics.RecordServiceCrash("queue")
		terminateService()
	}
}

func terminateService() {
	logger.Fatal().Msg("Terminating service due to health check failure.")
}
