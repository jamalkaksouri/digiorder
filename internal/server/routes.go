package server

import (
	"net/http"
	"time"

	"github.com/jamalkaksouri/DigiOrder/internal/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func (s *Server) registerRoutes() {
	// Initialize metrics collector
	metricsCollector := middleware.NewMetricsCollector()
	
	// Initialize request logger
	requestLogger := middleware.NewRequestLogger(s.router.Logger)

	// Global middleware
	s.router.Use(echomiddleware.Logger())
	s.router.Use(echomiddleware.Recover())
	s.router.Use(echomiddleware.CORS())
	s.router.Use(echomiddleware.RequestID())
	s.router.Use(echomiddleware.Secure())
	
	// Custom middleware
	s.router.Use(requestLogger.Middleware())
	s.router.Use(metricsCollector.Middleware())
	
	// Rate limiting - 100 requests per second, burst of 200
	s.router.Use(middleware.RateLimitMiddleware(100, 200))
	
	// Set custom error handler
	s.router.HTTPErrorHandler = s.customHTTPErrorHandler

	// Public endpoints
	s.router.GET("/health", s.healthCheck)
	s.router.GET("/metrics", func(c echo.Context) error {
		return c.JSON(http.StatusOK, metricsCollector.GetMetrics())
	})

	// API v1 group
	api := s.router.Group("/api/v1")

	// ==================== PUBLIC AUTH ENDPOINTS ====================
	auth := api.Group("/auth")
	{
		auth.POST("/login", s.Login)
		auth.POST("/refresh", s.RefreshToken)
	}

	// ==================== PROTECTED ENDPOINTS ====================
	// JWT middleware for all protected routes
	protected := api.Group("")
	protected.Use(middleware.JWTMiddleware())
	
	// API key rate limiting for authenticated users - 1000 requests per minute
	protected.Use(middleware.APIKeyRateLimitMiddleware(1000))

	// Auth profile endpoints (require authentication)
	{
		protected.GET("/auth/profile", s.GetProfile)
		protected.PUT("/auth/password", s.ChangePassword)
	}

	// Product routes (with caching for GET requests)
	products := protected.Group("/products")
	products.Use(middleware.CacheMiddleware(5*time.Minute, http.StatusOK))
	{
		products.POST("", s.CreateProduct, middleware.RequireRole("admin", "pharmacist"))
		products.GET("", s.ListProducts)
		products.GET("/search", s.SearchProducts)
		products.GET("/barcode/:barcode", s.SearchProductByBarcode)
		products.GET("/:id", s.GetProduct)
		products.PUT("/:id", s.UpdateProduct, middleware.RequireRole("admin", "pharmacist"))
		products.DELETE("/:id", s.DeleteProduct, middleware.RequireRole("admin"))
		products.GET("/:product_id/barcodes", s.GetBarcodesByProduct)
	}

	// Category routes
	categories := protected.Group("/categories")
	categories.Use(middleware.CacheMiddleware(10*time.Minute, http.StatusOK))
	{
		categories.POST("", s.CreateCategory, middleware.RequireRole("admin"))
		categories.GET("", s.ListCategories)
		categories.GET("/:id", s.GetCategory)
	}

	// Dosage Form routes
	dosageForms := protected.Group("/dosage_forms")
	dosageForms.Use(middleware.CacheMiddleware(10*time.Minute, http.StatusOK))
	{
		dosageForms.POST("", s.CreateDosageForm, middleware.RequireRole("admin"))
		dosageForms.GET("", s.ListDosageForms)
		dosageForms.GET("/:id", s.GetDosageForm)
	}

	// Order routes
	orders := protected.Group("/orders")
	{
		orders.POST("", s.CreateOrder)
		orders.GET("", s.ListOrders)
		orders.GET("/:id", s.GetOrder)
		orders.PUT("/:id/status", s.UpdateOrderStatus)
		orders.DELETE("/:id", s.DeleteOrder, middleware.RequireRole("admin"))
		orders.POST("/:order_id/items", s.CreateOrderItem)
		orders.GET("/:order_id/items", s.GetOrderItems)
	}

	// Order items routes
	orderItems := protected.Group("/order_items")
	{
		orderItems.PUT("/:id", s.UpdateOrderItem)
		orderItems.DELETE("/:id", s.DeleteOrderItem)
	}

	// Barcode routes
	barcodes := protected.Group("/barcodes")
	{
		barcodes.POST("", s.CreateBarcode, middleware.RequireRole("admin", "pharmacist"))
		barcodes.PUT("/:id", s.UpdateBarcode, middleware.RequireRole("admin", "pharmacist"))
		barcodes.DELETE("/:id", s.DeleteBarcode, middleware.RequireRole("admin"))
	}

	// User routes (admin only)
	users := protected.Group("/users")
	users.Use(middleware.RequireRole("admin"))
	{
		users.POST("", s.CreateUser)
		users.GET("", s.ListUsers)
		users.GET("/:id", s.GetUser)
		users.PUT("/:id", s.UpdateUser)
		users.DELETE("/:id", s.DeleteUser)
	}

	// Role routes (admin only)
	roles := protected.Group("/roles")
	roles.Use(middleware.RequireRole("admin"))
	{
		roles.POST("", s.CreateRole)
		roles.GET("", s.ListRoles)
		roles.GET("/:id", s.GetRole)
		roles.PUT("/:id", s.UpdateRole)
		roles.DELETE("/:id", s.DeleteRole)
	}
}

// Health check endpoint
func (s *Server) healthCheck(c echo.Context) error {
	// Check database connection
	err := s.db.Ping()
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":   "unhealthy",
			"service":  "DigiOrder API",
			"database": "disconnected",
			"error":    err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "healthy",
		"service":  "DigiOrder API",
		"database": "connected",
		"version":  "2.0.0",
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

	// Don't log 404 errors as they're expected
	if code != http.StatusNotFound {
		s.router.Logger.Error(err)
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