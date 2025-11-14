package middleware

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// RequestLogger provides detailed request logging
type RequestLogger struct {
	logger echo.Logger
}

// NewRequestLogger creates a new request logger
func NewRequestLogger(logger echo.Logger) *RequestLogger {
	return &RequestLogger{logger: logger}
}

// LogRequest logs detailed request information
func (rl *RequestLogger) LogRequest(c echo.Context, start time.Time, err error) {
	req := c.Request()
	res := c.Response()

	// Calculate request duration
	duration := time.Since(start)

	// Get user info if available
	userID := "anonymous"
	username := "anonymous"
	if uid, ok := c.Get("user_id").(string); ok {
		userID = uid
	}
	if uname, ok := c.Get("username").(string); ok {
		username = uname
	}

	// Determine log level based on status code
	status := res.Status
	var logLevel log.Lvl
	switch {
	case status >= 500:
		logLevel = log.ERROR
	case status >= 400:
		logLevel = log.WARN
	default:
		logLevel = log.INFO
	}

	// Build log message
	msg := fmt.Sprintf(
		"[%s] %s %s | Status: %d | Duration: %v | IP: %s | User: %s (%s) | Size: %d bytes",
		req.Method,
		req.RequestURI,
		req.Proto,
		status,
		duration,
		c.RealIP(),
		username,
		userID,
		res.Size,
	)

	// Log error if present
	if err != nil {
		msg += fmt.Sprintf(" | Error: %v", err)
	}

	// Log with appropriate level
	switch logLevel {
	case log.ERROR:
		rl.logger.Error(msg)
	case log.WARN:
		rl.logger.Warn(msg)
	default:
		rl.logger.Info(msg)
	}
}

// Middleware returns the logging middleware
func (rl *RequestLogger) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			rl.LogRequest(c, start, err)

			return err
		}
	}
}

// MetricsCollector collects API metrics
type MetricsCollector struct {
	totalRequests    int64
	totalErrors      int64
	totalDuration    time.Duration
	requestsByMethod map[string]int64
	requestsByPath   map[string]int64
	requestsByStatus map[int]int64
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requestsByMethod: make(map[string]int64),
		requestsByPath:   make(map[string]int64),
		requestsByStatus: make(map[int]int64),
	}
}

// RecordRequest records metrics for a request
func (mc *MetricsCollector) RecordRequest(c echo.Context, duration time.Duration, status int) {
	mc.totalRequests++
	mc.totalDuration += duration

	if status >= 400 {
		mc.totalErrors++
	}

	mc.requestsByMethod[c.Request().Method]++
	mc.requestsByPath[c.Path()]++
	mc.requestsByStatus[status]++
}

// GetMetrics returns current metrics
func (mc *MetricsCollector) GetMetrics() map[string]any {
	avgDuration := time.Duration(0)
	if mc.totalRequests > 0 {
		avgDuration = mc.totalDuration / time.Duration(mc.totalRequests)
	}

	return map[string]any{
		"total_requests":      mc.totalRequests,
		"total_errors":        mc.totalErrors,
		"error_rate":          float64(mc.totalErrors) / float64(mc.totalRequests) * 100,
		"average_duration_ms": avgDuration.Milliseconds(),
		"requests_by_method":  mc.requestsByMethod,
		"requests_by_path":    mc.requestsByPath,
		"requests_by_status":  mc.requestsByStatus,
	}
}

// Middleware returns the metrics collection middleware
func (mc *MetricsCollector) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status

			mc.RecordRequest(c, duration, status)

			return err
		}
	}
}
