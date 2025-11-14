// internal/server/server.go - ENHANCED ERROR HANDLING VERSION
package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/jamalkaksouri/DigiOrder/internal/logging"
	"github.com/jamalkaksouri/DigiOrder/internal/middleware"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

// Server holds the dependencies for our application.
type Server struct {
	db          *sql.DB
	queries     *db.Queries
	router      *echo.Echo
	validator   *validator.Validate
	server      *http.Server
	logger      *logging.Logger
	rateLimiter *middleware.PersistentRateLimiter
}

// New creates a new Server instance with all its dependencies.
func New(database *sql.DB) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	v := validator.New()
	registerCustomValidators(v)

	queries := db.New(database)
	logger := logging.NewLogger("digiorder", getEnv("ENV", "production"))
	rateLimiter := middleware.NewPersistentRateLimiter(queries,
		middleware.DefaultRateLimitConfig())

	server := &Server{
		db:          database,
		queries:     queries,
		router:      e,
		validator:   v,
		logger:      logger,
		rateLimiter: rateLimiter,
	}

	server.registerRoutes()
	return server
}

// Start runs the HTTP server on a specific address.
func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:           addr,
		Handler:        s.router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// registerCustomValidators adds custom validation rules
func registerCustomValidators(v *validator.Validate) {
	v.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		_, err := uuid.Parse(fl.Field().String())
		return err == nil
	})

	v.RegisterValidation("barcode_format", func(fl validator.FieldLevel) bool {
		barcode := fl.Field().String()
		if len(barcode) < 8 || len(barcode) > 128 {
			return false
		}
		return true
	})
}

// HandleDatabaseError converts database errors to meaningful HTTP responses
func HandleDatabaseError(c echo.Context, err error, entityName string) error {
	if err == nil {
		return nil
	}

	// Handle no rows found
	if err == sql.ErrNoRows {
		return RespondError(c, http.StatusNotFound, "not_found",
			fmt.Sprintf("%s not found.", entityName))
	}

	// Handle PostgreSQL specific errors
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23505": // unique_violation
			// Extract the constraint name to provide better error message
			if strings.Contains(pqErr.Message, "username") {
				return RespondError(c, http.StatusConflict, "duplicate_username",
					"A user with this username already exists.")
			}
			if strings.Contains(pqErr.Message, "barcode") {
				return RespondError(c, http.StatusConflict, "duplicate_barcode",
					"This barcode is already registered to another product.")
			}
			return RespondError(c, http.StatusConflict, "duplicate_entry",
				"This entry already exists in the database.")

		case "23503": // foreign_key_violation
			if strings.Contains(pqErr.Message, "role_id") {
				return RespondError(c, http.StatusBadRequest, "invalid_role",
					"The specified role does not exist.")
			}
			if strings.Contains(pqErr.Message, "product_id") {
				return RespondError(c, http.StatusBadRequest, "invalid_product",
					"The specified product does not exist.")
			}
			if strings.Contains(pqErr.Message, "category_id") {
				return RespondError(c, http.StatusBadRequest, "invalid_category",
					"The specified category does not exist.")
			}
			if strings.Contains(pqErr.Message, "dosage_form_id") {
				return RespondError(c, http.StatusBadRequest, "invalid_dosage_form",
					"The specified dosage form does not exist.")
			}
			return RespondError(c, http.StatusBadRequest, "foreign_key_violation",
				"Referenced entity does not exist.")

		case "23502": // not_null_violation
			field := pqErr.Column
			return RespondError(c, http.StatusBadRequest, "missing_required_field",
				fmt.Sprintf("Field '%s' is required and cannot be null.", field))

		case "23514": // check_violation
			return RespondError(c, http.StatusBadRequest, "constraint_violation",
				"Data violates database constraint.")

		case "22P02": // invalid_text_representation (bad UUID format)
			return RespondError(c, http.StatusBadRequest, "invalid_format",
				"Invalid data format provided.")

		case "42703": // undefined_column
			return RespondError(c, http.StatusInternalServerError, "database_error",
				"Database schema error. Please contact support.")

		default:
			// Log the actual error for debugging
			if logger := logging.GetLogger(c); logger != nil {
				logger.Error("Database error", fmt.Errorf("pq error: %s - %s", pqErr.Code, pqErr.Message), map[string]any{
					"code":    pqErr.Code,
					"message": pqErr.Message,
					"detail":  pqErr.Detail,
				})
			}
			return RespondError(c, http.StatusInternalServerError, "database_error",
				"A database error occurred. Please try again later.")
		}
	}

	// Handle connection errors
	if strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "connection reset") {
		return RespondError(c, http.StatusServiceUnavailable, "database_unavailable",
			"Database is temporarily unavailable. Please try again later.")
	}

	// Handle timeout errors
	if strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline exceeded") {
		return RespondError(c, http.StatusGatewayTimeout, "database_timeout",
			"Database operation timed out. Please try again.")
	}

	// Log unknown database errors
	if logger := logging.GetLogger(c); logger != nil {
		logger.Error("Unknown database error", err, map[string]any{
			"error_type": fmt.Sprintf("%T", err),
			"error_msg":  err.Error(),
		})
	}

	// Generic database error
	return RespondError(c, http.StatusInternalServerError, "database_error",
		"An unexpected database error occurred. Please try again later.")
}

