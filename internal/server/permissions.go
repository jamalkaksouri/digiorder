// internal/server/permissions.go - FULLY DYNAMIC VERSION

package server

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/jamalkaksouri/DigiOrder/internal/middleware"
	"github.com/labstack/echo/v4"
)

// Permission request structures
type CreatePermissionReq struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Resource    string `json:"resource" validate:"required"`
	Action      string `json:"action" validate:"required"` // FULLY DYNAMIC - any action name allowed
	Description string `json:"description,omitempty"`
}

type UpdatePermissionReq struct {
	Name        string `json:"name,omitempty"`
	Resource    string `json:"resource,omitempty"`
	Action      string `json:"action,omitempty"` // FULLY DYNAMIC
	Description string `json:"description,omitempty"`
}

// CreatePermission handles POST /api/v1/permissions
// FULLY DYNAMIC - accepts any resource:action combination
func (s *Server) CreatePermission(c echo.Context) error {
	var req CreatePermissionReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()
	permission, err := s.queries.CreatePermission(ctx, db.CreatePermissionParams{
		Name:        req.Name,
		Resource:    req.Resource,
		Action:      req.Action, // Any action is allowed
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		// FIXED: Check for duplicate permission
		if strings.Contains(err.Error(), "duplicate") ||
			strings.Contains(err.Error(), "unique constraint") {
			// Check which constraint was violated
			if strings.Contains(err.Error(), "permissions_name_key") {
				return RespondError(c, http.StatusConflict, "duplicate_permission_name",
					"A permission with this name already exists.")
			}
			if strings.Contains(err.Error(), "permissions_resource_action_key") {
				return RespondError(c, http.StatusConflict, "duplicate_permission",
					"A permission with this resource:action combination already exists.")
			}
			return RespondError(c, http.StatusConflict, "duplicate_permission",
				"This permission already exists.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to create permission.")
	}

	// Log audit
	currentUserID, _ := middleware.GetUserIDFromContext(c)
	s.logAudit(ctx, currentUserID, "create", "permission", strconv.Itoa(int(permission.ID)),
		nil, map[string]any{
			"name":     permission.Name,
			"resource": permission.Resource,
			"action":   permission.Action,
		}, c.RealIP(), c.Request().UserAgent())

	return RespondSuccess(c, http.StatusCreated, permission)
}

// ListPermissions handles GET /api/v1/permissions
func (s *Server) ListPermissions(c echo.Context) error {
	ctx := c.Request().Context()

	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")
	resource := c.QueryParam("resource")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var permissions []db.Permission
	if resource != "" {
		permissions, err = s.queries.ListPermissionsByResource(ctx, db.ListPermissionsByResourceParams{
			Resource: resource,
			Limit:    int32(limit),
			Offset:   int32(offset),
		})
	} else {
		permissions, err = s.queries.ListPermissions(ctx, db.ListPermissionsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
	}

	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve permissions.")
	}

	if permissions == nil {
		permissions = []db.Permission{}
	}

	return RespondSuccess(c, http.StatusOK, permissions)
}

// GetPermission handles GET /api/v1/permissions/:id
func (s *Server) GetPermission(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id",
			"The provided ID is not a valid number.")
	}

	ctx := c.Request().Context()
	permission, err := s.queries.GetPermission(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found",
				"Permission with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve permission.")
	}

	return RespondSuccess(c, http.StatusOK, permission)
}

// UpdatePermission handles PUT /api/v1/permissions/:id
func (s *Server) UpdatePermission(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id",
			"The provided ID is not a valid number.")
	}

	var req UpdatePermissionReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	ctx := c.Request().Context()

	// Get old values for audit
	oldPermission, err := s.queries.GetPermission(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found",
				"Permission with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve permission.")
	}

	params := db.UpdatePermissionParams{
		ID: int32(id),
	}

	if req.Name != "" {
		params.Name = sql.NullString{String: req.Name, Valid: true}
	}
	if req.Resource != "" {
		params.Resource = sql.NullString{String: req.Resource, Valid: true}
	}
	if req.Action != "" {
		params.Action = sql.NullString{String: req.Action, Valid: true} // Any action allowed
	}
	if req.Description != "" {
		params.Description = sql.NullString{String: req.Description, Valid: true}
	}

	permission, err := s.queries.UpdatePermission(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found",
				"Permission with the specified ID was not found.")
		}
		// FIXED: Check for duplicate
		if strings.Contains(err.Error(), "duplicate") ||
			strings.Contains(err.Error(), "unique constraint") {
			if strings.Contains(err.Error(), "permissions_name_key") {
				return RespondError(c, http.StatusConflict, "duplicate_permission_name",
					"A permission with this name already exists.")
			}
			if strings.Contains(err.Error(), "permissions_resource_action_key") {
				return RespondError(c, http.StatusConflict, "duplicate_permission",
					"A permission with this resource:action combination already exists.")
			}
			return RespondError(c, http.StatusConflict, "duplicate_permission",
				"This permission already exists.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to update permission.")
	}

	// Log audit
	currentUserID, _ := middleware.GetUserIDFromContext(c)
	s.logAudit(ctx, currentUserID, "update", "permission", strconv.Itoa(int(permission.ID)),
		map[string]any{
			"name":     oldPermission.Name,
			"resource": oldPermission.Resource,
			"action":   oldPermission.Action,
		},
		map[string]any{
			"name":     permission.Name,
			"resource": permission.Resource,
			"action":   permission.Action,
		}, c.RealIP(), c.Request().UserAgent())

	return RespondSuccess(c, http.StatusOK, permission)
}

