package middleware

import (
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

// GetJWTSecret retrieves the JWT secret from environment
func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your_secret_key_here" // Fallback for development
	}
	return secret
}

// GetJWTExpiry retrieves the JWT expiry duration from environment
func GetJWTExpiry() time.Duration {
	expiry := os.Getenv("JWT_EXPIRY")
	if expiry == "" {
		expiry = "24h" // Default 24 hours
	}
	duration, err := time.ParseDuration(expiry)
	if err != nil {
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
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetJWTSecret()))
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetJWTSecret()), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
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

			// Validate token
			claims, err := ValidateToken(tokenString)
			if err != nil {
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

// RequireRole middleware checks if user has required role
func RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roleName := c.Get("role_name").(string)

			for _, role := range allowedRoles {
				if roleName == role {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}
	return userID, nil
}

// GetRoleNameFromContext retrieves role name from context
func GetRoleNameFromContext(c echo.Context) (string, error) {
	roleName, ok := c.Get("role_name").(string)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}
	return roleName, nil
}