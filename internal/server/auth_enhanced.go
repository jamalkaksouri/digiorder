package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/jamalkaksouri/DigiOrder/internal/middleware"
	"github.com/labstack/echo/v4"
	"github.com/sqlc-dev/pqtype"
	"golang.org/x/crypto/bcrypt"
)

// parseUserAgent extracts device information from user agent string
func parseUserAgent(userAgent string) map[string]interface{} {
	deviceInfo := map[string]interface{}{
		"raw": userAgent,
	}

	if len(userAgent) > 0 {
		deviceInfo["is_mobile"] = false
		deviceInfo["is_desktop"] = true
		deviceInfo["browser"] = "unknown"
	}

	return deviceInfo
}

// logLoginAttempt logs all login attempts with comprehensive details
func (s *Server) logLoginAttempt(c echo.Context, username string, success bool,
	failureReason string, rateLimited bool) error {

	ctx := c.Request().Context()
	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()
	sessionID := c.Request().Header.Get("X-Session-ID")

	deviceInfo := parseUserAgent(userAgent)
	deviceInfoJSON, _ := json.Marshal(deviceInfo)

	_, err := s.queries.LogLoginAttempt(ctx, db.LogLoginAttemptParams{
		Username:      username,
		IpAddress:     ipAddress,
		UserAgent:     sql.NullString{String: userAgent, Valid: userAgent != ""},
		Success:       success,
		FailureReason: sql.NullString{String: failureReason, Valid: failureReason != ""},
		RateLimited:   sql.NullBool{Bool: rateLimited, Valid: true},
		SessionID:     sql.NullString{String: sessionID, Valid: sessionID != ""},
		DeviceInfo:    pqtype.NullRawMessage{RawMessage: deviceInfoJSON, Valid: true},
	})

	return err
}

// checkAndLogRateLimit checks if IP is rate limited and logs accordingly
// FIXED: Changed return type and parameter types
func (s *Server) checkAndLogRateLimit(c echo.Context, username string) (bool, error) {
	ctx := c.Request().Context()
	clientIP := c.RealIP()
	windowStart := time.Now().Add(-5 * time.Minute) // Changed from sql.NullTime

	// FIXED: Use time.Time instead of sql.NullTime
	count, err := s.queries.CountFailedAttempts(ctx, db.CountFailedAttemptsParams{
		IpAddress: clientIP,
		Since:     sql.NullTime{Time: windowStart, Valid: true},
	})

	if err != nil {
		return false, err
	}

	isRateLimited := count >= 5

	if isRateLimited {
		s.logLoginAttempt(c, username, false, "rate_limited", true)
	}

	return isRateLimited, nil
}

// logRateLimitRelease logs when a user is released from rate limiting
// FIXED: Changed BlockDuration type to sql.NullInt64
func (s *Server) logRateLimitRelease(c echo.Context, clientID, ipAddress, username string,
	blockedAt time.Time, releasedBy string, adminUserID *uuid.UUID,
	attemptsCount int, releaseReason string) error {

	ctx := c.Request().Context()
	blockDuration := time.Since(blockedAt)

	var adminID uuid.NullUUID
	if adminUserID != nil {
		adminID = uuid.NullUUID{UUID: *adminUserID, Valid: true}
	}

	// FIXED: Convert duration to int64 (seconds)
	_, err := s.queries.LogRateLimitRelease(ctx, db.LogRateLimitReleaseParams{
		ClientID:         clientID,
		IpAddress:        ipAddress,
		Username:         sql.NullString{String: username, Valid: username != ""},
		BlockedAt:        blockedAt,
		ReleasedBy:       releasedBy,
		ReleasedByUserID: adminID,
		BlockDuration:    sql.NullInt64{Int64: int64(blockDuration.Seconds()), Valid: true}, // FIXED
		AttemptsCount:    sql.NullInt32{Int32: int32(attemptsCount), Valid: true},
		ReleaseReason:    sql.NullString{String: releaseReason, Valid: releaseReason != ""},
	})

	if err == nil {
		s.queries.UpdateLoginAttemptRelease(ctx, db.UpdateLoginAttemptReleaseParams{
			IpAddress:  ipAddress,
			ReleasedBy: sql.NullString{String: releasedBy, Valid: true},
		})
	}

	return err
}

