package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
)

// CreateCategoryReq defines the request body for creating a new category.
type CreateCategoryReq struct {
	Name string `json:"name"`
}

// NewCreateCategoryHandler creates a handler for POST /api/v1/categories.
func NewCreateCategoryHandler(queries *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req CreateCategoryReq
		if err := c.Bind(&req); err != nil {
			return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
		}

		// Basic validation
		if req.Name == "" {
			return RespondError(c, http.StatusBadRequest, "validation_error", "Category name cannot be empty.")
		}

		ctx := c.Request().Context()
		category, err := queries.CreateCategory(ctx, req.Name)
		if err != nil {
			// In a real app, you might check for specific DB errors, like unique constraints.
			return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to create category.")
		}

		return RespondSuccess(c, http.StatusCreated, category)
	}
}

// NewListCategoriesHandler creates a handler for GET /api/v1/categories.
func NewListCategoriesHandler(queries *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		categories, err := queries.ListCategories(ctx)
		if err != nil {
			return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve categories.")
		}

		// Return an empty slice instead of null if no categories are found.
		if categories == nil {
			categories = []db.Category{}
		}

		return RespondSuccess(c, http.StatusOK, categories)
	}
}

// NewGetCategoryHandler creates a handler for GET /api/v1/categories/:id.
func NewGetCategoryHandler(queries *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid number.")
		}

		ctx := c.Request().Context()
		category, err := queries.GetCategory(ctx, int32(id))
		if err != nil {
			if err == sql.ErrNoRows {
				return RespondError(c, http.StatusNotFound, "not_found", "Category with the specified ID was not found.")
			}
			return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve category.")
		}

		return RespondSuccess(c, http.StatusOK, category)
	}
}