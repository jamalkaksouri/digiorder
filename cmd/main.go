package main

import (
	"log"
	"net/http"

	"github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/jamalkaksouri/DigiOrder/internal/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Set custom error handler
	e.HTTPErrorHandler = customHTTPErrorHandler

	// Database connection
	database, err := db.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Initialize queries
	queries := db.New(database)

	// Routes
	e.POST("/api/v1/products", handlers.NewCreateProductHandler(database, queries))
	e.GET("/api/v1/products", handlers.NewListProductsHandler(queries))

	// --- ADD THESE NEW ROUTES FOR CATEGORIES ---
	e.POST("/api/v1/categories", handlers.NewCreateCategoryHandler(queries))
	e.GET("/api/v1/categories", handlers.NewListCategoriesHandler(queries))
	e.GET("/api/v1/categories/:id", handlers.NewGetCategoryHandler(queries))
	// -----------------------------------------

	// --- ADD THESE NEW ROUTES FOR DOSAGE FORMS ---
	e.POST("/api/v1/dosage_forms", handlers.NewCreateDosageFormHandler(queries))
	e.GET("/api/v1/dosage_forms", handlers.NewListDosageFormsHandler(queries))
	e.GET("/api/v1/dosage_forms/:id", handlers.NewGetDosageFormHandler(queries))
	// -------------------------------------------

	// Start server
	e.Logger.Fatal(e.Start(":5582"))
}

// Custom HTTP error handler
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	msg := "internal_server_error"

	// Check error type
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			msg = m
		} else {
			msg = http.StatusText(code)
		}
	}

	// Standard JSON response
	if !c.Response().Committed {
		c.JSON(code, handlers.ErrorResponse{
			Error:   msg,
			Details: err.Error(),
		})
	}
}
