package server

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
)

// CreateProductReq defines the request body for creating a product
type CreateProductReq struct {
	Name         string `json:"name" validate:"required"`
	Brand        string `json:"brand,omitempty"`
	DosageFormID int32  `json:"dosage_form_id" validate:"required,gt=0"`
	Strength     string `json:"strength,omitempty"`
	Unit         string `json:"unit,omitempty"`
	CategoryID   int32  `json:"category_id" validate:"required,gt=0"`
	Description  string `json:"description,omitempty"`
}

// UpdateProductReq defines the request body for updating a product
type UpdateProductReq struct {
	Name         string `json:"name,omitempty"`
	Brand        string `json:"brand,omitempty"`
	DosageFormID *int32 `json:"dosage_form_id,omitempty"`
	Strength     string `json:"strength,omitempty"`
	Unit         string `json:"unit,omitempty"`
	CategoryID   *int32 `json:"category_id,omitempty"`
	Description  string `json:"description,omitempty"`
}

// CreateProduct handles POST /api/v1/products
func (s *Server) CreateProduct(c echo.Context) error {
	var req CreateProductReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()

	// Create product
	product, err := s.queries.CreateProduct(ctx, db.CreateProductParams{
		Name:         req.Name,
		Brand:        sql.NullString{String: req.Brand, Valid: req.Brand != ""},
		DosageFormID: sql.NullInt32{Int32: req.DosageFormID, Valid: true},
		Strength:     sql.NullString{String: req.Strength, Valid: req.Strength != ""},
		Unit:         sql.NullString{String: req.Unit, Valid: req.Unit != ""},
		CategoryID:   sql.NullInt32{Int32: req.CategoryID, Valid: true},
		Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to create product.")
	}

	return RespondSuccess(c, http.StatusCreated, product)
}

// ListProducts handles GET /api/v1/products
func (s *Server) ListProducts(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse pagination parameters
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

	// Fetch products
	products, err := s.queries.ListProducts(ctx, db.ListProductsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to fetch products.")
	}

	// Return empty array instead of null
	if products == nil {
		products = []db.Product{}
	}

	return RespondSuccess(c, http.StatusOK, products)
}

// GetProduct handles GET /api/v1/products/:id
func (s *Server) GetProduct(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid UUID.")
	}

	ctx := c.Request().Context()
	product, err := s.queries.GetProduct(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", "Product with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve product.")
	}

	return RespondSuccess(c, http.StatusOK, product)
}

// UpdateProduct handles PUT /api/v1/products/:id
func (s *Server) UpdateProduct(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid UUID.")
	}

	var req UpdateProductReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	ctx := c.Request().Context()

	// Build update params
	params := db.UpdateProductParams{
		ID:   id,
		Name: req.Name,
	}

	if req.Brand != "" {
		params.Brand = sql.NullString{String: req.Brand, Valid: true}
	}
	if req.DosageFormID != nil {
		params.DosageFormID = sql.NullInt32{Int32: *req.DosageFormID, Valid: true}
	}
	if req.Strength != "" {
		params.Strength = sql.NullString{String: req.Strength, Valid: true}
	}
	if req.Unit != "" {
		params.Unit = sql.NullString{String: req.Unit, Valid: true}
	}
	if req.CategoryID != nil {
		params.CategoryID = sql.NullInt32{Int32: *req.CategoryID, Valid: true}
	}
	if req.Description != "" {
		params.Description = sql.NullString{String: req.Description, Valid: true}
	}

	product, err := s.queries.UpdateProduct(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", "Product with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to update product.")
	}

	return RespondSuccess(c, http.StatusOK, product)
}

// DeleteProduct handles DELETE /api/v1/products/:id
func (s *Server) DeleteProduct(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid UUID.")
	}

	ctx := c.Request().Context()
	err = s.queries.DeleteProduct(ctx, id)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to delete product.")
	}

	return c.NoContent(http.StatusNoContent)
}

// SearchProducts handles GET /api/v1/products/search
func (s *Server) SearchProducts(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return RespondError(c, http.StatusBadRequest, "missing_query", "Search query parameter 'q' is required.")
	}

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

	ctx := c.Request().Context()
	products, err := s.queries.SearchProducts(ctx, db.SearchProductsParams{
		Column1: sql.NullString{String: query, Valid: true},
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to search products.")
	}

	if products == nil {
		products = []db.Product{}
	}

	return RespondSuccess(c, http.StatusOK, products)
}