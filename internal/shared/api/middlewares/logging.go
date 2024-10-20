package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/tracing"
	"github.com/rs/zerolog/log"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request path starts with /swagger/ or is /healthcheck
		if strings.HasPrefix(r.URL.Path, "/swagger/") || r.URL.Path == "/healthcheck" || r.URL.Path == "/" {
			// If it does, skip logging and serve the swagger request
			next.ServeHTTP(w, r)
			return
		}

		startTime := time.Now()
		logger := log.With().Str("path", r.URL.Path).Logger()

		// Attach traceId into each log within the request chain
		traceId := r.Context().Value(tracing.TraceIdKey)
		if traceId != nil {
			logger = logger.With().Interface("traceId", traceId).Logger()
		}

		logger.Debug().Msg("request received")
		r = r.WithContext(logger.WithContext(r.Context()))

		next.ServeHTTP(w, r)

		requestDuration := time.Since(startTime).Milliseconds()
		logEvent := logger.Debug()

		tracingInfo := r.Context().Value(tracing.TracingInfoKey)
		if tracingInfo != nil {
			logEvent = logEvent.Interface("tracingInfo", tracingInfo)
		}

		logEvent.Interface("requestDuration", requestDuration).Msg("Request completed")
	})
}