// DeletePermission handles DELETE /api/v1/permissions/:id
func (s *Server) DeletePermission(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id",
			"The provided ID is not a valid number.")
	}

	ctx := c.Request().Context()

	// Get permission for audit before deletion
	permission, err := s.queries.GetPermission(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found",
				"Permission with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve permission.")
	}

	err = s.queries.DeletePermission(ctx, int32(id))
	if err != nil {
		// Check if permission is in use
		if strings.Contains(err.Error(), "foreign key") ||
			strings.Contains(err.Error(), "violates foreign key") {
			return RespondError(c, http.StatusConflict, "permission_in_use",
				"Cannot delete permission because it is assigned to roles.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to delete permission.")
	}

	// Log audit
	currentUserID, _ := middleware.GetUserIDFromContext(c)
	s.logAudit(ctx, currentUserID, "delete", "permission", strconv.Itoa(int(permission.ID)),
		map[string]any{
			"name":     permission.Name,
			"resource": permission.Resource,
			"action":   permission.Action,
		}, nil, c.RealIP(), c.Request().UserAgent())

	return c.NoContent(http.StatusNoContent)
}

// AssignPermissionToRole handles POST /api/v1/roles/:role_id/permissions
func (s *Server) AssignPermissionToRole(c echo.Context) error {
	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_role_id",
			"The provided role ID is not a valid number.")
	}

	var req struct {
		PermissionID int32 `json:"permission_id" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()

	// Verify role exists
	_, err = s.queries.GetRole(ctx, int32(roleID))
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "role_not_found",
				"Role with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to verify role.")
	}

	// Verify permission exists
	permission, err := s.queries.GetPermission(ctx, req.PermissionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "permission_not_found",
				"Permission with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to verify permission.")
	}

	// Assign permission to role
	rolePermission, err := s.queries.AssignPermissionToRole(ctx, db.AssignPermissionToRoleParams{
		RoleID:       int32(roleID),
		PermissionID: req.PermissionID,
	})
	if err != nil {
		// FIXED: Check for duplicate assignment
		if strings.Contains(err.Error(), "duplicate") ||
			strings.Contains(err.Error(), "unique constraint") {
			return RespondError(c, http.StatusConflict, "permission_already_assigned",
				"This permission is already assigned to this role.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to assign permission to role.")
	}

	// Log audit
	currentUserID, _ := middleware.GetUserIDFromContext(c)
	s.logAudit(ctx, currentUserID, "assign", "role_permission",
		strconv.Itoa(int(rolePermission.ID)), nil, map[string]any{
			"role_id":       roleID,
			"permission_id": req.PermissionID,
			"permission":    permission.Name,
		}, c.RealIP(), c.Request().UserAgent())

	return RespondSuccess(c, http.StatusCreated, rolePermission)
}

// RevokePermissionFromRole handles DELETE /api/v1/roles/:role_id/permissions/:permission_id
func (s *Server) RevokePermissionFromRole(c echo.Context) error {
	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_role_id",
			"The provided role ID is not a valid number.")
	}

	permissionIDStr := c.Param("permission_id")
	permissionID, err := strconv.ParseInt(permissionIDStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_permission_id",
			"The provided permission ID is not a valid number.")
	}

	ctx := c.Request().Context()

	err = s.queries.RevokePermissionFromRole(ctx, db.RevokePermissionFromRoleParams{
		RoleID:       int32(roleID),
		PermissionID: int32(permissionID),
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to revoke permission from role.")
	}

	// Log audit
	currentUserID, _ := middleware.GetUserIDFromContext(c)
	s.logAudit(ctx, currentUserID, "revoke", "role_permission", "",
		map[string]any{
			"role_id":       roleID,
			"permission_id": permissionID,
		}, nil, c.RealIP(), c.Request().UserAgent())

	return c.NoContent(http.StatusNoContent)
}

// GetRolePermissions handles GET /api/v1/roles/:role_id/permissions
func (s *Server) GetRolePermissions(c echo.Context) error {
	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_role_id",
			"The provided role ID is not a valid number.")
	}

	ctx := c.Request().Context()
	permissions, err := s.queries.GetRolePermissions(ctx, int32(roleID))
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve role permissions.")
	}

	if permissions == nil {
		permissions = []db.Permission{}
	}

	return RespondSuccess(c, http.StatusOK, permissions)
}

// CheckUserPermission handles GET /api/v1/auth/check-permission
// FULLY DYNAMIC - checks any resource:action combination
func (s *Server) CheckUserPermission(c echo.Context) error {
	resource := c.QueryParam("resource")
	action := c.QueryParam("action")

	if resource == "" || action == "" {
		return RespondError(c, http.StatusBadRequest, "missing_parameters",
			"Resource and action parameters are required.")
	}

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Get user's role
	user, err := s.queries.GetUser(ctx, userID)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve user information.")
	}

	// Check if user has permission - works with ANY action
	hasPermission, err := s.queries.CheckRolePermission(ctx, db.CheckRolePermissionParams{
		RoleID:   user.RoleID.Int32,
		Resource: resource,
		Action:   action, // Any custom action like "tst" will work
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to check permission.")
	}

	return RespondSuccess(c, http.StatusOK, map[string]any{
		"has_permission": hasPermission,
		"resource":       resource,
		"action":         action,
	})
}
