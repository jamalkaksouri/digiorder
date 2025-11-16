// internal/middleware/rate_limiter_production.go - ENHANCED VERSION
package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"time"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
	"github.com/sqlc-dev/pqtype"
	"golang.org/x/time/rate"
)

// BannedIP represents a temporarily banned IP with expiry
type BannedIP struct {
	IP          string
	BannedUntil time.Time
	Reason      string
	Attempts    int
	mu          sync.RWMutex
}

// IPBanManager manages temporarily banned IPs
type IPBanManager struct {
	bans    map[string]*BannedIP
	mu      sync.RWMutex
	queries *db.Queries
	ticker  *time.Ticker
}

// NewIPBanManager creates a new IP ban manager with auto-cleanup
func NewIPBanManager(queries *db.Queries) *IPBanManager {
	manager := &IPBanManager{
		bans:    make(map[string]*BannedIP),
		queries: queries,
		ticker:  time.NewTicker(30 * time.Second), // Check every 30 seconds
	}

	// Start cleanup goroutine
	go manager.cleanupExpiredBans()

	return manager
}

// IsBanned checks if an IP is currently banned
func (m *IPBanManager) IsBanned(ip string) (bool, time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ban, exists := m.bans[ip]
	if !exists {
		return false, 0
	}

	ban.mu.RLock()
	defer ban.mu.RUnlock()

	if time.Now().After(ban.BannedUntil) {
		return false, 0
	}

	remaining := time.Until(ban.BannedUntil)
	return true, remaining
}

// BanIP temporarily bans an IP address
func (m *IPBanManager) BanIP(ip, reason string, duration time.Duration, attempts int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bannedUntil := time.Now().Add(duration)

	m.bans[ip] = &BannedIP{
		IP:          ip,
		BannedUntil: bannedUntil,
		Reason:      reason,
		Attempts:    attempts,
	}

	// Log to database for persistence
	if m.queries != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := m.queries.LogLoginAttempt(ctx, db.LogLoginAttemptParams{
				Username:      "system",
				IpAddress:     ip,
				UserAgent:     sql.NullString{String: "rate_limiter", Valid: true},
				Success:       false,
				FailureReason: sql.NullString{String: reason, Valid: true},
				RateLimited:   sql.NullBool{Bool: true, Valid: true},
				SessionID:     sql.NullString{String: "ban_" + time.Now().Format("20060102150405"), Valid: true},
				DeviceInfo: pqtype.NullRawMessage{
					Valid: true,
				},
			})
			if err != nil {
				// Log error but don't fail
				return
			}
		}()
	}
}

// UnbanIP manually removes a ban
func (m *IPBanManager) UnbanIP(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.bans, ip)
}

// GetBannedIPs returns all currently banned IPs
func (m *IPBanManager) GetBannedIPs() []BannedIP {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]BannedIP, 0, len(m.bans))
	now := time.Now()

	for _, ban := range m.bans {
		ban.mu.RLock()
		if now.Before(ban.BannedUntil) {
			result = append(result, BannedIP{
				IP:          ban.IP,
				BannedUntil: ban.BannedUntil,
				Reason:      ban.Reason,
				Attempts:    ban.Attempts,
			})
		}
		ban.mu.RUnlock()
	}

	return result
}

// cleanupExpiredBans removes expired bans automatically
func (m *IPBanManager) cleanupExpiredBans() {
	for range m.ticker.C {
		m.mu.Lock()
		now := time.Now()

		for ip, ban := range m.bans {
			ban.mu.RLock()
			if now.After(ban.BannedUntil) {
				delete(m.bans, ip)
			}
			ban.mu.RUnlock()
		}

		m.mu.Unlock()
	}
}

// Stop stops the cleanup goroutine
func (m *IPBanManager) Stop() {
	m.ticker.Stop()
}

// EnhancedRateLimiter with IP ban tracking
type EnhancedRateLimiter struct {
	limiters    map[string]*rate.Limiter
	mu          sync.RWMutex
	globalRate  rate.Limit
	globalBurst int
	queries     *db.Queries
	banManager  *IPBanManager
}

// NewEnhancedRateLimiter creates a production-ready rate limiter
func NewEnhancedRateLimiter(queries *db.Queries, globalRPS, burst int) *EnhancedRateLimiter {
	return &EnhancedRateLimiter{
		limiters:    make(map[string]*rate.Limiter),
		globalRate:  rate.Limit(globalRPS),
		globalBurst: burst,
		queries:     queries,
		banManager:  NewIPBanManager(queries),
	}
}

