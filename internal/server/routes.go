package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) registerRoutes() {
	// Middleware
	s.router.Use(middleware.Logger())
	s.router.Use(middleware.Recover())
	s.router.Use(middleware.CORS())
	
	// Set custom error handler
	s.router.HTTPErrorHandler = s.customHTTPErrorHandler

	// Health check
	s.router.GET("/health", s.healthCheck)

	// Group routes under /api/v1
	api := s.router.Group("/api/v1")

	// Product routes
	api.POST("/products", s.CreateProduct)
	api.GET("/products", s.ListProducts)
	api.GET("/products/:id", s.GetProduct)
	api.PUT("/products/:id", s.UpdateProduct)
	api.DELETE("/products/:id", s.DeleteProduct)
	api.GET("/products/search", s.SearchProducts)

	// Category routes
	api.POST("/categories", s.CreateCategory)
	api.GET("/categories", s.ListCategories)
	api.GET("/categories/:id", s.GetCategory)

	// Dosage Form routes
	api.POST("/dosage_forms", s.CreateDosageForm)
	api.GET("/dosage_forms", s.ListDosageForms)
	api.GET("/dosage_forms/:id", s.GetDosageForm)

	// Order routes
	api.POST("/orders", s.CreateOrder)
	api.GET("/orders", s.ListOrders)
	api.GET("/orders/:id", s.GetOrder)
	api.PUT("/orders/:id/status", s.UpdateOrderStatus)
	api.DELETE("/orders/:id", s.DeleteOrder)

	// Order items routes
	api.POST("/orders/:order_id/items", s.CreateOrderItem)
	api.GET("/orders/:order_id/items", s.GetOrderItems)
	api.PUT("/order_items/:id", s.UpdateOrderItem)
	api.DELETE("/order_items/:id", s.DeleteOrderItem)

	// User routes
	api.POST("/users", s.CreateUser)
	api.GET("/users", s.ListUsers)
	api.GET("/users/:id", s.GetUser)
	api.PUT("/users/:id", s.UpdateUser)
	api.DELETE("/users/:id", s.DeleteUser)

	// Role routes
	api.POST("/roles", s.CreateRole)
	api.GET("/roles", s.ListRoles)
	api.GET("/roles/:id", s.GetRole)
	api.PUT("/roles/:id", s.UpdateRole)
	api.DELETE("/roles/:id", s.DeleteRole)
}

// Health check endpoint
func (s *Server) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
		"service": "DigiOrder API",
	})
}

// Custom HTTP error handler
func (s *Server) customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	msg := "internal_server_error"
	details := ""

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			msg = m
		} else {
			msg = http.StatusText(code)
		}
		if he.Internal != nil {
			details = he.Internal.Error()
		}
	} else {
		details = err.Error()
	}

	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			c.NoContent(code)
		} else {
			c.JSON(code, ErrorResponse{
				Error:   msg,
				Details: details,
			})
		}
	}
}