// ValidateRequest validates request body and returns user-friendly errors
func (s *Server) ValidateRequest(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is malformed or invalid JSON.")
	}

	if err := s.validator.Struct(req); err != nil {
		// Parse validation errors to return user-friendly messages
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, fieldError := range validationErrors {
				errorMessages = append(errorMessages, formatValidationError(fieldError))
			}
			return RespondError(c, http.StatusBadRequest, "validation_error",
				strings.Join(errorMessages, "; "))
		}
		return RespondError(c, http.StatusBadRequest, "validation_error",
			"Request validation failed.")
	}

	return nil
}

// formatValidationError converts validator.FieldError to user-friendly message
func formatValidationError(fe validator.FieldError) string {
	field := fe.Field()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required", field)
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters", field, fe.Param())
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email address", field)
	case "uuid":
		return fmt.Sprintf("Field '%s' must be a valid UUID", field)
	case "gt":
		return fmt.Sprintf("Field '%s' must be greater than %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("Field '%s' must be greater than or equal to %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("Field '%s' must be less than %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("Field '%s' must be less than or equal to %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", field, fe.Param())
	default:
		return fmt.Sprintf("Field '%s' failed validation: %s", field, fe.Tag())
	}
}

// ParseUUID safely parses UUID from string parameter
func ParseUUID(c echo.Context, paramName string) (uuid.UUID, error) {
	idStr := c.Param(paramName)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, RespondError(c, http.StatusBadRequest, "invalid_id",
			fmt.Sprintf("The provided %s is not a valid UUID.", paramName))
	}
	return id, nil
}

// ParseInt safely parses integer from string parameter
func ParseInt(c echo.Context, paramName string) (int32, error) {
	idStr := c.Param(paramName)
	var id int64
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		return 0, RespondError(c, http.StatusBadRequest, "invalid_id",
			fmt.Sprintf("The provided %s is not a valid number.", paramName))
	}
	return int32(id), nil
}

// GetUserLoginHistory - Admin endpoint to view user's login history
func (s *Server) GetUserLoginHistory(c echo.Context) error {
	ctx := c.Request().Context()
	username := c.Param("username")

	if username == "" {
		return RespondError(c, http.StatusBadRequest, "missing_username",
			"Username parameter is required.")
	}

	limit := 50
	offset := 0

	history, err := s.queries.GetUserLoginHistory(ctx, db.GetUserLoginHistoryParams{
		Username: username,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})

	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve login history.")
	}

	if history == nil {
		history = []db.GetUserLoginHistoryRow{}
	}

	return RespondSuccess(c, http.StatusOK, history)
}
