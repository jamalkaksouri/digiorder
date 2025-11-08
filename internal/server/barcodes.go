package server

import (
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
)

// CreateBarcodeReq defines the request for creating a barcode
type CreateBarcodeReq struct {
	ProductID   string `json:"product_id" validate:"required,uuid"`
	Barcode     string `json:"barcode" validate:"required"`
	BarcodeType string `json:"barcode_type,omitempty"` // EAN-13, UPC-A, Code128, etc.
}

// UpdateBarcodeReq defines the request for updating a barcode
type UpdateBarcodeReq struct {
	Barcode     string `json:"barcode,omitempty"`
	BarcodeType string `json:"barcode_type,omitempty"`
}

// SearchByBarcodeReq defines the request for barcode search
type SearchByBarcodeReq struct {
	Barcode string `json:"barcode" validate:"required"`
}

// CreateBarcode handles POST /api/v1/barcodes
func (s *Server) CreateBarcode(c echo.Context) error {
	var req CreateBarcodeReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_product_id", "Invalid product ID format.")
	}

	ctx := c.Request().Context()

	// Verify product exists
	_, err = s.queries.GetProduct(ctx, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "product_not_found", "Product does not exist.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to verify product.")
	}

	// Create barcode
	barcode, err := s.queries.CreateBarcode(ctx, db.CreateBarcodeParams{
		ProductID:   uuid.NullUUID{UUID: productID, Valid: true},
		Barcode:     req.Barcode,
		BarcodeType: sql.NullString{String: req.BarcodeType, Valid: req.BarcodeType != ""},
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to create barcode.")
	}

	return RespondSuccess(c, http.StatusCreated, barcode)
}

// GetBarcodesByProduct handles GET /api/v1/products/:product_id/barcodes
func (s *Server) GetBarcodesByProduct(c echo.Context) error {
	productIDStr := c.Param("product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_product_id", "Invalid product ID format.")
	}

	ctx := c.Request().Context()
	barcodes, err := s.queries.GetBarcodesByProduct(ctx, uuid.NullUUID{UUID: productID, Valid: true})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve barcodes.")
	}

	if barcodes == nil {
		barcodes = []db.ProductBarcode{}
	}

	return RespondSuccess(c, http.StatusOK, barcodes)
}

// SearchProductByBarcode handles GET /api/v1/products/barcode/:barcode
func (s *Server) SearchProductByBarcode(c echo.Context) error {
	barcode := c.Param("barcode")
	if barcode == "" {
		return RespondError(c, http.StatusBadRequest, "missing_barcode", "Barcode parameter is required.")
	}

	ctx := c.Request().Context()
	product, err := s.queries.GetProductByBarcode(ctx, barcode)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", "Product with this barcode not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to search product.")
	}

	return RespondSuccess(c, http.StatusOK, product)
}

// UpdateBarcode handles PUT /api/v1/barcodes/:id
func (s *Server) UpdateBarcode(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "Invalid barcode ID format.")
	}

	var req UpdateBarcodeReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	ctx := c.Request().Context()

	params := db.UpdateBarcodeParams{
		ID: id,
	}

	if req.Barcode != "" {
		params.Barcode = req.Barcode
	}
	if req.BarcodeType != "" {
		params.BarcodeType = sql.NullString{String: req.BarcodeType, Valid: true}
	}

	barcode, err := s.queries.UpdateBarcode(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", "Barcode not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to update barcode.")
	}

	return RespondSuccess(c, http.StatusOK, barcode)
}

// DeleteBarcode handles DELETE /api/v1/barcodes/:id
func (s *Server) DeleteBarcode(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "Invalid barcode ID format.")
	}

	ctx := c.Request().Context()
	err = s.queries.DeleteBarcode(ctx, id)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to delete barcode.")
	}

	return c.NoContent(http.StatusNoContent)
}