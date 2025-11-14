// internal/middleware/auth.go - FIXED VERSION
package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	RoleID   int32     `json:"role_id"`
	RoleName string    `json:"role_name"`
	jwt.RegisteredClaims
}

// GetJWTSecret retrieves the JWT secret from environment - FIXED
func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters")
	}
	return secret
}

// GetJWTExpiry retrieves the JWT expiry duration from environment
func GetJWTExpiry() time.Duration {
	expiry := os.Getenv("JWT_EXPIRY")
	if expiry == "" {
		expiry = "24h"
	}
	duration, err := time.ParseDuration(expiry)
	if err != nil {
		log.Printf("Invalid JWT_EXPIRY format, using default 24h: %v", err)
		return 24 * time.Hour
	}
	return duration
}

// GenerateToken generates a new JWT token for a user
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
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetJWTSecret()))
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(GetJWTSecret()), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Additional validation
		if time.Now().After(claims.ExpiresAt.Time) {
			return nil, jwt.ErrTokenExpired
		}
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// JWTMiddleware validates JWT tokens from requests
func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			// Extract token (format: "Bearer <token>")
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			tokenString := parts[1]
			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "empty token")
			}

			// Validate token
			claims, err := ValidateToken(tokenString)
			if err != nil {
				if err == jwt.ErrTokenExpired {
					return echo.NewHTTPError(http.StatusUnauthorized, "token has expired")
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			// Store claims in context
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role_id", claims.RoleID)
			c.Set("role_name", claims.RoleName)

			return next(c)
		}
	}
}

// RequireRole middleware checks if user has required role - FIXED
func RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Safe type assertion
			roleNameVal := c.Get("role_name")
			if roleNameVal == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
			}

			roleName, ok := roleNameVal.(string)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, "invalid role type in context")
			}

			// Check if user has allowed role
			for _, role := range allowedRoles {
				if roleName == role {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}

// GetUserIDFromContext retrieves user ID from context - FIXED
func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userIDVal := c.Get("user_id")
	if userIDVal == nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusInternalServerError, "invalid user ID type")
	}

	return userID, nil
}

// GetRoleNameFromContext retrieves role name from context - FIXED
func GetRoleNameFromContext(c echo.Context) (string, error) {
	roleNameVal := c.Get("role_name")
	if roleNameVal == nil {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	roleName, ok := roleNameVal.(string)
	if !ok {
		return "", echo.NewHTTPError(http.StatusInternalServerError, "invalid role name type")
	}

	return roleName, nil
}
