package server

import (
	"database/sql"
	"net/http"
	"strconv"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
)

// CreateRoleReq defines the request body for creating a new role
type CreateRoleReq struct {
	Name string `json:"name" validate:"required"`
}

// UpdateRoleReq defines the request body for updating a role
type UpdateRoleReq struct {
	Name string `json:"name" validate:"required"`
}

// CreateRole handles POST /api/v1/roles
func (s *Server) CreateRole(c echo.Context) error {
	var req CreateRoleReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()
	role, err := s.queries.CreateRole(ctx, req.Name)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to create role.")
	}

	return RespondSuccess(c, http.StatusCreated, role)
}

// ListRoles handles GET /api/v1/roles
func (s *Server) ListRoles(c echo.Context) error {
	ctx := c.Request().Context()
	roles, err := s.queries.ListRoles(ctx)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve roles.")
	}

	if roles == nil {
		roles = []db.Role{}
	}

	return RespondSuccess(c, http.StatusOK, roles)
}

// GetRole handles GET /api/v1/roles/:id
func (s *Server) GetRole(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid number.")
	}

	ctx := c.Request().Context()
	role, err := s.queries.GetRole(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", "Role with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve role.")
	}

	return RespondSuccess(c, http.StatusOK, role)
}

// UpdateRole handles PUT /api/v1/roles/:id
func (s *Server) UpdateRole(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid number.")
	}

	var req UpdateRoleReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()
	role, err := s.queries.UpdateRole(ctx, db.UpdateRoleParams{
		ID:   int32(id),
		Name: req.Name,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", "Role with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to update role.")
	}

	return RespondSuccess(c, http.StatusOK, role)
}

// DeleteRole handles DELETE /api/v1/roles/:id
func (s *Server) DeleteRole(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid number.")
	}

	ctx := c.Request().Context()
	err = s.queries.DeleteRole(ctx, int32(id))
	if err != nil {
		// Check if there are users still using this role
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to delete role. It may be in use by existing users.")
	}

	return c.NoContent(http.StatusNoContent)
}