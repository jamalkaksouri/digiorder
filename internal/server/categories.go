package server

import (
	"database/sql"
	"net/http"
	"strconv"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
)

// CreateCategoryReq defines the request body for creating a new category.
type CreateCategoryReq struct {
	Name string `json:"name" validate:"required"`
}

// CreateCategory handles POST /api/v1/categories
func (s *Server) CreateCategory(c echo.Context) error {
	var req CreateCategoryReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()
	category, err := s.queries.CreateCategory(ctx, req.Name)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to create category.")
	}

	return RespondSuccess(c, http.StatusCreated, category)
}

// ListCategories handles GET /api/v1/categories
func (s *Server) ListCategories(c echo.Context) error {
	ctx := c.Request().Context()
	categories, err := s.queries.ListCategories(ctx)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve categories.")
	}

	if categories == nil {
		categories = []db.Category{}
	}

	return RespondSuccess(c, http.StatusOK, categories)
}

// GetCategory handles GET /api/v1/categories/:id
func (s *Server) GetCategory(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid number.")
	}

	ctx := c.Request().Context()
	category, err := s.queries.GetCategory(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", "Category with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve category.")
	}

	return RespondSuccess(c, http.StatusOK, category)
}