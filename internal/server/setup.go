// internal/server/setup.go - FIXED VERSION
package server

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/google/uuid"
	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/jamalkaksouri/DigiOrder/internal/security"
	"github.com/labstack/echo/v4"
)

// InitialSetupRequest defines the secure setup request
type InitialSetupRequest struct {
	Username        string `json:"username" validate:"required,min=3,max=50"`
	Password        string `json:"password" validate:"required,min=12"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
	FullName        string `json:"full_name" validate:"required"`
	SetupToken      string `json:"setup_token" validate:"required"`
}

// InitialSetup handles POST /api/v1/setup/initialize
// This is a ONE-TIME endpoint that creates the first admin user
func (s *Server) InitialSetup(c echo.Context) error {
	ctx := c.Request().Context()

	// Check if setup is already complete
	setupStatus, err := s.queries.GetSystemSetupStatus(ctx)
	if err != nil && err != sql.ErrNoRows {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to check setup status.")
	}

	if setupStatus.AdminCreated.Valid && setupStatus.AdminCreated.Bool {
		return RespondError(c, http.StatusForbidden, "already_setup",
			"System has already been initialized. This endpoint is disabled.")
	}

	// Parse request
	var req InitialSetupRequest
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	// Verify passwords match
	if req.Password != req.ConfirmPassword {
		return RespondError(c, http.StatusBadRequest, "password_mismatch",
			"Passwords do not match.")
	}

	// Verify setup token (should be provided via secure channel)
	expectedToken := getEnv("INITIAL_SETUP_TOKEN", "")
	if expectedToken == "" || req.SetupToken != expectedToken {
		return RespondError(c, http.StatusUnauthorized, "invalid_setup_token",
			"Invalid or missing setup token.")
	}

	// Validate password strength
	if err := security.ValidatePassword(req.Password,
		security.DefaultPasswordRequirements()); err != nil {
		suggestions := security.SuggestPasswordImprovement(req.Password)
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error":       "weak_password",
			"details":     err.Error(),
			"suggestions": suggestions,
		})
	}

	// Hash password
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "hash_error",
			"Failed to process password.")
	}

	// Create admin user with fixed UUID for protection
	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	user, err := s.queries.CreateAdminUser(ctx, db.CreateAdminUserParams{
		ID:           adminID,
		Username:     req.Username,
		FullName:     sql.NullString{String: req.FullName, Valid: true},
		PasswordHash: hashedPassword,
		RoleID:       sql.NullInt32{Int32: 1, Valid: true}, // Admin role
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to create admin user.")
	}

	// Mark setup as complete
	_, err = s.queries.CompleteSystemSetup(ctx, db.CompleteSystemSetupParams{
		AdminCreated: sql.NullBool{Bool: true, Valid: true},
		SetupByIp:    sql.NullString{String: c.RealIP(), Valid: true},
	})
	if err != nil {
		// Log error but don't fail the request since admin was created
		// FIXED: Only use logger if it's not nil
		if s.logger != nil {
			s.logger.Error("Failed to mark setup as complete", err, nil)
		} else {
			s.router.Logger.Error("Failed to mark setup as complete:", err)
		}
	}

	// Don't return password hash
	user.PasswordHash = ""

	return RespondSuccess(c, http.StatusCreated, map[string]any{
		"message": "System initialized successfully. Initial admin user created.",
		"user":    user,
		"next_steps": []string{
			"1. Login with your credentials",
			"2. Create additional users as needed",
			"3. Configure system settings",
			"4. Remove INITIAL_SETUP_TOKEN from environment",
		},
	})
}

// GetSetupStatus handles GET /api/v1/setup/status
func (s *Server) GetSetupStatus(c echo.Context) error {
	ctx := c.Request().Context()

	setupStatus, err := s.queries.GetSystemSetupStatus(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondSuccess(c, http.StatusOK, map[string]any{
				"setup_required": true,
				"admin_exists":   false,
			})
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to check setup status.")
	}

	return RespondSuccess(c, http.StatusOK, map[string]any{
		"setup_required":     !(setupStatus.AdminCreated.Valid && setupStatus.AdminCreated.Bool),
		"admin_exists":       setupStatus.AdminCreated.Valid && setupStatus.AdminCreated.Bool,
		"setup_completed_at": setupStatus.SetupCompletedAt,
	})
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
