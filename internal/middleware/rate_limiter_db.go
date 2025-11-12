// internal/middleware/rate_limiter_db.go - Database-backed rate limiter
package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"time"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// PersistentRateLimiter tracks rate limits in database
type PersistentRateLimiter struct {
	queries      *db.Queries
	inMemory     map[string]*rate.Limiter
	mu           sync.RWMutex
	globalRate   rate.Limit
	globalBurst  int
	windowSize   time.Duration
	cleanupTicker *time.Ticker
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	GlobalRPS         int           // Requests per second
	GlobalBurst       int           // Burst capacity
	AuthenticatedRPM  int           // Requests per minute for auth users
	LoginMaxAttempts  int           // Max login attempts per IP
	LoginWindow       time.Duration // Time window for login attempts
	WindowSize        time.Duration // Database record window
}

// DefaultRateLimitConfig returns sensible defaults
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GlobalRPS:        100,
		GlobalBurst:      200,
		AuthenticatedRPM: 1000,
		LoginMaxAttempts: 5,
		LoginWindow:      5 * time.Minute,
		WindowSize:       1 * time.Minute,
	}
}

// NewPersistentRateLimiter creates a rate limiter with DB backing
func NewPersistentRateLimiter(queries *db.Queries, config RateLimitConfig) *PersistentRateLimiter {
	rl := &PersistentRateLimiter{
		queries:      queries,
		inMemory:     make(map[string]*rate.Limiter),
		globalRate:   rate.Limit(config.GlobalRPS),
		globalBurst:  config.GlobalBurst,
		windowSize:   config.WindowSize,
		cleanupTicker: time.NewTicker(5 * time.Minute),
	}

	// Start background cleanup
	go rl.cleanup()

	return rl
}

// GetLimiter returns or creates a limiter for a client
func (rl *PersistentRateLimiter) GetLimiter(clientID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.inMemory[clientID]
	if !exists {
		limiter = rate.NewLimiter(rl.globalRate, rl.globalBurst)
		rl.inMemory[clientID] = limiter
	}

	return limiter
}

// RecordRequest records a rate limit entry in the database
func (rl *PersistentRateLimiter) RecordRequest(ctx context.Context, clientID, endpoint string, allowed bool) error {
	// This is async to not block the request
	go func() {
		// Create new context for async operation
		asyncCtx := context.Background()
		windowStart := time.Now().Truncate(rl.windowSize)
		
		// Try to insert or update
		_, err := rl.queries.GetOrCreateRateLimit(asyncCtx, db.GetOrCreateRateLimitParams{
			ClientID:    clientID,
			Endpoint:    endpoint,
			WindowStart: windowStart,
		})
		
		if err != nil {
			// Log error but don't fail the request
			// In production, send to error tracking service
			return
		}
	}()

	return nil
}

// CheckRateLimit verifies if request is within limits (DB-backed)
func (rl *PersistentRateLimiter) CheckRateLimit(ctx context.Context, clientID, endpoint string, maxRequests int32) (bool, error) {
	windowStart := time.Now().Truncate(rl.windowSize)

	// Get current count from DB
	limit, err := rl.queries.GetRateLimitByWindow(ctx, db.GetRateLimitByWindowParams{
		ClientID:    clientID,
		Endpoint:    endpoint,
		WindowStart: windowStart,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			// First request in this window
			return true, nil
		}
		return false, err
	}

	return limit.RequestsCount < maxRequests, nil
}

// cleanup removes old rate limit records
func (rl *PersistentRateLimiter) cleanup() {
	for range rl.cleanupTicker.C {
		ctx := context.Background()
		cutoff := time.Now().Add(-24 * time.Hour)

		// Clean database
		_ = rl.queries.DeleteOldRateLimits(ctx, cutoff)

		// Clean in-memory
		rl.mu.Lock()
		for id, limiter := range rl.inMemory {
			if limiter.Tokens() == float64(rl.globalBurst) {
				delete(rl.inMemory, id)
			}
		}
		rl.mu.Unlock()
	}
}

// Stop stops the cleanup goroutine
func (rl *PersistentRateLimiter) Stop() {
	rl.cleanupTicker.Stop()
}

// PersistentRateLimitMiddleware creates middleware with DB backing
func PersistentRateLimitMiddleware(queries *db.Queries, config RateLimitConfig) echo.MiddlewareFunc {
	limiter := NewPersistentRateLimiter(queries, config)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get client identifier (IP + User Agent hash for better uniqueness)
			clientIP := c.RealIP()
			c.Request().UserAgent()
			clientID := clientIP // Simple version, can be enhanced
			
			endpoint := c.Path()
			ctx := c.Request().Context()

			// Check in-memory rate limit first (fast path)
			inMemLimiter := limiter.GetLimiter(clientID)
			if !inMemLimiter.Allow() {
				// Record the rate limit hit
				limiter.RecordRequest(ctx, clientID, endpoint, false)
				
				// Record metrics
				RecordRateLimitExceeded(endpoint)
				
				return echo.NewHTTPError(http.StatusTooManyRequests, 
					"Rate limit exceeded. Please try again later.")
			}

			// For critical endpoints, also check DB
			if endpoint == "/api/v1/auth/login" {
				allowed, err := limiter.CheckRateLimit(ctx, clientID, endpoint, 
					int32(config.LoginMaxAttempts))
				if err != nil {
					// Log error but allow request
					c.Logger().Error("Rate limit DB check failed:", err)
				} else if !allowed {
					limiter.RecordRequest(ctx, clientID, endpoint, false)
					return echo.NewHTTPError(http.StatusTooManyRequests, 
						"Too many login attempts. Please try again later.")
				}
			}

			// Record successful request
			limiter.RecordRequest(ctx, clientID, endpoint, true)

			return next(c)
		}
	}
}

// LoginRateLimitMiddleware specifically for login endpoint
func LoginRateLimitMiddleware(queries *db.Queries, maxAttempts int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientIP := c.RealIP()
			ctx := c.Request().Context()
			windowStart := time.Now().Add(-window)

			// Check login attempts in the time window
			count, err := queries.CountLoginAttempts(ctx, db.CountLoginAttemptsParams{
				ClientID:    clientIP,
				WindowStart: windowStart,
			})

			if err != nil {
				c.Logger().Error("Failed to check login attempts:", err)
				// Allow request on error to prevent DOS via DB errors
			} else if count >= int64(maxAttempts) {
				RecordRateLimitExceeded("/api/v1/auth/login")
				return echo.NewHTTPError(http.StatusTooManyRequests, 
					"Too many login attempts. Please try again in 5 minutes.")
			}

			// Continue with request
			err = next(c)

			// Record login attempt after processing
			// Use goroutine to avoid blocking
			go func() {
				asyncCtx := context.Background()
				_, recordErr := queries.RecordLoginAttempt(asyncCtx, clientIP)
				if recordErr != nil {
					// Log but don't fail the request
					c.Logger().Error("Failed to record login attempt:", recordErr)
				}
			}()

			return err
		}
	}
}