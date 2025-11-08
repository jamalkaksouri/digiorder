package server

import (
	"database/sql"
	"net/http"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/jamalkaksouri/DigiOrder/internal/middleware"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest defines the login request body
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse defines the login response
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn string `json:"expires_in"`
	User      UserInfo `json:"user"`
}

// UserInfo contains basic user information
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	RoleID   int32  `json:"role_id"`
	RoleName string `json:"role_name"`
}

// RefreshTokenRequest defines the refresh token request
type RefreshTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// Login handles POST /api/v1/auth/login
func (s *Server) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()

	// Get user by username
	user, err := s.queries.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to authenticate user.")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return RespondError(c, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password.")
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
		return RespondError(c, http.StatusInternalServerError, "token_error", "Failed to generate authentication token.")
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

// RefreshToken handles POST /api/v1/auth/refresh
func (s *Server) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	// Validate existing token
	claims, err := middleware.ValidateToken(req.Token)
	if err != nil {
		return RespondError(c, http.StatusUnauthorized, "invalid_token", "Invalid or expired token.")
	}

	// Generate new token with same claims
	newToken, err := middleware.GenerateToken(claims.UserID, claims.Username, claims.RoleID, claims.RoleName)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "token_error", "Failed to refresh token.")
	}

	response := map[string]interface{}{
		"token":      newToken,
		"expires_in": middleware.GetJWTExpiry().String(),
	}

	return RespondSuccess(c, http.StatusOK, response)
}

// GetProfile handles GET /api/v1/auth/profile
func (s *Server) GetProfile(c echo.Context) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	user, err := s.queries.GetUser(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", "User not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve user profile.")
	}

	// Get role name
	var roleName string
	if user.RoleID.Valid {
		role, err := s.queries.GetRole(ctx, user.RoleID.Int32)
		if err == nil {
			roleName = role.Name
		}
	}

	userInfo := UserInfo{
		ID:       user.ID.String(),
		Username: user.Username,
		FullName: user.FullName.String,
		RoleID:   user.RoleID.Int32,
		RoleName: roleName,
	}

	return RespondSuccess(c, http.StatusOK, userInfo)
}

// ChangePassword handles PUT /api/v1/auth/password
func (s *Server) ChangePassword(c echo.Context) error {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=6"`
	}

	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()

	// Get current user
	user, err := s.queries.GetUser(ctx, userID)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve user.")
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword))
	if err != nil {
		return RespondError(c, http.StatusUnauthorized, "invalid_password", "Current password is incorrect.")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "hash_error", "Failed to hash password.")
	}

	// Update password
	err = s.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to update password.")
	}

	return RespondSuccess(c, http.StatusOK, map[string]string{
		"message": "Password updated successfully",
	})
}