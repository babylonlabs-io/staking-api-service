package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type Outcome string

const (
	Success                  Outcome       = "success"
	Error                    Outcome       = "error"
	MetricRequestTimeout     time.Duration = 5 * time.Second
	MetricRequestIdleTimeout time.Duration = 10 * time.Second
)

func (O Outcome) String() string {
	return string(O)
}

var (
	once                             sync.Once
	metricsRouter                    *chi.Mux
	httpRequestDurationHistogram     *prometheus.HistogramVec
	eventProcessingDurationHistogram *prometheus.HistogramVec
	unprocessableEntityCounter       *prometheus.CounterVec
	queueOperationFailureCounter     *prometheus.CounterVec
	httpResponseWriteFailureCounter  *prometheus.CounterVec
	clientRequestDurationHistogram   *prometheus.HistogramVec
	serviceCrashCounter              *prometheus.CounterVec
	dbErrorsCounter                  *prometheus.CounterVec
)

// Init initializes the metrics package.
func Init(metricsPort int) {
	once.Do(func() {
		initMetricsRouter(metricsPort)
		registerMetrics()
	})
}

// initMetricsRouter initializes the metrics router.
func initMetricsRouter(metricsPort int) {
	metricsRouter = chi.NewRouter()
	metricsRouter.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})
	// Create a custom server with timeout settings
	metricsAddr := fmt.Sprintf(":%d", metricsPort)
	server := &http.Server{
		Addr:         metricsAddr,
		Handler:      metricsRouter,
		ReadTimeout:  MetricRequestTimeout,
		WriteTimeout: MetricRequestTimeout,
		IdleTimeout:  MetricRequestIdleTimeout,
	}

	// Start the server in a separate goroutine
	go func() {
		log.Printf("Starting metrics server on %s", metricsAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msgf("Error starting metrics server on %s", metricsAddr)
		}
	}()
}

// registerMetrics initializes and register the Prometheus metrics.
func registerMetrics() {
	defaultHistogramBucketsSeconds := []float64{0.1, 0.5, 1, 2.5, 5, 10, 30}

	httpRequestDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of http request durations in seconds.",
			Buckets: defaultHistogramBucketsSeconds,
		},
		[]string{"endpoint", "status"},
	)

	eventProcessingDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "event_processing_duration_seconds",
			Help:    "Histogram of event processing durations in seconds.",
			Buckets: defaultHistogramBucketsSeconds,
		},
		[]string{"queuename", "status", "attempts"},
	)

	unprocessableEntityCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "unprocessable_entity_total",
			Help: "Total number of unprocessable entities from the event processing.",
		},
		[]string{"entity"},
	)

	queueOperationFailureCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queue_operation_failure_total",
			Help: "Total number of failed queue operations per queue name.",
		},
		[]string{"operation", "queuename"},
	)

	httpResponseWriteFailureCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_write_failure_total",
			Help: "Total number of failed http response writes.",
		},
		[]string{"status"},
	)

	// client requests are the ones sending to other service
	clientRequestDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "client_request_duration_seconds",
			Help:    "Histogram of outgoing client request durations in seconds.",
			Buckets: defaultHistogramBucketsSeconds,
		},
		[]string{"baseurl", "method", "path", "status"},
	)
	serviceCrashCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_crash_total",
			Help: "",
		},
		[]string{"type"},
	)
	dbErrorsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_errors",
			Help: "",
		},
		[]string{"method"},
	)

	prometheus.MustRegister(
		httpRequestDurationHistogram,
		eventProcessingDurationHistogram,
		unprocessableEntityCounter,
		queueOperationFailureCounter,
		httpResponseWriteFailureCounter,
		clientRequestDurationHistogram,
		serviceCrashCounter,
	)
}

// StartHttpRequestDurationTimer starts a timer to measure http request handling duration.
func StartHttpRequestDurationTimer(endpoint string) func(statusCode int) {
	startTime := time.Now()
	return func(statusCode int) {
		duration := time.Since(startTime).Seconds()
		httpRequestDurationHistogram.WithLabelValues(
			endpoint,
			fmt.Sprintf("%d", statusCode),
		).Observe(duration)
	}
}

func StartEventProcessingDurationTimer(queuename string, attempts int32) func(statusCode int) {
	startTime := time.Now()
	return func(statusCode int) {
		duration := time.Since(startTime).Seconds()
		eventProcessingDurationHistogram.WithLabelValues(
			queuename,
			fmt.Sprintf("%d", statusCode),
			fmt.Sprintf("%d", attempts),
		).Observe(duration)
	}
}

// RecordUnprocessableEntity increments the unprocessable entity counter.
// This is basically the number of items will show up in the unprocessable entity collection
func RecordUnprocessableEntity(entity string) {
	unprocessableEntityCounter.WithLabelValues(entity).Inc()
}

// RecordQueueOperationFailure increments the queue operation failure counter.
func RecordQueueOperationFailure(operation, queuename string) {
	queueOperationFailureCounter.WithLabelValues(operation, queuename).Inc()
}

// RecordHttpResponseWriteFailure increments the http response write failure counter.
func RecordHttpResponseWriteFailure(statusCode int) {
	httpResponseWriteFailureCounter.WithLabelValues(fmt.Sprintf("%d", statusCode)).Inc()
}

// StartClientRequestDurationTimer starts a timer to measure outgoing client request duration.
func StartClientRequestDurationTimer(baseUrl, method, path string) func(statusCode int) {
	startTime := time.Now()
	return func(statusCode int) {
		duration := time.Since(startTime).Seconds()
		clientRequestDurationHistogram.WithLabelValues(
			baseUrl,
			method,
			path,
			fmt.Sprintf("%d", statusCode),
		).Observe(duration)
	}
}

// RecordServiceCrash increments the service crash counter.
func RecordServiceCrash(service string) {
	serviceCrashCounter.WithLabelValues(service).Inc()
}

func RecordDbError(method string) {
	dbErrorsCounter.WithLabelValues(method).Inc()
}