// LoginEnhanced - Enhanced login handler with comprehensive logging
func (s *Server) LoginEnhanced(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		s.logLoginAttempt(c, req.Username, false, "invalid_request", false)
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		s.logLoginAttempt(c, req.Username, false, "validation_error", false)
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()

	// Check rate limiting BEFORE attempting authentication
	isRateLimited, err := s.checkAndLogRateLimit(c, req.Username)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to check rate limit", err, nil)
		}
	}

	if isRateLimited {
		return RespondError(c, http.StatusTooManyRequests, "rate_limited",
			"Too many login attempts. Please try again in 5 minutes.")
	}

	// Get user by username
	user, err := s.queries.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logLoginAttempt(c, req.Username, false, "invalid_username", false)
			return RespondError(c, http.StatusUnauthorized, "invalid_credentials",
				"Invalid username or password.")
		}
		s.logLoginAttempt(c, req.Username, false, "database_error", false)
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to authenticate user.")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		s.logLoginAttempt(c, req.Username, false, "invalid_password", false)
		return RespondError(c, http.StatusUnauthorized, "invalid_credentials",
			"Invalid username or password.")
	}

	// Get role name
	var roleName string
	if user.RoleID.Valid {
		role, err := s.queries.GetRole(ctx, user.RoleID.Int32)
		if err == nil {
			roleName = role.Name
		}
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID, user.Username, user.RoleID.Int32, roleName)
	if err != nil {
		s.logLoginAttempt(c, req.Username, false, "token_generation_error", false)
		return RespondError(c, http.StatusInternalServerError, "token_error",
			"Failed to generate authentication token.")
	}

	// Log successful login
	s.logLoginAttempt(c, req.Username, true, "", false)

	// FIXED: Use CountFailedAttempts instead of CountLoginAttempts
	clientIP := c.RealIP()
	windowStart := time.Now().Add(-5 * time.Minute)
	count, _ := s.queries.CountFailedAttempts(ctx, db.CountFailedAttemptsParams{
		IpAddress: clientIP,
		Since:     sql.NullTime{Time: windowStart, Valid: true},
	})

	if count > 0 {
		s.logRateLimitRelease(c, clientIP, clientIP, req.Username,
			time.Now().Add(-5*time.Minute), "automatic_expiry", nil,
			int(count), "successful_authentication")
	}

	// Prepare response
	response := LoginResponse{
		Token:     token,
		ExpiresIn: middleware.GetJWTExpiry().String(),
		User: UserInfo{
			ID:       user.ID.String(),
			Username: user.Username,
			FullName: user.FullName.String,
			RoleID:   user.RoleID.Int32,
			RoleName: roleName,
		},
	}

	return RespondSuccess(c, http.StatusOK, response)
}

// GetLoginAttempts - Admin endpoint to view login attempts
func (s *Server) GetLoginAttempts(c echo.Context) error {
	ctx := c.Request().Context()

	ipAddress := c.QueryParam("ip_address")
	username := c.QueryParam("username")
	limit := 50
	offset := 0

	var attempts []db.LoginAttemptsLog
	var err error

	if ipAddress != "" {
		attempts, err = s.queries.GetRecentLoginAttempts(ctx, db.GetRecentLoginAttemptsParams{
			IpAddress: ipAddress,
			Since:     sql.NullTime{Time: time.Now().Add(-24 * time.Hour), Valid: true},
			Limit:     int32(limit),
		})
	} else if username != "" {
		attempts, err = s.queries.GetLoginAttemptsByUsername(ctx, db.GetLoginAttemptsByUsernameParams{
			Username: username,
			Since:    sql.NullTime{Time: time.Now().Add(-24 * time.Hour), Valid: true},
			Limit:    int32(limit),
		})
	} else {
		attempts, err = s.queries.GetRateLimitedAttempts(ctx, db.GetRateLimitedAttemptsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
	}

	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve login attempts.")
	}

	if attempts == nil {
		attempts = []db.LoginAttemptsLog{}
	}

	return RespondSuccess(c, http.StatusOK, attempts)
}

// GetLoginSecurityReport - Admin endpoint for security analysis
func (s *Server) GetLoginSecurityReport(c echo.Context) error {
	ctx := c.Request().Context()

	report, err := s.queries.GetLoginSecurityReport(ctx, 20)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to generate security report.")
	}

	if report == nil {
		report = []db.GetLoginSecurityReportRow{}
	}

	return RespondSuccess(c, http.StatusOK, report)
}

// GetCurrentlyBlockedIPs - Admin endpoint to see blocked IPs
func (s *Server) GetCurrentlyBlockedIPs(c echo.Context) error {
	ctx := c.Request().Context()

	blocked, err := s.queries.GetCurrentlyBlockedIPs(ctx)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve blocked IPs.")
	}

	if blocked == nil {
		blocked = []db.CurrentlyBlockedIp{}
	}

	return RespondSuccess(c, http.StatusOK, blocked)
}

// ManuallyReleaseIP - Admin endpoint to manually release a blocked IP
func (s *Server) ManuallyReleaseIP(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		IPAddress     string `json:"ip_address" validate:"required"`
		ReleaseReason string `json:"release_reason" validate:"required"`
	}

	if err := s.ValidateRequest(c, &req); err != nil {
		return err
	}

	adminUserID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	err = s.queries.ManuallyReleaseRateLimit(ctx, req.IPAddress)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to release IP from rate limiting.")
	}

	s.logRateLimitRelease(c, req.IPAddress, req.IPAddress, "",
		time.Now().Add(-5*time.Minute), "manual_admin", &adminUserID,
		0, req.ReleaseReason)

	return RespondSuccess(c, http.StatusOK, map[string]string{
		"message": "IP address has been released from rate limiting",
		"ip":      req.IPAddress,
	})
}

// CleanupOldData - Admin endpoint to trigger cleanup
func (s *Server) CleanupOldData(c echo.Context) error {
	ctx := c.Request().Context()

	err := s.queries.ArchiveOldRateLimits(ctx)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to archive rate limits.")
	}

	err = s.queries.CleanupOldLoginAttempts(ctx)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to cleanup login attempts.")
	}

	return RespondSuccess(c, http.StatusOK, map[string]string{
		"message": "Cleanup completed successfully",
		"actions": "Archived old rate limits (7+ days) and deleted old login attempts (90+ days)",
	})
}
