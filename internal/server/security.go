// internal/server/security.go - Security Monitoring Handlers
package server

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/jamalkaksouri/DigiOrder/internal/middleware"
)

// GetLoginAttempts - Admin endpoint to view login attempts
func (s *Server) GetLoginAttempts(c echo.Context) error {
	ctx := c.Request().Context()

	limit := 50
	offset := 0

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get rate limited attempts
	attempts, err := s.queries.GetRateLimitedAttempts(ctx, db.GetRateLimitedAttemptsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})

	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve login attempts.")
	}

	if attempts == nil {
		attempts = []db.LoginAttemptsLog{}
	}

	return RespondSuccess(c, http.StatusOK, attempts)
}

// GetLoginSecurityReport - Get security report of suspicious IPs
func (s *Server) GetLoginSecurityReport(c echo.Context) error {
	ctx := c.Request().Context()

	limit := 50
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	report, err := s.queries.GetLoginSecurityReport(ctx, int32(limit))
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve security report.")
	}

	if report == nil {
		report = []db.GetLoginSecurityReportRow{}
	}

	return RespondSuccess(c, http.StatusOK, report)
}

// GetCurrentlyBlockedIPs - View currently rate-limited IPs
func (s *Server) GetCurrentlyBlockedIPs(c echo.Context) error {
	ctx := c.Request().Context()

	blockedIPs, err := s.queries.GetCurrentlyBlockedIPs(ctx)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve blocked IPs.")
	}

	if blockedIPs == nil {
		blockedIPs = []db.CurrentlyBlockedIp{}
	}

	return RespondSuccess(c, http.StatusOK, blockedIPs)
}

// ManuallyReleaseIP - Admin manually releases an IP from rate limiting
func (s *Server) ManuallyReleaseIP(c echo.Context) error {
	var req struct {
		IPAddress string `json:"ip_address" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()

	// Delete rate limit records for this IP
	err := s.queries.ManuallyReleaseRateLimit(ctx, req.IPAddress)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to release IP from rate limit.")
	}

	// Update login attempts to mark as released
	err = s.queries.UpdateLoginAttemptRelease(ctx, db.UpdateLoginAttemptReleaseParams{
		ReleasedBy: sql.NullString{String: "admin_manual", Valid: true},
		IpAddress:  req.IPAddress,
	})
	if err != nil {
		// Log but don't fail
		if s.logger != nil {
			s.logger.Error("Failed to update login attempt release", err, nil)
		}
	}

	// Log rate limit release
	adminID, _ := middleware.GetUserIDFromContext(c)
	_, err = s.queries.LogRateLimitRelease(ctx, db.LogRateLimitReleaseParams{
		ClientID:         req.IPAddress,
		IpAddress:        req.IPAddress,
		Username:         sql.NullString{},
		BlockedAt:        time.Now().Add(-5 * time.Minute), // Approximate
		ReleasedBy:       "admin_manual",
		ReleasedByUserID: uuid.NullUUID{UUID: adminID, Valid: true},
		BlockDuration:    sql.NullInt64{},
		AttemptsCount:    sql.NullInt32{},
		ReleaseReason:    sql.NullString{String: "manually_released_by_admin", Valid: true},
	})

	if err != nil {
		// Log but don't fail
		if s.logger != nil {
			s.logger.Error("Failed to log rate limit release", err, nil)
		}
	}

	return RespondSuccess(c, http.StatusOK, map[string]string{
		"message": "IP successfully released from rate limit",
		"ip":      req.IPAddress,
	})
}

// CleanupOldData - Admin endpoint to cleanup old security logs
func (s *Server) CleanupOldData(c echo.Context) error {
	ctx := c.Request().Context()

	// Cleanup old login attempts
	err := s.queries.CleanupOldLoginAttempts(ctx)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to cleanup old login attempts.")
	}

	// Archive old rate limits
	err = s.queries.ArchiveOldRateLimits(ctx)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to archive old rate limits.")
	}

	return RespondSuccess(c, http.StatusOK, map[string]string{
		"message": "Old security data cleaned up successfully",
	})
}