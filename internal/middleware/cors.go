// internal/middleware/cors.go - Secure CORS configuration
package middleware

import (
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns production-ready CORS settings
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: getEnvList("CORS_ALLOWED_ORIGINS", []string{
			"http://localhost:3000",
			"http://localhost:5173", // Vite default
		}),
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.PUT,
			echo.PATCH,
			echo.DELETE,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
			"X-Trace-ID",
			"X-Cache",
			"X-Cache-Age",
		},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}
}

// SecureCORSMiddleware creates CORS middleware with security checks
func SecureCORSMiddleware() echo.MiddlewareFunc {
	config := DefaultCORSConfig()

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     config.AllowMethods,
		AllowHeaders:     config.AllowHeaders,
		ExposeHeaders:    config.ExposeHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	})
}

// getEnvList retrieves a comma-separated list from environment
func getEnvList(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	// Split by comma and trim spaces
	origins := strings.Split(value, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return origins
}

// ValidateOrigin checks if an origin is allowed
func ValidateOrigin(origin string, allowedOrigins []string) bool {
	// Check exact matches
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}

		// Support wildcard subdomains
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}
