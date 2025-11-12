# DigiOrder v3.0 - Enterprise Pharmacy Order Management System

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-A+-green.svg)](#security-features)
[![Production Ready](https://img.shields.io/badge/Production-Ready-brightgreen.svg)](#production-deployment)

A secure, high-performance order management system for pharmacies built with Go, PostgreSQL, and modern security practices.

---

## 🚀 Features

### Core Functionality

- ✅ **Complete Product Management** - CRUD operations with barcode support
- ✅ **Order Processing** - Draft, submitted, processing, completed workflow
- ✅ **User Management** - Role-based access control (Admin, Pharmacist, Clerk)
- ✅ **Barcode Support** - EAN-13, UPC-A, Code128 scanning
- ✅ **Multi-language** - Persian/English support

### Security Features

- 🔐 **JWT Authentication** - Secure token-based authentication
- 🛡️ **Strong Password Policy** - 12+ characters with complexity requirements
- 🚦 **Rate Limiting** - Multi-layer protection (in-memory + database-backed)
- 🔒 **Protected Admin Account** - Primary admin cannot be deleted
- 📝 **Audit Logging** - Complete activity tracking with IP and user agent
- 🎯 **Permission System** - Granular resource-action based permissions
- 🌐 **CORS Security** - Configurable origin whitelist

### Performance Features

- ⚡ **Response Caching** - 5-minute TTL for GET requests
- 📊 **Query Optimization** - No N+1 queries, JOIN-based fetching
- 🔄 **Connection Pooling** - Optimized database connections
- 💾 **Soft Deletes** - Recoverable data deletion
- 🎯 **Efficient Indexing** - Optimized database indexes

### Observability

- 📈 **Prometheus Metrics** - Request rates, latencies, error rates
- 📊 **Grafana Dashboards** - System and business metrics visualization
- 🔔 **Alertmanager** - Automated alerting for critical issues
- 📝 **Structured Logging** - JSON logs with request context and trace IDs
- 🔍 **Distributed Tracing** - Request tracking across services

---

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Security](#security-features)
- [API Documentation](#api-endpoints)
- [Configuration](#configuration)
- [Development](#development)
- [Production Deployment](#production-deployment)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

---

## 🎯 Quick Start

### Prerequisites

- Go 1.22 or higher
- PostgreSQL 15 or higher
- Docker & Docker Compose (optional)
- Make

### 5-Minute Setup

```bash
# 1. Clone repository
git clone https://github.com/jamalkaksouri/DigiOrder.git
cd DigiOrder

# 2. Start services with monitoring
docker-compose -f docker-compose.monitoring.yml up -d

# 3. Wait for services to be ready (30 seconds)
sleep 30

# 4. Initialize system (first-time setup)
curl -X POST http://localhost:5582/api/v1/setup/initialize \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "SecureP@ssw0rd2024!",
    "confirm_password": "SecureP@ssw0rd2024!",
    "full_name": "System Administrator",
    "setup_token": "YOUR_SETUP_TOKEN"
  }'

# 5. Login
curl -X POST http://localhost:5582/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "SecureP@ssw0rd2024!"
  }'
```

**Access Points**:

- API: http://localhost:5582
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Alertmanager: http://localhost:9093

---

## 💻 Installation

### Local Development Setup

#### 1. Install Dependencies

```bash
go mod download
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

#### 2. Configure Environment

```bash
cp .env.example .env

# Edit .env with your settings
nano .env
```

**Required Environment Variables**:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=digiorder_db
DB_SSLMODE=disable

# Server
SERVER_PORT=5582
SERVER_HOST=0.0.0.0
ENV=development

# Security
JWT_SECRET=<generate_with_openssl_rand_-base64_64>
JWT_EXPIRY=24h
INITIAL_SETUP_TOKEN=<generate_with_openssl_rand_-hex_32>

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

#### 3. Start Database

```bash
# Option A: Using Docker
make docker-up

# Option B: Using existing PostgreSQL
# Ensure PostgreSQL is running and accessible
```

#### 4. Run Migrations

```bash
make migrate-up
```

#### 5. Generate SQLC Code

```bash
make sqlc
```

#### 6. Build and Run

```bash
# Build
make build

# Run
make run

# Or combine
make build && make run
```

---

## 🔒 Security Features

### Password Security

- **Minimum Length**: 12 characters
- **Complexity Required**:
  - At least 1 uppercase letter
  - At least 1 lowercase letter
  - At least 1 digit
  - At least 1 special character
- **Hashing**: Bcrypt with cost factor 12
- **Common Password Detection**: Blocks easily guessable passwords

### Rate Limiting

- **Global**: 100 requests/second (burst: 200)
- **Authenticated Users**: 1000 requests/minute
- **Login Attempts**: 5 attempts per 5 minutes per IP
- **Storage**: Database-backed with automatic cleanup

### Admin Protection

- **Primary Admin**: UUID `00000000-0000-0000-0000-000000000001` cannot be deleted
- **Last Admin**: System prevents deletion of last admin user
- **User Creation**: Only admins can create new users

### Audit Logging

Every action is logged with:

- User ID and username
- Action type (create, update, delete)
- Entity type and ID
- Old and new values (JSON)
- IP address and User Agent
- Timestamp

### CORS Security

- Whitelist-based origin validation
- Configurable via environment variables
- Wildcard subdomain support

---

## 📚 API Endpoints

### Authentication

#### Login

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "SecureP@ssw0rd2024!"
}

Response: 200 OK
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": "24h",
    "user": {
      "id": "...",
      "username": "admin",
      "role_name": "admin"
    }
  }
}
```

#### Refresh Token

```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "token": "current_token_here"
}
```

#### Get Profile

```bash
GET /api/v1/auth/profile
Authorization: Bearer <token>
```

#### Change Password

```bash
PUT /api/v1/auth/password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "current_password",
  "new_password": "NewSecureP@ssw0rd2024!"
}
```

### Products

```bash
# Create Product (Admin/Pharmacist)
POST /api/v1/products
Authorization: Bearer <token>
{
  "name": "Product Name",
  "brand": "Brand Name",
  "dosage_form_id": 1,
  "strength": "500mg",
  "unit": "tablet",
  "category_id": 1,
  "description": "Description"
}

# List Products (All authenticated users)
GET /api/v1/products?limit=50&offset=0

# Search Products
GET /api/v1/products/search?q=aspirin

# Get Product by Barcode
GET /api/v1/products/barcode/5901234123457

# Update Product (Admin/Pharmacist)
PUT /api/v1/products/:id

# Delete Product (Admin only)
DELETE /api/v1/products/:id
```

### Barcodes

```bash
# Add Barcode to Product
POST /api/v1/barcodes
{
  "product_id": "uuid",
  "barcode": "5901234123457",
  "barcode_type": "EAN-13"
}

# List Product Barcodes
GET /api/v1/products/:product_id/barcodes

# Update Barcode
PUT /api/v1/barcodes/:id

# Delete Barcode
DELETE /api/v1/barcodes/:id
```

### Orders

```bash
# Create Order
POST /api/v1/orders
{
  "status": "draft",
  "notes": "Weekly order"
}

# Add Item to Order
POST /api/v1/orders/:order_id/items
{
  "product_id": "uuid",
  "requested_qty": 10,
  "unit": "boxes"
}

# List Orders
GET /api/v1/orders?limit=50&offset=0

# Update Order Status
PUT /api/v1/orders/:id/status
{
  "status": "submitted"
}
```

### Users (Admin Only)

```bash
# Create User
POST /api/v1/users
{
  "username": "pharmacist1",
  "full_name": "John Doe",
  "password": "SecureP@ssw0rd123!",
  "role_id": 2
}

# List Users
GET /api/v1/users?limit=50&offset=0

# Update User
PUT /api/v1/users/:id

# Delete User (with protection)
DELETE /api/v1/users/:id
```

### Permissions (Admin Only)

```bash
# Create Permission
POST /api/v1/permissions
{
  "name": "export_reports",
  "resource": "reports",
  "action": "export",
  "description": "Export system reports"
}

# Assign Permission to Role
POST /api/v1/roles/:role_id/permissions
{
  "permission_id": 5
}

# Check User Permission
GET /api/v1/auth/check-permission?resource=products&action=create
```

### Audit Logs (Admin Only)

```bash
# List Audit Logs
GET /api/v1/audit-logs?limit=50&offset=0

# Get Entity History
GET /api/v1/audit-logs/entity/product/:product_id

# Get User Activity
GET /api/v1/users/:user_id/activity

# Get Audit Statistics
GET /api/v1/audit-logs/stats
```

### Monitoring

```bash
# Health Check (Public)
GET /health

# Prometheus Metrics (Public)
GET /metrics
```

---

## ⚙️ Configuration

### Database Configuration

```env
DB_HOST=localhost              # Database host
DB_PORT=5432                   # Database port
DB_USER=postgres               # Database user
DB_PASSWORD=secure_password    # Database password
DB_NAME=digiorder_db          # Database name
DB_SSLMODE=disable            # SSL mode (require in production)
DB_MAX_OPEN_CONNS=25          # Max open connections
DB_MAX_IDLE_CONNS=5           # Max idle connections
```

### Security Configuration

```env
JWT_SECRET=<64_char_random>    # JWT signing secret
JWT_EXPIRY=24h                 # Token expiration
INITIAL_SETUP_TOKEN=<random>   # One-time setup token (remove after use)
```

### CORS Configuration

```env
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://app.example.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS,PATCH
CORS_MAX_AGE=3600
```

### Logging Configuration

```env
LOG_LEVEL=info                 # debug, info, warn, error, fatal
LOG_FORMAT=json                # json or text
LOG_OUTPUT=stdout              # stdout or file path
```

### Rate Limiting Configuration

```env
RATE_LIMIT_GLOBAL_RPS=100      # Global requests per second
RATE_LIMIT_GLOBAL_BURST=200    # Burst capacity
RATE_LIMIT_AUTH_RPM=1000       # Authenticated requests per minute
RATE_LIMIT_LOGIN_ATTEMPTS=5    # Max login attempts
RATE_LIMIT_LOGIN_WINDOW=5m     # Login window duration
```

---

## 🔧 Development

### Available Make Commands

```bash
make help           # Show all available commands
make build          # Build the application
make run            # Run the application
make test           # Run tests
make clean          # Clean build artifacts
make migrate-up     # Run database migrations
make migrate-down   # Rollback migrations
make sqlc           # Generate SQLC code
make docker-up      # Start PostgreSQL in Docker
make docker-down    # Stop PostgreSQL
make lint           # Run linter
make fmt            # Format code
```

### Project Structure

```
DigiOrder/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── db/                     # Database layer
│   │   ├── connection.go
│   │   ├── query/              # SQL queries
│   │   └── *.sql.go           # Generated SQLC code
│   ├── server/                 # HTTP server
│   │   ├── server.go
│   │   ├── routes.go
│   │   ├── auth.go
│   │   ├── products.go
│   │   ├── orders.go
│   │   ├── users.go
│   │   ├── permissions.go
│   │   ├── audit.go
│   │   └── setup.go
│   ├── middleware/             # HTTP middleware
│   │   ├── auth.go
│   │   ├── rate_limiter_db.go
│   │   ├── cors.go
│   │   ├── cache.go
│   │   ├── logging.go
│   │   └── observability.go
│   ├── security/               # Security utilities
│   │   └── password.go
│   └── logging/                # Structured logging
│       └── logger.go
├── migrations/                 # Database migrations
├── monitoring/                 # Monitoring configuration
│   ├── prometheus/
│   ├── grafana/
│   └── alertmanager/
├── scripts/                    # Utility scripts
├── Dockerfile                  # Production Docker image
├── docker-compose.yml          # Development setup
├── docker-compose.monitoring.yml  # Full stack with monitoring
└── Makefile                    # Build automation
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
go test -v -cover ./...

# Run specific package tests
go test -v ./internal/security/

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Database Operations

```bash
# Create new migration
migrate create -ext sql -dir migrations -seq add_new_feature

# Check migration status
migrate -path migrations -database "postgresql://..." version

# Migrate to specific version
migrate -path migrations -database "postgresql://..." goto 3

# Force version (if migrations are stuck)
migrate -path migrations -database "postgresql://..." force 2
```

---

## 🚀 Production Deployment

### Docker Deployment

#### 1. Build Production Image

```bash
docker build -t digiorder:latest .
```

#### 2. Deploy with Docker Compose

```bash
# Production stack
docker-compose -f docker-compose.prod.yml up -d

# With monitoring
docker-compose -f docker-compose.monitoring.yml up -d
```

#### 3. Verify Deployment

```bash
# Check service health
curl http://your-server:5582/health

# Check logs
docker-compose -f docker-compose.prod.yml logs -f api
```

### Manual Deployment

#### 1. Build Binary

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o digiorder cmd/main.go
```

#### 2. Deploy to Server

```bash
# Copy binary
scp digiorder user@server:/opt/digiorder/

# Copy migrations
scp -r migrations user@server:/opt/digiorder/

# Copy environment file
scp .env.production user@server:/opt/digiorder/.env
```

#### 3. Create Systemd Service

```bash
sudo nano /etc/systemd/system/digiorder.service
```

```ini
[Unit]
Description=DigiOrder API Service
After=network.target postgresql.service

[Service]
Type=simple
User=digiorder
WorkingDirectory=/opt/digiorder
Environment="ENV=production"
ExecStart=/opt/digiorder/digiorder
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start service
sudo systemctl enable digiorder
sudo systemctl start digiorder
sudo systemctl status digiorder
```

### Production Checklist

- [ ] Strong passwords for all accounts
- [ ] JWT_SECRET is 64+ random characters
- [ ] INITIAL_SETUP_TOKEN removed after setup
- [ ] DB_SSLMODE=require in production
- [ ] CORS_ALLOWED_ORIGINS set to production domains
- [ ] SSL/TLS certificates configured
- [ ] Firewall rules configured
- [ ] Database backups automated
- [ ] Monitoring alerts configured
- [ ] Log rotation configured
- [ ] Rate limits adjusted for expected load

---

## 📊 Monitoring

### Prometheus Metrics

Access: http://localhost:9090

**Key Metrics**:

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `http_requests_in_flight` - Concurrent requests
- `db_connections_active` - Active database connections
- `cache_hits_total` - Cache hit count
- `auth_attempts_total` - Authentication attempts
- `rate_limit_exceeded_total` - Rate limit violations

**Sample Queries**:

```promql
# Request rate per second
rate(http_requests_total[5m])

# 95th percentile latency
histogram_quantile(0.95, http_request_duration_seconds_bucket)

# Error rate percentage
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# Cache hit rate
sum(rate(cache_hits_total[5m])) / (sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m]))) * 100
```

### Grafana Dashboards

Access: http://localhost:3000 (admin/admin)

**Included Dashboards**:

1. **System Overview** - API health, request rates, latencies
2. **Business Metrics** - Orders, products, users activity
3. **Performance** - Database queries, cache performance
4. **Security** - Failed logins, rate limit violations

### Alertmanager

Access: http://localhost:9093

**Configured Alerts**:

- API downtime
- High error rate (>5%)
- High response time (>1s)
- Database connection issues
- High authentication failure rate (>30%)

### Log Aggregation

```bash
# View JSON logs
tail -f logs/digiorder.log | jq

# Filter by level
tail -f logs/digiorder.log | jq 'select(.level == "ERROR")'

# Filter by user
tail -f logs/digiorder.log | jq 'select(.user_id == "...")'

# View request logs
tail -f logs/digiorder.log | jq 'select(.message == "HTTP Request")'
```

---

## 🐛 Troubleshooting

### Common Issues

#### 1. "Invalid setup token"

**Problem**: Setup token doesn't match

**Solution**:

```bash
# Check your .env file
grep INITIAL_SETUP_TOKEN .env

# Ensure token matches in request
```

#### 2. "Rate limit exceeded"

**Problem**: Too many requests

**Solution**:

```bash
# Wait for rate limit window to reset (1-5 minutes)
# Or check rate limit records
psql -d digiorder_db -c "SELECT * FROM api_rate_limits WHERE client_id='YOUR_IP';"
```

#### 3. "Password does not meet requirements"

**Problem**: Weak password

**Solution**: Ensure password has:

- 12+ characters
- 1 uppercase letter
- 1 lowercase letter
- 1 digit
- 1 special character (!@#$%^&\*...)

#### 4. "Database connection failed"

**Problem**: Cannot connect to database

**Solution**:

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Verify credentials
psql -h localhost -U postgres -d digiorder_db

# Check Docker logs if using Docker
docker-compose logs postgres
```

#### 5. "SQLC generation errors"

**Problem**: SQL queries not generating

**Solution**:

```bash
# Ensure sqlc is installed
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Regenerate
make sqlc

# Check for syntax errors in .sql files
```

### Getting Help

- **GitHub Issues**: https://github.com/jamalkaksouri/DigiOrder/issues
- **Logs**: Check `logs/digiorder.log` or Docker logs
- **Health Endpoint**: http://localhost:5582/health
- **Metrics**: http://localhost:5582/metrics

---

## 📄 License

MIT License - see LICENSE file for details

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

---

## 🙏 Acknowledgments

Built with:

- [Echo](https://echo.labstack.com/) - High performance Go web framework
- [SQLC](https://sqlc.dev/) - Compile-time safe SQL queries
- [PostgreSQL](https://www.postgresql.org/) - Robust relational database
- [Prometheus](https://prometheus.io/) - Monitoring and alerting
- [Grafana](https://grafana.com/) - Metrics visualization

---

## 📞 Support

For support, please open an issue on GitHub or contact the development team.

---

**DigiOrder v3.0** - Built with ❤️ for modern pharmacy management
