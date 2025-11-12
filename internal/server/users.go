// internal/server/users.go - Enhanced version with admin protection

package server

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/jamalkaksouri/DigiOrder/internal/middleware"
	"github.com/jamalkaksouri/DigiOrder/internal/security"
	"github.com/labstack/echo/v4"
)

// Constants
const (
	RoleAdmin      = 1                                      // Admin role ID
	PrimaryAdminID = "00000000-0000-0000-0000-000000000001" // Primary admin UUID
)

// CreateUserReq defines the request body for creating a user
type CreateUserReq struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	FullName string `json:"full_name,omitempty"`
	Password string `json:"password" validate:"required,min=8"`
	RoleID   int32  `json:"role_id" validate:"required,gt=0"`
}

// UpdateUserReq defines the request body for updating a user
type UpdateUserReq struct {
	FullName string `json:"full_name,omitempty"`
	RoleID   *int32 `json:"role_id,omitempty"`
}

// CreateUser handles POST /api/v1/users (Admin only)
func (s *Server) CreateUser(c echo.Context) error {
	// Verify admin role
	roleName, err := middleware.GetRoleNameFromContext(c)
	if err != nil || roleName != "admin" {
		return RespondError(c, http.StatusForbidden, "insufficient_permissions",
			"Only administrators can create users.")
	}

	var req CreateUserReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	// Use new password validation
	if err := security.ValidatePassword(req.Password,
		security.DefaultPasswordRequirements()); err != nil {

		suggestions := security.SuggestPasswordImprovement(req.Password)
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error":       "weak_password",
			"details":     err.Error(),
			"suggestions": suggestions,
			"requirements": map[string]any{
				"min_length": 12,
				"requires":   []string{"uppercase", "lowercase", "digit", "special char"},
			},
		})
	}

	// Hash with stronger cost
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError,
			"hash_error", "Failed to hash password.")
	}

	// Verify role exists
	ctx := c.Request().Context()
	role, err := s.queries.GetRole(ctx, req.RoleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusBadRequest, "invalid_role",
				"The specified role does not exist.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to verify role.")
	}

	// Create user
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Username:     req.Username,
		FullName:     sql.NullString{String: req.FullName, Valid: req.FullName != ""},
		PasswordHash: string(hashedPassword),
		RoleID:       sql.NullInt32{Int32: req.RoleID, Valid: true},
	})
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_username_key\"" {
			return RespondError(c, http.StatusConflict, "duplicate_username",
				"A user with this username already exists.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to create user.")
	}

	// Log audit
	currentUserID, _ := middleware.GetUserIDFromContext(c)
	s.logAudit(ctx, currentUserID, "create", "user", user.ID.String(), nil, map[string]any{
		"username": user.Username,
		"role":     role.Name,
	}, c.RealIP(), c.Request().UserAgent())

	// Don't return password hash
	user.PasswordHash = ""

	return RespondSuccess(c, http.StatusCreated, user)
}

// GetUser handles GET /api/v1/users/:id
func (s *Server) GetUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id",
			"The provided ID is not a valid UUID.")
	}

	ctx := c.Request().Context()
	user, err := s.queries.GetUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found",
				"User with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve user.")
	}

	// Check if user is deleted
	if user.DeletedAt.Valid {
		return RespondError(c, http.StatusNotFound, "not_found",
			"User has been deleted.")
	}

	// Get role name
	var roleName string
	if user.RoleID.Valid {
		role, err := s.queries.GetRole(ctx, user.RoleID.Int32)
		if err == nil {
			roleName = role.Name
		}
	}

	// Don't return password hash
	// user.PasswordHash = ""

	response := map[string]any{
		"id":         user.ID,
		"username":   user.Username,
		"full_name":  user.FullName.String,
		"role_id":    user.RoleID.Int32,
		"role_name":  roleName,
		"created_at": user.CreatedAt,
	}

	return RespondSuccess(c, http.StatusOK, response)
}

