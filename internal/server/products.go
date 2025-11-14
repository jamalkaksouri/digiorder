// internal/server/products.go - ENHANCED ERROR HANDLING VERSION
package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
)

type CreateProductReq struct {
	Name         string `json:"name" validate:"required,min=1,max=255"`
	Brand        string `json:"brand,omitempty"`
	DosageFormID int32  `json:"dosage_form_id" validate:"required,gt=0"`
	Strength     string `json:"strength,omitempty"`
	Unit         string `json:"unit,omitempty"`
	CategoryID   int32  `json:"category_id" validate:"required,gt=0"`
	Description  string `json:"description,omitempty"`
}

type UpdateProductReq struct {
	Name         string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Brand        string `json:"brand,omitempty"`
	DosageFormID *int32 `json:"dosage_form_id,omitempty" validate:"omitempty,gt=0"`
	Strength     string `json:"strength,omitempty"`
	Unit         string `json:"unit,omitempty"`
	CategoryID   *int32 `json:"category_id,omitempty" validate:"omitempty,gt=0"`
	Description  string `json:"description,omitempty"`
}

// CreateProduct handles POST /api/v1/products
func (s *Server) CreateProduct(c echo.Context) error {
	var req CreateProductReq
	if err := s.ValidateRequest(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Verify dosage form exists
	_, err := s.queries.GetDosageForm(ctx, req.DosageFormID)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusBadRequest, "invalid_dosage_form",
				fmt.Sprintf("Dosage form with ID %d does not exist.", req.DosageFormID))
		}
		return HandleDatabaseError(c, err, "Dosage Form")
	}

	// Verify category exists
	_, err = s.queries.GetCategory(ctx, req.CategoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusBadRequest, "invalid_category",
				fmt.Sprintf("Category with ID %d does not exist.", req.CategoryID))
		}
		return HandleDatabaseError(c, err, "Category")
	}

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
		return HandleDatabaseError(c, err, "Product")
	}

	return RespondSuccess(c, http.StatusCreated, product)
}

// ListProducts handles GET /api/v1/products
func (s *Server) ListProducts(c echo.Context) error {
	ctx := c.Request().Context()

	limit := 50
	offset := 0

	if limitStr := c.QueryParam("limit"); limitStr != "" {
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

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
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

	products, err := s.queries.ListProducts(ctx, db.ListProductsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return HandleDatabaseError(c, err, "Products")
	}

	if products == nil {
		products = []db.Product{}
	}

	return RespondSuccess(c, http.StatusOK, products)
}

// GetProduct handles GET /api/v1/products/:id
func (s *Server) GetProduct(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	product, err := s.queries.GetProduct(ctx, id)
	if err != nil {
		return HandleDatabaseError(c, err, "Product")
	}

	// Check if product is deleted
	if product.DeletedAt.Valid {
		return RespondError(c, http.StatusNotFound, "not_found",
			"Product has been deleted and is no longer available.")
	}

	return RespondSuccess(c, http.StatusOK, product)
}

// UpdateProduct handles PUT /api/v1/products/:id
func (s *Server) UpdateProduct(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return err
	}

	var req UpdateProductReq
	if err := s.ValidateRequest(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Verify product exists
	existingProduct, err := s.queries.GetProduct(ctx, id)
	if err != nil {
		return HandleDatabaseError(c, err, "Product")
	}

	if existingProduct.DeletedAt.Valid {
		return RespondError(c, http.StatusNotFound, "not_found",
			"Product has been deleted and cannot be updated.")
	}

	// Verify dosage form if provided
	if req.DosageFormID != nil {
		_, err := s.queries.GetDosageForm(ctx, *req.DosageFormID)
		if err != nil {
			if err == sql.ErrNoRows {
				return RespondError(c, http.StatusBadRequest, "invalid_dosage_form",
					fmt.Sprintf("Dosage form with ID %d does not exist.", *req.DosageFormID))
			}
			return HandleDatabaseError(c, err, "Dosage Form")
		}
	}

	// Verify category if provided
	if req.CategoryID != nil {
		_, err := s.queries.GetCategory(ctx, *req.CategoryID)
		if err != nil {
			if err == sql.ErrNoRows {
				return RespondError(c, http.StatusBadRequest, "invalid_category",
					fmt.Sprintf("Category with ID %d does not exist.", *req.CategoryID))
			}
			return HandleDatabaseError(c, err, "Category")
		}
	}

	// Build update params
	params := db.UpdateProductParams{
		ID: id,
	}

	if req.Name != "" {
		params.Name = req.Name
	} else {
		params.Name = existingProduct.Name
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
		return HandleDatabaseError(c, err, "Product")
	}

	return RespondSuccess(c, http.StatusOK, product)
}

// DeleteProduct handles DELETE /api/v1/products/:id
func (s *Server) DeleteProduct(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Verify product exists
	product, err := s.queries.GetProduct(ctx, id)
	if err != nil {
		return HandleDatabaseError(c, err, "Product")
	}

	if product.DeletedAt.Valid {
		return RespondError(c, http.StatusNotFound, "not_found",
			"Product has already been deleted.")
	}

	err = s.queries.DeleteProduct(ctx, id)
	if err != nil {
		return HandleDatabaseError(c, err, "Product")
	}

	return c.NoContent(http.StatusNoContent)
}

// SearchProducts handles GET /api/v1/products/search
func (s *Server) SearchProducts(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return RespondError(c, http.StatusBadRequest, "missing_query",
			"Search query parameter 'q' is required.")
	}

	if len(query) < 2 {
		return RespondError(c, http.StatusBadRequest, "query_too_short",
			"Search query must be at least 2 characters long.")
	}

	limit := 50
	offset := 0

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			return RespondError(c, http.StatusBadRequest, "invalid_limit",
				"Limit parameter must be a valid number.")
		}
		if parsedLimit <= 0 || parsedLimit > 100 {
			return RespondError(c, http.StatusBadRequest, "invalid_limit",
				"Limit must be between 1 and 100.")
		}
		limit = parsedLimit
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
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

	ctx := c.Request().Context()
	products, err := s.queries.SearchProducts(ctx, db.SearchProductsParams{
		Column1: sql.NullString{String: query, Valid: true},
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return HandleDatabaseError(c, err, "Products")
	}

	if products == nil {
		products = []db.Product{}
	}

	return RespondSuccess(c, http.StatusOK, products)
}
