package server

import (
	"database/sql"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// Server holds the dependencies for our application.
type Server struct {
	db        *sql.DB
	queries   *db.Queries
	router    *echo.Echo
	validator *validator.Validate
}

// New creates a new Server instance with all its dependencies.
func New(database *sql.DB) *Server {
	e := echo.New()
	
	// Create a new validator instance
	v := validator.New()

	server := &Server{
		db:        database,
		queries:   db.New(database),
		router:    e,
		validator: v,
	}

	// Register all routes
	server.registerRoutes()

	return server
}

// Start runs the HTTP server on a specific address.
func (s *Server) Start(addr string) error {
	return s.router.Start(addr)
}