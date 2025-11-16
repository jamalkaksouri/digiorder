// internal/middleware/jwt.go - Complete JWT Implementation
package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	ErrMissingToken      = errors.New("missing authorization token")
	ErrInvalidToken      = errors.New("invalid token format")
	ErrExpiredToken      = errors.New("token has expired")
	ErrInvalidSignature  = errors.New("invalid token signature")
	ErrMissingClaims     = errors.New("missing required claims")
)

// JWTClaims represents the claims stored in JWT
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	RoleID   int32     `json:"role_id"`
	RoleName string    `json:"role_name"`
	jwt.RegisteredClaims
}

// GetJWTSecret retrieves JWT secret from environment
func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is not set")
	}
	return []byte(secret)
}

// GetJWTExpiry returns JWT expiration duration
func GetJWTExpiry() time.Duration {
	expiryStr := os.Getenv("JWT_EXPIRY")
	if expiryStr == "" {
		return 24 * time.Hour // Default 24 hours
	}
	
	duration, err := time.ParseDuration(expiryStr)
	if err != nil {
		return 24 * time.Hour
	}
	
	return duration
}

// GenerateToken creates a new JWT token
func GenerateToken(userID uuid.UUID, username string, roleID int32, roleName string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(GetJWTExpiry())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "digiorder-api",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTSecret())
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return GetJWTSecret(), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidSignature
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrMissingClaims
	}

	return claims, nil
}

// ExtractToken extracts JWT token from Authorization header
func ExtractToken(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrMissingToken
	}

	// Check for "Bearer " prefix
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}

// JWTMiddleware validates JWT tokens
func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token
			tokenString, err := ExtractToken(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{
					"error":   "unauthorized",
					"message": "Missing or invalid authorization token",
				})
			}

			// Validate token
			claims, err := ValidateToken(tokenString)
			if err != nil {
				var message string
				switch {
				case errors.Is(err, ErrExpiredToken):
					message = "Token has expired. Please login again."
				case errors.Is(err, ErrInvalidSignature):
					message = "Invalid token signature."
				default:
					message = "Invalid authentication token."
				}

				return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{
					"error":   "invalid_token",
					"message": message,
				})
			}

			// Store claims in context
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role_id", claims.RoleID)
			c.Set("role_name", claims.RoleName)
			c.Set("jwt_claims", claims)

			return next(c)
		}
	}
}

// RequireRole middleware ensures user has required role
func RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roleName, err := GetRoleNameFromContext(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{
					"error":   "unauthorized",
					"message": "Authentication required",
				})
			}

			// Check if user's role is in allowed roles
			for _, allowed := range allowedRoles {
				if roleName == allowed {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, map[string]string{
				"error":   "insufficient_permissions",
				"message": "You don't have permission to access this resource",
			})
		}
	}
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	return userID, nil
}

// GetUsernameFromContext retrieves username from context
func GetUsernameFromContext(c echo.Context) (string, error) {
	username, ok := c.Get("username").(string)
	if !ok {
		return "", errors.New("username not found in context")
	}
	return username, nil
}

// GetRoleIDFromContext retrieves role ID from context
func GetRoleIDFromContext(c echo.Context) (int32, error) {
	roleID, ok := c.Get("role_id").(int32)
	if !ok {
		return 0, errors.New("role ID not found in context")
	}
	return roleID, nil
}

// GetRoleNameFromContext retrieves role name from context
func GetRoleNameFromContext(c echo.Context) (string, error) {
	roleName, ok := c.Get("role_name").(string)
	if !ok {
		return "", errors.New("role name not found in context")
	}
	return roleName, nil
}

// GetJWTClaims retrieves full JWT claims from context
func GetJWTClaims(c echo.Context) (*JWTClaims, error) {
	claims, ok := c.Get("jwt_claims").(*JWTClaims)
	if !ok {
		return nil, errors.New("JWT claims not found in context")
	}
	return claims, nil
}