// ListUsers handles GET /api/v1/users
func (s *Server) ListUsers(c echo.Context) error {
	ctx := c.Request().Context()

	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	users, err := s.queries.ListActiveUsers(ctx, db.ListActiveUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to fetch users.")
	}

	if users == nil {
		users = []db.User{}
	}

	// Remove password hashes and add role names
	result := make([]map[string]any, len(users))
	for i, user := range users {
		var roleName string
		if user.RoleID.Valid {
			role, err := s.queries.GetRole(ctx, user.RoleID.Int32)
			if err == nil {
				roleName = role.Name
			}
		}

		result[i] = map[string]any{
			"id":         user.ID,
			"username":   user.Username,
			"full_name":  user.FullName.String,
			"role_id":    user.RoleID.Int32,
			"role_name":  roleName,
			"created_at": user.CreatedAt,
		}
	}

	return RespondSuccess(c, http.StatusOK, result)
}

// UpdateUser handles PUT /api/v1/users/:id
func (s *Server) UpdateUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id",
			"The provided ID is not a valid UUID.")
	}

	// Check if trying to update primary admin
	if id.String() == PrimaryAdminID {
		return RespondError(c, http.StatusForbidden, "protected_user",
			"The primary administrator account cannot be modified.")
	}

	var req UpdateUserReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	ctx := c.Request().Context()

	// Get old values for audit
	oldUser, err := s.queries.GetUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found",
				"User with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve user.")
	}

	params := db.UpdateUserParams{
		ID: id,
	}

	if req.FullName != "" {
		params.FullName = sql.NullString{String: req.FullName, Valid: true}
	}
	if req.RoleID != nil {
		// Verify new role exists
		_, err := s.queries.GetRole(ctx, *req.RoleID)
		if err != nil {
			if err == sql.ErrNoRows {
				return RespondError(c, http.StatusBadRequest, "invalid_role",
					"The specified role does not exist.")
			}
			return RespondError(c, http.StatusInternalServerError, "db_error",
				"Failed to verify role.")
		}
		params.RoleID = sql.NullInt32{Int32: *req.RoleID, Valid: true}
	}

	user, err := s.queries.UpdateUser(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found",
				"User with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to update user.")
	}

	// Log audit
	currentUserID, _ := middleware.GetUserIDFromContext(c)
	s.logAudit(ctx, currentUserID, "update", "user", user.ID.String(),
		map[string]any{
			"full_name": oldUser.FullName.String,
			"role_id":   oldUser.RoleID.Int32,
		},
		map[string]any{
			"full_name": user.FullName.String,
			"role_id":   user.RoleID.Int32,
		},
		c.RealIP(), c.Request().UserAgent())

	// Don't return password hash
	user.PasswordHash = ""

	return RespondSuccess(c, http.StatusOK, user)
}

// DeleteUser handles DELETE /api/v1/users/:id (Soft delete)
func (s *Server) DeleteUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id",
			"The provided ID is not a valid UUID.")
	}

	// CRITICAL: Protect primary admin from deletion
	if id.String() == PrimaryAdminID {
		return RespondError(c, http.StatusForbidden, "protected_user",
			"The primary administrator account cannot be deleted. This account is essential for system administration.")
	}

	ctx := c.Request().Context()

	// Get user for audit
	user, err := s.queries.GetUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found",
				"User with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve user.")
	}

	// Check if user is the last admin
	if user.RoleID.Int32 == RoleAdmin {
		admins, err := s.queries.CountAdminUsers(ctx)
		if err != nil {
			return RespondError(c, http.StatusInternalServerError, "db_error",
				"Failed to verify admin count.")
		}
		if admins <= 1 {
			return RespondError(c, http.StatusForbidden, "last_admin",
				"Cannot delete the last administrator. At least one admin must exist in the system.")
		}
	}

	// Soft delete
	err = s.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to delete user.")
	}

	// Log audit
	currentUserID, _ := middleware.GetUserIDFromContext(c)
	s.logAudit(ctx, currentUserID, "delete", "user", user.ID.String(),
		map[string]any{
			"username": user.Username,
			"deleted":  false,
		},
		map[string]any{
			"deleted": true,
		},
		c.RealIP(), c.Request().UserAgent())

	return c.NoContent(http.StatusNoContent)
}
