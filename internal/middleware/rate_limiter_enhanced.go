// internal/middleware/rate_limiter_enhanced.go - Enhanced rate limiter excluding health/metrics
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
)

// List of endpoints to exclude from rate limit tracking
var excludedEndpoints = map[string]bool{
	"/health":      true,
	"/metrics":     true,
	"/api/health":  true,
	"/api/metrics": true,
	"/_health":     true,
	"/_metrics":    true,
	"/healthz":     true,
	"/readiness":   true,
	"/liveness":    true,
	"/status":      true,
	"/ping":        true,
}

// shouldTrackEndpoint determines if an endpoint should be tracked in rate limits
func shouldTrackEndpoint(endpoint string) bool {
	// Remove query parameters for matching
	if idx := strings.Index(endpoint, "?"); idx != -1 {
		endpoint = endpoint[:idx]
	}

	// Check if endpoint is in excluded list
	if excludedEndpoints[endpoint] {
		return false
	}

	// Check if endpoint starts with excluded prefix
	for excluded := range excludedEndpoints {
		if strings.HasPrefix(endpoint, excluded) {
			return false
		}
	}

	return true
}

// EnhancedPersistentRateLimitMiddleware with exclusions
func EnhancedPersistentRateLimitMiddleware(queries *db.Queries, config RateLimitConfig) echo.MiddlewareFunc {
	limiter := NewPersistentRateLimiter(queries, config)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientIP := c.RealIP()
			endpoint := c.Path()
			ctx := c.Request().Context()

			// Skip rate limiting for excluded endpoints
			if !shouldTrackEndpoint(endpoint) {
				return next(c)
			}

			// Check in-memory rate limit first (fast path)
			inMemLimiter := limiter.GetLimiter(clientIP)
			if !inMemLimiter.Allow() {
				// Record the rate limit hit (only for trackable endpoints)
				limiter.RecordRequest(ctx, clientIP, endpoint, false)
				RecordRateLimitExceeded(endpoint)

				return echo.NewHTTPError(http.StatusTooManyRequests,
					"Rate limit exceeded. Please try again later.")
			}

			// For login endpoint, also check DB
			if endpoint == "/api/v1/auth/login" {
				allowed, err := limiter.CheckRateLimit(ctx, clientIP, endpoint,
					int32(config.LoginMaxAttempts))
				if err != nil {
					c.Logger().Error("Rate limit DB check failed:", err)
				} else if !allowed {
					limiter.RecordRequest(ctx, clientIP, endpoint, false)
					return echo.NewHTTPError(http.StatusTooManyRequests,
						"Too many login attempts. Please try again later.")
				}
			}

			// Record successful request (only for trackable endpoints)
			limiter.RecordRequest(ctx, clientIP, endpoint, true)

			return next(c)
		}
	}
}

// CleanupMiddleware with exclusion support
func CleanupRateLimitsMiddleware(queries *db.Queries) echo.MiddlewareFunc {
	// Run cleanup periodically
	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		for range ticker.C {
			ctx := context.Background()
			cutoff := time.Now().Add(-24 * time.Hour)

			// Delete old rate limits excluding health/metrics
			err := queries.DeleteOldRateLimitsExcludingHealthMetrics(ctx, cutoff)
			if err != nil {
				// Log error but don't crash
				return
			}
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
}

// GetRateLimitInfo returns current rate limit info for a client (debugging)
func GetRateLimitInfo(c echo.Context, queries *db.Queries) map[string]interface{} {
	ctx := c.Request().Context()
	clientIP := c.RealIP()
	endpoint := c.Path()

	// Get rate limit records
	windowStart := time.Now().Truncate(1 * time.Minute)
	limit, _ := queries.GetRateLimitByWindow(ctx, db.GetRateLimitByWindowParams{
		ClientID:    clientIP,
		Endpoint:    endpoint,
		WindowStart: windowStart,
	})

	return map[string]interface{}{
		"client_ip":      clientIP,
		"endpoint":       endpoint,
		"requests_count": limit.RequestsCount,
		"window_start":   limit.WindowStart,
		"is_tracked":     shouldTrackEndpoint(endpoint),
	}
}
