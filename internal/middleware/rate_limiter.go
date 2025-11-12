package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for clients
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// GetLimiter returns the rate limiter for a given IP
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
	}

	return limiter
}

// CleanupVisitors removes old visitors periodically
func (rl *RateLimiter) CleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, limiter := range rl.visitors {
			if limiter.Tokens() == float64(rl.burst) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(requestsPerSecond int, burst int) echo.MiddlewareFunc {
	limiter := NewRateLimiter(rate.Limit(requestsPerSecond), burst)
	
	// Start cleanup goroutine
	go limiter.CleanupVisitors()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get client IP
			ip := c.RealIP()
			
			// Get limiter for this IP
			l := limiter.GetLimiter(ip)
			
			if !l.Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}

			return next(c)
		}
	}
}

// APIKeyRateLimiter manages rate limiting based on API keys
type APIKeyRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewAPIKeyRateLimiter creates a new API key-based rate limiter
func NewAPIKeyRateLimiter(r rate.Limit, b int) *APIKeyRateLimiter {
	return &APIKeyRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// GetLimiter returns the rate limiter for a given API key
func (rl *APIKeyRateLimiter) GetLimiter(apiKey string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[apiKey]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[apiKey] = limiter
	}

	return limiter
}

// APIKeyRateLimitMiddleware creates an API key-based rate limiting middleware
func APIKeyRateLimitMiddleware(requestsPerMinute int) echo.MiddlewareFunc {
	limiter := NewAPIKeyRateLimiter(rate.Limit(float64(requestsPerMinute)/60.0), requestsPerMinute)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get API key from header or use IP as fallback
			apiKey := c.Request().Header.Get("X-API-Key")
			if apiKey == "" {
				apiKey = c.RealIP()
			}

			l := limiter.GetLimiter(apiKey)
			
			if !l.Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "API rate limit exceeded")
			}

			return next(c)
		}
	}
}