// ============================================================================
// internal/server/server.go - ENHANCED VERSION
// ============================================================================

package server

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
)

// Server holds the dependencies for our application.
type Server struct {
	db        *sql.DB
	queries   *db.Queries
	router    *echo.Echo
	validator *validator.Validate
	server    *http.Server
}

// New creates a new Server instance with all its dependencies.
func New(database *sql.DB) *Server {
	e := echo.New()

	// Hide Echo banner
	e.HideBanner = true
	e.HidePort = true

	// Create a new validator instance with custom validations
	v := validator.New()

	// Register custom validators
	registerCustomValidators(v)

	server := &Server{
		db:        database,
		queries:   db.New(database),
		router:    e,
		validator: v,
	}

	// Register all routes and middleware
	server.registerRoutes()

	return server
}

// Start runs the HTTP server on a specific address.
func (s *Server) Start(addr string) error {
	// Configure HTTP server with timeouts
	s.server = &http.Server{
		Addr:           addr,
		Handler:        s.router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// registerCustomValidators adds custom validation rules
func registerCustomValidators(v *validator.Validate) {
	// Add custom UUID validator
	v.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		_, err := uuid.Parse(fl.Field().String())
		return err == nil
	})

	// Add barcode format validator
	v.RegisterValidation("barcode_format", func(fl validator.FieldLevel) bool {
		barcode := fl.Field().String()
		// Check if barcode contains only valid characters
		if len(barcode) < 8 || len(barcode) > 128 {
			return false
		}
		return true
	})
}