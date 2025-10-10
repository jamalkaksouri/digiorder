package handlers

import (
    "database/sql"
    "net/http"
    "strconv"

    "github.com/labstack/echo/v4"
    dbq "github.com/jamalkaksouri/DigiOrder/internal/db"
)

// CreateProduct request body
type CreateProductReq struct {
    Name          string `json:"name"`
    Brand         string `json:"brand,omitempty"`
    DosageFormID  int32  `json:"dosage_form_id"`
    Strength      string `json:"strength,omitempty"`
    Unit          string `json:"unit,omitempty"`
    CategoryID    int32  `json:"category_id"`
    Description   string `json:"description,omitempty"`
}

// Handler factory
func NewCreateProductHandler(db *sql.DB, queries *dbq.Queries) echo.HandlerFunc {
    return func(c echo.Context) error {
        var req CreateProductReq
        if err := c.Bind(&req); err != nil {
            return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
        }

        ctx := c.Request().Context()

        product, err := queries.CreateProduct(ctx, dbq.CreateProductParams{
            Name:         req.Name,
            Brand:        sql.NullString{String: req.Brand, Valid: req.Brand != ""},
            DosageFormID: sql.NullInt32{Int32: req.DosageFormID, Valid: true},
            Strength:     sql.NullString{String: req.Strength, Valid: req.Strength != ""},
            Unit:         sql.NullString{String: req.Unit, Valid: req.Unit != ""},
            CategoryID:   sql.NullInt32{Int32: req.CategoryID, Valid: true},
            Description:  sql.NullString{String: req.Description, Valid: req.Description != ""},
        })
        if err != nil {
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create product"})
        }

        return c.JSON(http.StatusCreated, product)
    }
}

// Handler برای لیست محصولات با pagination
func NewListProductsHandler(queries *dbq.Queries) echo.HandlerFunc {
    return func(c echo.Context) error {
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

        products, err := queries.ListProducts(ctx, dbq.ListProductsParams{
            Limit:  int32(limit),
            Offset: int32(offset),
        })
        if err != nil {
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch products"})
        }

        return c.JSON(http.StatusOK, products)
    }
}
