# DigiOrder Security Implementation Guide

## Overview

This guide implements all critical security fixes for the DigiOrder application.

---

## 🔒 Security Issues Fixed

### 1. **Persistent Rate Limiting** ✅

- **Problem**: Rate limits stored in memory only, reset on restart
- **Solution**: Database-backed rate limiting with cleanup
- **Files**:
  - `internal/middleware/rate_limiter_db.go`
  - `internal/db/query/rate_limits.sql`

### 2. **Password Security** ✅

- **Problem**: Weak passwords (6 chars), hardcoded admin password
- **Solution**:
  - Minimum 12 characters
  - Uppercase, lowercase, digit, special char required
  - Password strength scoring
  - Bcrypt cost increased to 12
- **Files**: `internal/security/password.go`

### 3. **Environment File Protection** ✅

- **Problem**: Sensitive `.env` files in repository
- **Solution**: Updated `.gitignore` to exclude all environment files
- **Files**: `.gitignore`

### 4. **Login Rate Limiting** ✅

- **Problem**: No specific login attempt limiting
- **Solution**: 5 attempts per 5 minutes per IP
- **Files**: `internal/middleware/rate_limiter_db.go`

### 5. **CORS Configuration** ✅

- **Problem**: CORS not explicitly configured
- **Solution**: Whitelist-based CORS with environment variables
- **Files**: `internal/middleware/cors.go`

### 6. **N+1 Query Problem** ✅

- **Problem**: Role names fetched in loop when listing users
- **Solution**: JOIN queries to fetch users with roles in single query
- **Files**: `internal/db/query/users_optimized.sql`

### 7. **Structured Logging** ✅

- **Problem**: Basic logging without context
- **Solution**: JSON structured logging with request context, trace IDs
- **Files**: `internal/logging/logger.go`

### 8. **Secure Admin Setup** ✅

- **Problem**: Hardcoded admin password in migration
- **Solution**: One-time setup endpoint with token verification
- **Files**: `internal/server/setup.go`, `migrations/000004_secure_admin_setup.up.sql`

---

## 📋 Implementation Steps

### Step 1: Update Dependencies

```bash
go get golang.org/x/crypto@latest
go mod tidy
```

### Step 2: Add New Files

Copy all artifact files to your project:

- `internal/middleware/rate_limiter_db.go`
- `internal/middleware/cors.go`
- `internal/security/password.go`
- `internal/logging/logger.go`
- `internal/server/setup.go`
- `internal/db/query/rate_limits.sql`
- `internal/db/query/users_optimized.sql`

### Step 3: Update `.gitignore`

```bash
cp .gitignore .gitignore.backup
# Copy new .gitignore content
```

**⚠️ CRITICAL**: Remove any committed `.env` files from Git history:

```bash
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch .env .env.* " \
  --prune-empty --tag-name-filter cat -- --all
```

### Step 4: Run New Migration

```bash
# Create migration
migrate create -ext sql -dir migrations -seq secure_admin_setup

# Copy migration content from artifact

# Run migration
make migrate-up
```

### Step 5: Regenerate SQLC

```bash
make sqlc
```

### Step 6: Update Server Initialization

Replace in `internal/server/server.go`:

```go
import (
	"github.com/jamalkaksouri/DigiOrder/internal/logging"
	"github.com/jamalkaksouri/DigiOrder/internal/security"
)

type Server struct {
	db        *sql.DB
	queries   *db.Queries
	router    *echo.Echo
	validator *validator.Validate
	server    *http.Server
	logger    *logging.Logger // NEW
	rateLimiter *middleware.PersistentRateLimiter // NEW
}

func New(database *sql.DB) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	v := validator.New()
	registerCustomValidators(v)

	// Create structured logger
	logger := logging.NewLogger("digiorder", getEnv("ENV", "production"))

	// Create rate limiter with DB backing
	queries := db.New(database)
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
```

### Step 7: Update Routes

Replace in `internal/server/routes.go`:

```go
func (s *Server) registerRoutes() {
	// Add structured logging middleware FIRST
	s.router.Use(logging.LoggingMiddleware(s.logger))

	// Secure CORS
	s.router.Use(middleware.SecureCORSMiddleware())

	// Persistent rate limiting
	s.router.Use(middleware.PersistentRateLimitMiddleware(
		s.queries,
		middleware.DefaultRateLimitConfig(),
	))

	// ... rest of middleware ...

	// Setup endpoints (before auth)
	setup := s.router.Group("/api/v1/setup")
	{
		setup.GET("/status", s.GetSetupStatus)
		setup.POST("/initialize", s.InitialSetup)
	}

	// Auth routes
	auth := api.Group("/auth")
	{
		// Special rate limiting for login
		auth.POST("/login", s.Login,
			middleware.LoginRateLimitMiddleware(s.queries, 5, 5*time.Minute))
		auth.POST("/refresh", s.RefreshToken)
	}

	// ... rest of routes ...
}
```

### Step 8: Update User Creation

Replace in `internal/server/users.go`:

```go
import "github.com/jamalkaksouri/DigiOrder/internal/security"

func (s *Server) CreateUser(c echo.Context) error {
	// ... existing validation ...

	// Use new password validation
	if err := security.ValidatePassword(req.Password,
		security.DefaultPasswordRequirements()); err != nil {

		suggestions := security.SuggestPasswordImprovement(req.Password)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":       "weak_password",
			"details":     err.Error(),
			"suggestions": suggestions,
			"requirements": map[string]interface{}{
				"min_length": 12,
				"requires":   []string{"uppercase", "lowercase", "digit", "special char"},
			},
		})
	}

	// Hash with stronger cost
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError,
			"hash_error", "Failed to hash password.")
	}

	// ... rest of function ...
}
```

### Step 9: Update Login Handler

```go
func (s *Server) Login(c echo.Context) error {
	// Get logger from context
	logger := logging.GetLogger(c)

	// ... existing validation ...

	// Use secure password comparison
	err = security.ComparePassword(user.PasswordHash, req.Password)
	if err != nil {
		logger.Warn("Failed login attempt", map[string]interface{}{
			"username": req.Username,
			"ip":       c.RealIP(),
		})
		return RespondError(c, http.StatusUnauthorized,
			"invalid_credentials", "Invalid username or password.")
	}

	logger.Info("Successful login", map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
	})

	// ... rest of function ...
}
```

### Step 10: Environment Configuration

Update `.env.example`:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=CHANGE_ME
DB_NAME=digiorder_db
DB_SSLMODE=disable

# Server
SERVER_PORT=5582
SERVER_HOST=0.0.0.0
ENV=development

# Security
JWT_SECRET=GENERATE_RANDOM_64_CHARS_HERE
JWT_EXPIRY=24h

# Initial Setup (ONE-TIME USE - REMOVE AFTER SETUP)
INITIAL_SETUP_TOKEN=GENERATE_RANDOM_TOKEN_HERE

# CORS (comma-separated)
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

Create `.env` (never commit):

```bash
cp .env.example .env

# Generate strong JWT secret
openssl rand -base64 64 | tr -d '\n'
# Paste result as JWT_SECRET

# Generate setup token
openssl rand -hex 32 | tr -d '\n'
# Paste result as INITIAL_SETUP_TOKEN
```

---

## 🚀 First-Time Setup

### 1. Check Setup Status

```bash
curl http://localhost:5582/api/v1/setup/status
```

**Response**:

```json
{
  "data": {
    "setup_required": true,
    "admin_exists": false
  }
}
```

### 2. Initialize System

```bash
curl -X POST http://localhost:5582/api/v1/setup/initialize \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "SecureP@ssw0rd2024!",
    "confirm_password": "SecureP@ssw0rd2024!",
    "full_name": "System Administrator",
    "setup_token": "YOUR_SETUP_TOKEN_FROM_ENV"
  }'
```

**Success Response**:

```json
{
  "data": {
    "message": "System initialized successfully",
    "user": {
      "id": "00000000-0000-0000-0000-000000000001",
      "username": "admin",
      "full_name": "System Administrator"
    },
    "next_steps": [
      "1. Login with your credentials",
      "2. Create additional users",
      "3. Remove INITIAL_SETUP_TOKEN from environment"
    ]
  }
}
```

### 3. Remove Setup Token

After successful setup, remove from `.env`:

```bash
# Remove or comment out
# INITIAL_SETUP_TOKEN=...
```

Restart the application. The setup endpoint will now return 403 Forbidden.

---

## 📊 Rate Limiting Details

### Global Rate Limits

- **Public endpoints**: 100 requests/second, burst 200
- **Authenticated endpoints**: 1000 requests/minute
- **Login endpoint**: 5 attempts per 5 minutes per IP

