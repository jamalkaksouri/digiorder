package server

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// Common error types
var (
	ErrInvalidRequest     = NewAPIError(http.StatusBadRequest, "invalid_request", "The request body is malformed or invalid")
	ErrValidationFailed   = NewAPIError(http.StatusBadRequest, "validation_error", "Request validation failed")
	ErrUnauthorized       = NewAPIError(http.StatusUnauthorized, "unauthorized", "Authentication required")
	ErrForbidden          = NewAPIError(http.StatusForbidden, "forbidden", "Insufficient permissions")
	ErrNotFound           = NewAPIError(http.StatusNotFound, "not_found", "Resource not found")
	ErrConflict           = NewAPIError(http.StatusConflict, "conflict", "Resource already exists")
	ErrRateLimited        = NewAPIError(http.StatusTooManyRequests, "rate_limited", "Too many requests")
	ErrDatabaseTimeout    = NewAPIError(http.StatusGatewayTimeout, "database_timeout", "Database operation timed out")
	ErrServiceUnavailable = NewAPIError(http.StatusServiceUnavailable, "service_unavailable", "Service temporarily unavailable")
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func NewAPIError(statusCode int, code, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) WithDetails(details string) *APIError {
	return &APIError{
		StatusCode: e.StatusCode,
		Code:       e.Code,
		Message:    details,
	}
}

func (e *APIError) Send(c echo.Context) error {
	return RespondError(c, e.StatusCode, e.Code, e.Message)
}

// Helper function for easy error returns
func SendError(c echo.Context, err error) error {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Send(c)
	}
	// Unknown error - this should be logged
	return RespondError(c, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
}
