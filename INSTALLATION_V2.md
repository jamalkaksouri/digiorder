# DigiOrder v2.0 - Installation Guide

## Quick Installation (5 Minutes)

```bash
# 1. Update dependencies
go mod download

# 2. Run new migrations
make migrate-up

# 3. Regenerate SQLC code
make sqlc

# 4. Build application
make build

# 5. Run application
make run
```

That's it! Your v2.0 application is running on http://localhost:5582

---

## Detailed Installation Steps

### Step 1: Prepare Environment

```bash
# Ensure you have Go 1.22+
go version

# Install required tools
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Step 2: Update Go Dependencies

```bash
# Download new dependencies (JWT, rate limiter, etc.)
go mod download

# Verify dependencies
go mod verify
```

### Step 3: Database Setup

#### Option A: Existing Database

```bash
# Run new migrations
make migrate-up

# Verify migrations
migrate -path migrations -database "postgresql://postgres:root@localhost:5432/digiorder_db?sslmode=disable" version
```

#### Option B: Fresh Install

```bash
# Start PostgreSQL (if not running)
make docker-up

# Wait for database to be ready
sleep 5

# Run all migrations
make migrate-up
```

### Step 4: Generate Database Code

```bash
# Generate SQLC code with new queries
make sqlc

# Verify generated files
ls internal/db/*.go
```

You should see new files:

- `internal/db/barcodes.sql.go` (NEW)
- Updated `categories.sql.go`, `products.sql.go`, etc.

### Step 5: Update Configuration

```bash
# Update .env file with JWT secret
echo "JWT_SECRET=$(openssl rand -base64 64)" >> .env
echo "JWT_EXPIRY=24h" >> .env
```

Your `.env` should now include:

```bash
JWT_SECRET=<generated_random_string>
JWT_EXPIRY=24h
```

### Step 6: Build Application

```bash
# Clean build
make clean
make build

# Verify binary
./bin/digiorder --help
```

### Step 7: Create Initial Admin User

Before running the application, you need an admin user:

```sql
-- Connect to database
psql -U postgres -d digiorder_db

-- Create admin user (password: admin123456)
INSERT INTO users (username, full_name, password_hash, role_id)
VALUES (
  'admin',
  'System Administrator',
  '$2a$10$Zu7yVNJ0e9Fn9vwUy9vRbO5CqPQZMB8l5k8hEWnGvhkrFUKqj9iEW',
  1
);

-- Verify
SELECT username, full_name, role_id FROM users;
\q
```

Or use this one-liner:

```bash
docker-compose exec -T postgres psql -U postgres -d digiorder_db <<EOF
INSERT INTO users (username, full_name, password_hash, role_id)
VALUES ('admin', 'System Administrator', '\$2a\$10\$Zu7yVNJ0e9Fn9vwUy9vRbO5CqPQZMB8l5k8hEWnGvhkrFUKqj9iEW', 1)
ON CONFLICT (username) DO NOTHING;
EOF
```

### Step 8: Start Application

```bash
# Run in foreground
make run

# Or run in background
nohup ./bin/digiorder > digiorder.log 2>&1 &
```

### Step 9: Verify Installation

```bash
# Check health
curl http://localhost:5582/health

# Should return:
# {"status":"healthy","service":"DigiOrder API","database":"connected","version":"2.0.0"}

# Check metrics endpoint
curl http://localhost:5582/metrics
```

### Step 10: Test Authentication

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:5582/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }' | jq -r '.data.token')

echo "Token: ${TOKEN:0:50}..."

# Get profile
curl http://localhost:5582/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN" | jq

# List products (protected endpoint)
curl http://localhost:5582/api/v1/products \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## File Checklist

Make sure you have all these new/updated files:

### New Files (Must Add)

- [ ] `internal/middleware/auth.go`
- [ ] `internal/middleware/rate_limiter.go`
- [ ] `internal/middleware/cache.go`
- [ ] `internal/middleware/logging.go`
- [ ] `internal/server/auth.go`
- [ ] `internal/server/barcodes.go`
- [ ] `internal/server/products_test.go`
- [ ] `internal/db/query/barcodes.sql`
- [ ] `migrations/000002_add_features.up.sql`
- [ ] `migrations/000002_add_features.down.sql`
- [ ] `Dockerfile`
- [ ] `docker-compose.prod.yml`
- [ ] `.env.production`
- [ ] `.github/workflows/ci.yml`

### Updated Files

- [ ] `internal/server/routes.go` (with auth middleware)
- [ ] `internal/server/server.go` (if needed)
- [ ] `go.mod` (new dependencies)
- [ ] `README.md` (v2.0 features)

---

## Troubleshooting

### Error: "invalid or expired token"

**Problem**: JWT token validation failing

**Solution**:

```bash
# Ensure JWT_SECRET is set in .env
grep JWT_SECRET .env

# If not set, add it
echo "JWT_SECRET=$(openssl rand -base64 64)" >> .env

# Restart application
```

### Error: "migration failed"

**Problem**: Migrations not running

**Solution**:

```bash
# Check current version
migrate -path migrations -database "postgresql://postgres:root@localhost:5432/digiorder_db?sslmode=disable" version

# Force version if needed
migrate -path migrations -database "postgresql://postgres:root@localhost:5432/digiorder_db?sslmode=disable" force 1

# Re-run migrations
make migrate-up
```

### Error: "sqlc: not found"

**Problem**: SQLC not installed

**Solution**:

```bash
# Install SQLC
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Verify installation
which sqlc

# Add to PATH if needed
export PATH=$PATH:$(go env GOPATH)/bin
```

### Error: "rate limit exceeded"

**Problem**: Too many requests

**Solution**:

```bash
# Wait a few seconds and retry
sleep 5

# Or increase rate limits in code
# Edit internal/server/routes.go
# Change: middleware.RateLimitMiddleware(100, 200)
# To: middleware.RateLimitMiddleware(1000, 2000)
```

### Error: "insufficient permissions"

**Problem**: User role doesn't have access

**Solution**:

```bash
# Check your role
curl http://localhost:5582/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN" | jq '.data.role_name'

# If not admin, login as admin or request admin to grant permissions
```

---

## Migration From v1.0 to v2.0

### 1. Backup Database

```bash
# Create backup
pg_dump -U postgres -d digiorder_db > digiorder_v1_backup.sql
```

### 2. Update Code

```bash
git pull origin main
# or
# Download new files manually
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run New Migrations

```bash
# This adds new tables and columns
make migrate-up
```

### 5. Regenerate SQLC

```bash
make sqlc
```

### 6. Update Configuration

```bash
# Add JWT configuration
cat >> .env <<EOF
JWT_SECRET=$(openssl rand -base64 64)
JWT_EXPIRY=24h
EOF
```

### 7. Rebuild & Deploy

```bash
make build
./bin/digiorder
```

### 8. Create Admin User

```bash
# Use the SQL script from Step 7 above
```

### 9. Test Everything

```bash
# Run test suite
make test

# Manual testing
./test_authentication.sh
```

---

## Docker Installation

### Quick Docker Setup

```bash
# Production deployment
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

### Docker Environment

Create `.env` file:

```bash
DB_USER=digiorder
DB_PASSWORD=secure_password_here
DB_NAME=digiorder_production
JWT_SECRET=your_secret_64_chars_minimum
JWT_EXPIRY=24h
```

Then:

```bash
docker-compose -f docker-compose.prod.yml up -d
```

---

## Testing Your Installation

### Automated Test Script

Save as `test_installation.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:5582"
PASS=0
FAIL=0

echo "=== DigiOrder v2.0 Installation Test ==="

# Test 1: Health Check
echo -n "1. Testing health endpoint... "
HEALTH=$(curl -s $BASE_URL/health | jq -r '.status')
if [ "$HEALTH" = "healthy" ]; then
    echo "✅ PASS"
    ((PASS++))
else
    echo "❌ FAIL"
    ((FAIL++))
fi

# Test 2: Metrics Endpoint
echo -n "2. Testing metrics endpoint... "
METRICS=$(curl -s $BASE_URL/metrics | jq -r '.total_requests')
if [ ! -z "$METRICS" ]; then
    echo "✅ PASS"
    ((PASS++))
else
    echo "❌ FAIL"
    ((FAIL++))
fi

# Test 3: Authentication
echo -n "3. Testing authentication... "
TOKEN=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123456"}' | jq -r '.data.token')
if [ ! -z "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    echo "✅ PASS"
    ((PASS++))
else
    echo "❌ FAIL"
    ((FAIL++))
fi

# Test 4: Protected Endpoint
echo -n "4. Testing protected endpoint... "
PROFILE=$(curl -s $BASE_URL/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN" | jq -r '.data.username')
if [ "$PROFILE" = "admin" ]; then
    echo "✅ PASS"
    ((PASS++))
else
    echo "❌ FAIL"
    ((FAIL++))
fi

# Test 5: Rate Limiting
echo -n "5. Testing rate limiting... "
for i in {1..250}; do
    curl -s $BASE_URL/health > /dev/null
done
RATE_LIMITED=$(curl -s $BASE_URL/health | jq -r '.error')
if [ "$RATE_LIMITED" = "rate limit exceeded" ]; then
    echo "✅ PASS"
    ((PASS++))
else
    echo "⚠️  SKIP (may need more requests)"
fi

# Summary
echo ""
echo "=== Test Summary ==="
echo "Passed: $PASS"
echo "Failed: $FAIL"
echo ""

if [ $FAIL -eq 0 ]; then
    echo "✅ All tests passed! Installation successful."
    exit 0
else
    echo "❌ Some tests failed. Please check the errors above."
    exit 1
fi
```

Run it:

```bash
chmod +x test_installation.sh
./test_installation.sh
```

---

## Next Steps

After successful installation:

1. **Read Documentation**

   - `AUTHENTICATION_GUIDE.md` - Learn about authentication
   - `DEPLOYMENT_GUIDE.md` - Deploy to production
   - `API_TESTING_GUIDE.md` - Test all endpoints

2. **Create Additional Users**

   ```bash
   curl -X POST http://localhost:5582/api/v1/users \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{ ... }'
   ```

3. **Configure Production Settings**

   - Set strong JWT_SECRET
   - Configure CORS
   - Set up SSL/TLS
   - Configure rate limits

4. **Set Up Monitoring**

   - Check `/metrics` regularly
   - Set up automated health checks
   - Configure log aggregation

5. **Enable CI/CD**
   - Set up GitHub Actions
   - Configure secrets
   - Test deployment pipeline

---

## Support

If you encounter any issues:

1. **Check Logs**

   ```bash
   tail -f digiorder.log
   # or
   docker-compose logs -f api
   ```

2. **Verify Database**

   ```bash
   psql -U postgres -d digiorder_db -c "SELECT COUNT(*) FROM users;"
   ```

3. **Test Connectivity**

   ```bash
   curl -v http://localhost:5582/health
   ```

4. **Open Issue**
   - https://github.com/jamalkaksouri/DigiOrder/issues

---

## Congratulations! 🎉

Your DigiOrder v2.0 installation is complete. You now have a production-ready, enterprise-grade pharmacy order management system with:

- ✅ JWT Authentication
- ✅ Role-Based Authorization
- ✅ Rate Limiting
- ✅ Performance Caching
- ✅ Comprehensive Monitoring
- ✅ Barcode Support
- ✅ Automated Testing
- ✅ Docker Deployment
- ✅ CI/CD Pipeline

Happy coding! 🚀
