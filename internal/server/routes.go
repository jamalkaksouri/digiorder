package server

import (
	"net/http"

	"github.com/jamalkaksouri/DigiOrder/internal/handlers"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) registerRoutes() {
	// Middleware
	s.router.Use(middleware.Logger())
	s.router.Use(middleware.Recover())
	
	// Set custom error handler
	s.router.HTTPErrorHandler = customHTTPErrorHandler

	// Group routes under /api/v1
	api := s.router.Group("/api/v1")

	// Product routes
	api.POST("/products", s.CreateProduct)
	api.GET("/products", handlers.NewListProductsHandler(s.queries)) // Can be refactored later

	// Category routes
	api.POST("/categories", s.CreateCategory)
	api.GET("/categories", s.ListCategories)
	api.GET("/categories/:id", s.GetCategory)

	// Dosage Form routes
	api.POST("/dosage_forms", s.CreateDosageForm)
	api.GET("/dosage_forms", s.ListDosageForms)
	api.GET("/dosage_forms/:id", s.GetDosageForm)
}

// Custom HTTP error handler can also be part of the server package
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	msg := "internal_server_error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			msg = m
		} else {
			msg = http.StatusText(code)
		}
	}

	if !c.Response().Committed {
		c.JSON(code, handlers.ErrorResponse{
			Error:   msg,
			Details: err.Error(),
		})
	}
}