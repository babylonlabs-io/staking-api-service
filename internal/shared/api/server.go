package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/handlers"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/api/middlewares"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/config"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/services"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Server struct {
	httpServer *http.Server
	handlers   *handlers.Handlers
	cfg        *config.Config
}

func New(
	ctx context.Context, cfg *config.Config, services *services.Services,
) (*Server, error) {
	r := chi.NewRouter()

	logLevel, err := zerolog.ParseLevel(cfg.Server.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error while parsing log level: %w", err)
	}
	zerolog.SetGlobalLevel(logLevel)

	r.Use(middlewares.CorsMiddleware(cfg))
	r.Use(middlewares.SecurityHeadersMiddleware())
	r.Use(middlewares.TracingMiddleware)
	r.Use(middlewares.LoggingMiddleware)
	r.Use(middlewares.ContentLengthMiddleware(cfg))

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		WriteTimeout: cfg.Server.WriteTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		Handler:      r,
	}

	if err != nil {
		return nil, fmt.Errorf("error while setting up handlers: %w", err)
	}

	handlers, err := handlers.New(ctx, cfg, services)
	if err != nil {
		return nil, fmt.Errorf("error while setting up handlers: %w", err)
	}

	server := &Server{
		httpServer: srv,
		handlers:   handlers,
		cfg:        cfg,
	}
	server.SetupRoutes(r)
	return server, nil
}

func (a *Server) Start() error {
	log.Info().Msgf("Starting server on %s", a.httpServer.Addr)
	return a.httpServer.ListenAndServe()
}
