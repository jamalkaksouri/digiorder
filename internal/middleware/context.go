// internal/middleware/context.go - Context Helper Functions
package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SetUserContext sets user information in the Echo context
func SetUserContext(c echo.Context, userID uuid.UUID, username string, roleID int32, roleName string) {
	c.Set("user_id", userID)
	c.Set("username", username)
	c.Set("role_id", roleID)
	c.Set("role_name", roleName)
}

// ClearUserContext removes user information from context
func ClearUserContext(c echo.Context) {
	c.Set("user_id", nil)
	c.Set("username", nil)
	c.Set("role_id", nil)
	c.Set("role_name", nil)
}

// HasRole checks if user has a specific role
func HasRole(c echo.Context, roleName string) bool {
	userRole, err := GetRoleNameFromContext(c)
	if err != nil {
		return false
	}
	return userRole == roleName
}

// IsAuthenticated checks if user is authenticated
func IsAuthenticated(c echo.Context) bool {
	_, err := GetUserIDFromContext(c)
	return err == nil
}

// GetRequestID retrieves request ID from context
func GetRequestID(c echo.Context) string {
	if requestID, ok := c.Get("request_id").(string); ok {
		return requestID
	}
	return ""
}

// GetTraceID retrieves trace ID from context
func GetTraceID(c echo.Context) string {
	if traceID, ok := c.Get("trace_id").(string); ok {
		return traceID
	}
	return ""
}