// GetLimiter returns or creates a limiter for an IP
func (rl *EnhancedRateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.globalRate, rl.globalBurst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// CheckRateLimit checks both in-memory and database rate limits
func (rl *EnhancedRateLimiter) CheckRateLimit(c echo.Context, endpoint string) error {
	clientIP := c.RealIP()

	// Check if IP is banned first
	if banned, remaining := rl.banManager.IsBanned(clientIP); banned {
		RecordRateLimitExceeded(endpoint)

		minutes := int(remaining.Minutes())
		seconds := int(remaining.Seconds()) % 60

		return echo.NewHTTPError(http.StatusTooManyRequests, map[string]any{
			"error":   "ip_temporarily_banned",
			"message": "Your IP has been temporarily banned due to too many failed requests",
			"details": map[string]any{
				"banned_until": time.Now().Add(remaining).Format(time.RFC3339),
				"time_remaining": map[string]int{
					"minutes": minutes,
					"seconds": seconds,
				},
				"reason": "excessive_failed_login_attempts",
			},
		})
	}

	// Check in-memory rate limit
	limiter := rl.GetLimiter(clientIP)
	if !limiter.Allow() {
		RecordRateLimitExceeded(endpoint)

		// Check if this is a login endpoint - stricter enforcement
		if endpoint == "/api/v1/auth/login" {
			// Check failed attempts in last 5 minutes
			ctx := c.Request().Context()
			windowStart := time.Now().Add(-5 * time.Minute)

			count, err := rl.queries.CountFailedAttempts(ctx, db.CountFailedAttemptsParams{
				IpAddress: clientIP,
				Since:     sql.NullTime{Time: windowStart, Valid: true},
			})

			if err == nil && count >= 5 {
				// Ban for 5 minutes
				rl.banManager.BanIP(clientIP, "too_many_failed_logins", 5*time.Minute, int(count))

				return echo.NewHTTPError(http.StatusTooManyRequests, map[string]any{
					"error":   "ip_banned",
					"message": "Too many failed login attempts. Your IP has been banned for 5 minutes.",
					"details": map[string]any{
						"failed_attempts": count,
						"ban_duration":    "5 minutes",
						"retry_after":     time.Now().Add(5 * time.Minute).Format(time.RFC3339),
					},
				})
			}
		}

		return echo.NewHTTPError(http.StatusTooManyRequests,
			"Rate limit exceeded. Please slow down your requests.")
	}

	return nil
}

// ProductionRateLimitMiddleware - Enhanced rate limiting with bans
func ProductionRateLimitMiddleware(queries *db.Queries) echo.MiddlewareFunc {
	limiter := NewEnhancedRateLimiter(queries, 100, 200) // 100 req/sec, burst 200

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			endpoint := c.Path()

			// Skip rate limiting for health/metrics
			if shouldSkipRateLimit(endpoint) {
				return next(c)
			}

			// Check rate limit
			if err := limiter.CheckRateLimit(c, endpoint); err != nil {
				return err
			}

			return next(c)
		}
	}
}

// shouldSkipRateLimit determines if endpoint should bypass rate limiting
func shouldSkipRateLimit(endpoint string) bool {
	skipEndpoints := []string{
		"/health",
		"/metrics",
		"/api/health",
		"/api/metrics",
	}

	for _, skip := range skipEndpoints {
		if endpoint == skip {
			return true
		}
	}

	return false
}

// GetBannedIPsHandler - Handler to view currently banned IPs (admin only)
func GetBannedIPsHandler(limiter *EnhancedRateLimiter) echo.HandlerFunc {
	return func(c echo.Context) error {
		banned := limiter.banManager.GetBannedIPs()

		result := make([]map[string]any, len(banned))
		now := time.Now()

		for i, ban := range banned {
			remaining := ban.BannedUntil.Sub(now)
			result[i] = map[string]any{
				"ip":              ban.IP,
				"banned_until":    ban.BannedUntil.Format(time.RFC3339),
				"reason":          ban.Reason,
				"failed_attempts": ban.Attempts,
				"time_remaining": map[string]int{
					"minutes": int(remaining.Minutes()),
					"seconds": int(remaining.Seconds()) % 60,
				},
			}
		}

		return c.JSON(http.StatusOK, map[string]any{
			"data":  result,
			"count": len(result),
		})
	}
}

// UnbanIPHandler - Handler to manually unban an IP (admin only)
func UnbanIPHandler(limiter *EnhancedRateLimiter) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			IPAddress string `json:"ip_address" validate:"required"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
		}

		limiter.banManager.UnbanIP(req.IPAddress)

		return c.JSON(http.StatusOK, map[string]string{
			"message": "IP successfully unbanned",
			"ip":      req.IPAddress,
		})
	}
}
