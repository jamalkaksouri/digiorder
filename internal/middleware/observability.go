// internal/middleware/observability.go - Complete Observability System

package middleware

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being served",
		},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{100, 1000, 5000, 10000, 50000, 100000},
		},
		[]string{"method", "endpoint"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 1000, 5000, 10000, 50000, 100000, 500000},
		},
		[]string{"method", "endpoint", "status"},
	)

	// Database metrics
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	// Authentication metrics
	authAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"status"},
	)

	authTokensActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_tokens_active",
			Help: "Number of active authentication tokens",
		},
	)

	// Cache metrics
	cacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	cacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	cacheEntriesTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cache_entries_total",
			Help: "Total number of entries in cache",
		},
	)

	// Rate limiting metrics
	rateLimitExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded events",
		},
		[]string{"endpoint"},
	)

	// Business metrics
	ordersCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created",
		},
		[]string{"status"},
	)

	productsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_created_total",
			Help: "Total number of products created",
		},
	)

	usersActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "users_active_total",
			Help: "Number of active users",
		},
	)

	// Error metrics
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "endpoint"},
	)
)

// PrometheusMiddleware creates metrics middleware
func PrometheusMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Start timer
			start := time.Now()
			
			// Increment in-flight requests
			httpRequestsInFlight.Inc()
			defer httpRequestsInFlight.Dec()

			// Track request size
			if c.Request().ContentLength > 0 {
				httpRequestSize.WithLabelValues(
					c.Request().Method,
					c.Path(),
				).Observe(float64(c.Request().ContentLength))
			}

			// Generate unique request ID
			requestID := uuid.New().String()
			c.Response().Header().Set("X-Request-ID", requestID)
			c.Set("request_id", requestID)

			// Call next handler
			err := next(c)

			// Calculate duration
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(c.Response().Status)

			// Record metrics
			httpRequestsTotal.WithLabelValues(
				c.Request().Method,
				c.Path(),
				status,
			).Inc()

			httpRequestDuration.WithLabelValues(
				c.Request().Method,
				c.Path(),
				status,
			).Observe(duration)

			// Track response size
			httpResponseSize.WithLabelValues(
				c.Request().Method,
				c.Path(),
				status,
			).Observe(float64(c.Response().Size))

			// Track errors
			if c.Response().Status >= 400 {
				errorType := "client_error"
				if c.Response().Status >= 500 {
					errorType = "server_error"
				}
				errorsTotal.WithLabelValues(errorType, c.Path()).Inc()
			}

			return err
		}
	}
}

// RecordAuthAttempt records authentication attempt
func RecordAuthAttempt(success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	authAttemptsTotal.WithLabelValues(status).Inc()
}

// RecordCacheHit records cache hit
func RecordCacheHit() {
	cacheHitsTotal.Inc()
}

// RecordCacheMiss records cache miss
func RecordCacheMiss() {
	cacheMissesTotal.Inc()
}

// UpdateCacheSize updates cache size metric
func UpdateCacheSize(size int) {
	cacheEntriesTotal.Set(float64(size))
}

// RecordRateLimitExceeded records rate limit exceeded
func RecordRateLimitExceeded(endpoint string) {
	rateLimitExceeded.WithLabelValues(endpoint).Inc()
}

// RecordOrderCreated records order creation
func RecordOrderCreated(status string) {
	ordersCreated.WithLabelValues(status).Inc()
}

// RecordProductCreated records product creation
func RecordProductCreated() {
	productsCreated.Inc()
}

// UpdateActiveUsers updates active users count
func UpdateActiveUsers(count int) {
	usersActive.Set(float64(count))
}

// RecordDBQuery records database query metrics
func RecordDBQuery(operation, table string, duration time.Duration) {
	dbQueriesTotal.WithLabelValues(operation, table).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// UpdateDBConnections updates database connection metrics
func UpdateDBConnections(active, idle int) {
	dbConnectionsActive.Set(float64(active))
	dbConnectionsIdle.Set(float64(idle))
}

// TracingMiddleware adds distributed tracing
func TracingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get or create trace ID
			traceID := c.Request().Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = uuid.New().String()
			}

			// Get or create span ID
			spanID := uuid.New().String()

			// Set trace headers
			c.Response().Header().Set("X-Trace-ID", traceID)
			c.Response().Header().Set("X-Span-ID", spanID)

			// Store in context
			c.Set("trace_id", traceID)
			c.Set("span_id", spanID)

			// Continue
			return next(c)
		}
	}
}