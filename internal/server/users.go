// internal/server/users.go - ENHANCED ERROR HANDLING VERSION
package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/jamalkaksouri/DigiOrder/internal/middleware"
	"github.com/jamalkaksouri/DigiOrder/internal/security"
	"github.com/labstack/echo/v4"
)

const (
	RoleAdmin      = 1
	PrimaryAdminID = "00000000-0000-0000-0000-000000000001"
)

type CreateUserReq struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	FullName string `json:"full_name,omitempty"`
	Password string `json:"password" validate:"required,min=12"`
	RoleID   int32  `json:"role_id" validate:"required,gt=0"`
}

type UpdateUserReq struct {
	FullName string `json:"full_name,omitempty"`
	RoleID   *int32 `json:"role_id,omitempty"`
}

// CreateUser handles POST /api/v1/users (Admin only)
func (s *Server) CreateUser(c echo.Context) error {
	// Verify admin role
	roleName, err := middleware.GetRoleNameFromContext(c)
	if err != nil {
		return err // Already formatted error from middleware
	}
	if roleName != "admin" {
		return RespondError(c, http.StatusForbidden, "insufficient_permissions",
			"Only administrators can create users.")
	}

	var req CreateUserReq
	if err := s.ValidateRequest(c, &req); err != nil {
		return err // Already formatted by ValidateRequest
	}

	ctx := c.Request().Context()

	// Validate password strength
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

	// Hash password
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to hash password", err, nil)
		}
		return RespondError(c, http.StatusInternalServerError, "hash_error",
			"Failed to process password. Please try again.")
	}

	// Verify role exists
	role, err := s.queries.GetRole(ctx, req.RoleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusBadRequest, "invalid_role",
				fmt.Sprintf("Role with ID %d does not exist.", req.RoleID))
		}
		return HandleDatabaseError(c, err, "Role")
	}

	// Create user
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Username:     req.Username,
		FullName:     sql.NullString{String: req.FullName, Valid: req.FullName != ""},
		PasswordHash: hashedPassword,
		RoleID:       sql.NullInt32{Int32: req.RoleID, Valid: true},
	})
	if err != nil {
		return HandleDatabaseError(c, err, "User")
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
	id, err := ParseUUID(c, "id")
	if err != nil {
		return err // Already formatted by ParseUUID
	}

	ctx := c.Request().Context()
	user, err := s.queries.GetUser(ctx, id)
	if err != nil {
		return HandleDatabaseError(c, err, "User")
	}

	// Check if user is deleted
	if user.DeletedAt.Valid {
		return RespondError(c, http.StatusNotFound, "not_found",
			"User has been deleted and is no longer available.")
	}

	// Get role name
	var roleName string
	if user.RoleID.Valid {
		role, err := s.queries.GetRole(ctx, user.RoleID.Int32)
		if err != nil {
			// Don't fail the entire request if role fetch fails
			if s.logger != nil {
				s.logger.Error("Failed to fetch role name", err, map[string]any{
					"user_id": user.ID,
					"role_id": user.RoleID.Int32,
				})
			}
		} else {
			roleName = role.Name
		}
	}

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

	limit := 50
	offset := 0

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			return RespondError(c, http.StatusBadRequest, "invalid_limit",
				"Limit parameter must be a valid number.")
		}
		if parsedLimit <= 0 {
			return RespondError(c, http.StatusBadRequest, "invalid_limit",
				"Limit must be greater than 0.")
		}
		if parsedLimit > 100 {
			return RespondError(c, http.StatusBadRequest, "invalid_limit",
				"Limit cannot exceed 100.")
		}
		limit = parsedLimit
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return RespondError(c, http.StatusBadRequest, "invalid_offset",
				"Offset parameter must be a valid number.")
		}
		if parsedOffset < 0 {
			return RespondError(c, http.StatusBadRequest, "invalid_offset",
				"Offset cannot be negative.")
		}
		offset = parsedOffset
	}

	users, err := s.queries.ListActiveUsers(ctx, db.ListActiveUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return HandleDatabaseError(c, err, "Users")
	}

	if users == nil {
		users = []db.User{}
	}

	// Format response with role names
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
	id, err := ParseUUID(c, "id")
	if err != nil {
		return err
	}

	// Check if trying to update primary admin
	if id.String() == PrimaryAdminID {
		return RespondError(c, http.StatusForbidden, "protected_user",
			"The primary administrator account cannot be modified through this endpoint.")
	}

	var req UpdateUserReq
	if err := s.ValidateRequest(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Get old values for audit
	oldUser, err := s.queries.GetUser(ctx, id)
	if err != nil {
		return HandleDatabaseError(c, err, "User")
	}

	if oldUser.DeletedAt.Valid {
		return RespondError(c, http.StatusNotFound, "not_found",
			"User has been deleted and cannot be updated.")
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
					fmt.Sprintf("Role with ID %d does not exist.", *req.RoleID))
			}
			return HandleDatabaseError(c, err, "Role")
		}
		params.RoleID = sql.NullInt32{Int32: *req.RoleID, Valid: true}
	}

	user, err := s.queries.UpdateUser(ctx, params)
	if err != nil {
		return HandleDatabaseError(c, err, "User")
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

	user.PasswordHash = ""
	return RespondSuccess(c, http.StatusOK, user)
}

// DeleteUser handles DELETE /api/v1/users/:id (Soft delete)
func (s *Server) DeleteUser(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return err
	}

	// CRITICAL: Protect primary admin
	if id.String() == PrimaryAdminID {
		return RespondError(c, http.StatusForbidden, "protected_user",
			"The primary administrator account cannot be deleted. This account is essential for system administration.")
	}

	ctx := c.Request().Context()

	// Get user for checks
	user, err := s.queries.GetUser(ctx, id)
	if err != nil {
		return HandleDatabaseError(c, err, "User")
	}

	if user.DeletedAt.Valid {
		return RespondError(c, http.StatusNotFound, "not_found",
			"User has already been deleted.")
	}

	// Check if user is the last admin
	if user.RoleID.Int32 == RoleAdmin {
		admins, err := s.queries.CountAdminUsers(ctx)
		if err != nil {
			if s.logger != nil {
				s.logger.Error("Failed to count admin users", err, nil)
			}
			return RespondError(c, http.StatusInternalServerError, "database_error",
				"Failed to verify admin count. Please try again.")
		}
		if admins <= 1 {
			return RespondError(c, http.StatusForbidden, "last_admin",
				"Cannot delete the last administrator. At least one admin must exist in the system.")
		}
	}

	// Soft delete
	err = s.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		return HandleDatabaseError(c, err, "User")
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