### Database Tables

Rate limits are stored in `api_rate_limits`:

```sql
SELECT
    client_id,
    endpoint,
    SUM(requests_count) as total,
    MAX(window_start) as last_request
FROM api_rate_limits
WHERE window_start >= NOW() - INTERVAL '1 hour'
GROUP BY client_id, endpoint
ORDER BY total DESC
LIMIT 10;
```

### Monitoring

```bash
# Check rate limit stats
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:5582/api/v1/admin/rate-limits/stats
```

---

## 🧪 Testing

### Test Password Validation

```bash
# Weak password (should fail)
curl -X POST http://localhost:5582/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test",
    "password": "weak",
    "role_id": 3
  }'

# Response includes suggestions
{
  "error": "weak_password",
  "details": "password must be at least 12 characters long",
  "suggestions": [
    "Increase length to at least 12 characters",
    "Add uppercase letters",
    "Add special characters (!@#$%^&*)"
  ]
}
```

### Test Login Rate Limiting

```bash
# Rapid login attempts (6th should fail)
for i in {1..6}; do
  curl -X POST http://localhost:5582/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"wrong"}'
  echo "Attempt $i"
done
```

### Test CORS

```bash
curl -H "Origin: http://unauthorized.com" \
  -H "Access-Control-Request-Method: POST" \
  -X OPTIONS \
  http://localhost:5582/api/v1/products
# Should return CORS error
```

---

## 📈 Monitoring & Logging

### View Structured Logs

```bash
# JSON format
tail -f logs/digiorder.log | jq

# Example output:
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "service": "digiorder",
  "message": "HTTP Request",
  "request_id": "abc123",
  "user_id": "user-uuid",
  "method": "POST",
  "path": "/api/v1/products",
  "status_code": 201,
  "duration_ms": 45
}
```

### Check Rate Limit Records

```bash
psql -U postgres -d digiorder_db -c "
  SELECT * FROM api_rate_limits
  WHERE window_start >= NOW() - INTERVAL '1 hour'
  ORDER BY requests_count DESC
  LIMIT 10;
"
```

---

## 🔐 Production Checklist

- [ ] All `.env*` files removed from Git
- [ ] Strong JWT_SECRET generated (64+ chars)
- [ ] Initial admin created with 12+ char password
- [ ] INITIAL_SETUP_TOKEN removed from environment
- [ ] CORS_ALLOWED_ORIGINS configured for production domains
- [ ] LOG_LEVEL set to `info` or `warn`
- [ ] Database backups configured
- [ ] SSL/TLS enabled (DB_SSLMODE=require)
- [ ] Monitoring dashboards reviewed
- [ ] Rate limit thresholds adjusted for load

---

## 🐛 Troubleshooting

### Issue: "Setup already complete"

**Solution**: This is correct behavior. Setup endpoint is one-time only.

### Issue: "Invalid setup token"

**Solution**: Check `INITIAL_SETUP_TOKEN` in `.env` matches request.

### Issue: Rate limit DB errors

**Solution**:

```bash
# Check table exists
psql -d digiorder_db -c "\d api_rate_limits"

# Regenerate SQLC if needed
make sqlc
```

### Issue: Password validation failing

**Solution**: Ensure password meets all requirements:

- 12+ characters
- 1 uppercase, 1 lowercase
- 1 digit
- 1 special character

---

## 📚 Additional Resources

- **OWASP Password Guidelines**: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- **Rate Limiting Best Practices**: https://www.rfc-editor.org/rfc/rfc6585.html
- **Structured Logging**: https://www.honeycomb.io/blog/structured-logging-and-your-team

---

## ✅ Summary

All critical security issues have been addressed:

1. ✅ **Persistent Rate Limiting** - DB-backed with proper cleanup
2. ✅ **Strong Passwords** - 12+ chars with complexity requirements
3. ✅ **Secure Secrets** - No hardcoded passwords, gitignore updated
4. ✅ **Login Protection** - 5 attempts per 5 minutes
5. ✅ **CORS Security** - Whitelist-based configuration
6. ✅ **Query Optimization** - No N+1 problems
7. ✅ **Structured Logging** - Full request context tracking
8. ✅ **Secure Setup** - One-time admin creation with token

Your application is now production-ready with enterprise-grade security! 🎉
