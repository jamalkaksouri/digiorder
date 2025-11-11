# DigiOrder - Project Overview

## Table of Contents

1. [Introduction](#introduction)
2. [System Architecture](#system-architecture)
3. [Technology Stack](#technology-stack)
4. [Directory Structure](#directory-structure)
5. [Database Schema](#database-schema)
6. [Security & Authentication](#security--authentication)
7. [API Design Patterns](#api-design-patterns)
8. [Deployment](#deployment)
9. [Monitoring & Observability](#monitoring--observability)
10. [Development Guidelines](#development-guidelines)

---

## Introduction

**DigiOrder v3.0** is an enterprise-grade pharmacy order management system built with Go, PostgreSQL, and modern DevOps practices. It provides a complete solution for managing pharmaceutical inventory, orders, users, and audit trails with robust security and observability features.

### Key Features

- 🔐 **JWT-based Authentication** with role-based access control (RBAC)
- 📦 **Product Management** with barcode scanning support
- 📋 **Order Processing** with status tracking and item management
- 👥 **User Management** with protected admin accounts
- 🔍 **Audit Logging** for complete activity tracking
- 🚦 **Rate Limiting** to prevent API abuse
- 📊 **Full Observability** with Prometheus, Grafana, and distributed tracing
- ♻️ **Soft Deletes** for data recovery
- 🔑 **Permission System** for granular access control

---

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Applications                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Web UI   │  │  Mobile  │  │ Barcode  │  │  API     │   │
│  │          │  │   App    │  │ Scanner  │  │ Consumer │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │
└───────┼─────────────┼─────────────┼─────────────┼──────────┘
        │             │              │             │
        └─────────────┴──────────────┴─────────────┘
                            │
                    ┌───────▼────────┐
                    │  Load Balancer │
                    │     (Nginx)    │
                    └───────┬────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
    ┌───▼────┐         ┌───▼────┐         ┌───▼────┐
    │  API   │         │  API   │         │  API   │
    │Instance│         │Instance│         │Instance│
    │   #1   │         │   #2   │         │   #3   │
    └───┬────┘         └───┬────┘         └───┬────┘
        │                  │                   │
        └──────────────────┼───────────────────┘
                           │
        ┌──────────────────┴──────────────────┐
        │                                     │
    ┌───▼────────┐                    ┌──────▼──────┐
    │ PostgreSQL │                    │    Redis    │
    │  Database  │                    │   (Cache)   │
    └────────────┘                    └─────────────┘
```

### Application Layer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Request Flow                         │
└─────────────────────────────────────────────────────────────┘
                            │
                    ┌───────▼────────┐
                    │  Echo Router   │
                    └───────┬────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │          Middleware Stack             │
        │  ┌────────────────────────────────┐  │
        │  │ 1. Logger (Request Logging)    │  │
        │  │ 2. Recover (Panic Recovery)    │  │
        │  │ 3. CORS (Cross-Origin)         │  │
        │  │ 4. Request ID (Tracing)        │  │
        │  │ 5. Security Headers            │  │
        │  │ 6. Prometheus (Metrics)        │  │
        │  │ 7. Rate Limiter (100/sec)      │  │
        │  │ 8. JWT Validator (Auth)        │  │
        │  │ 9. Role Checker (RBAC)         │  │
        │  │ 10. Cache (5min TTL)           │  │
        │  └────────────────────────────────┘  │
        └───────────────────┬───────────────────┘
                            │
                    ┌───────▼────────┐
                    │    Handlers    │
                    │  (Controllers) │
                    └───────┬────────┘
                            │
                    ┌───────▼────────┐
                    │  Validators    │
                    │ (go-playground)│
                    └───────┬────────┘
                            │
                    ┌───────▼────────┐
                    │  SQLC Queries  │
                    │ (Type-safe DB) │
                    └───────┬────────┘
                            │
                    ┌───────▼────────┐
                    │   PostgreSQL   │
                    └────────────────┘
```

---

## Technology Stack

### Backend

- **Language:** Go 1.25+
- **Web Framework:** Echo v4
- **Database:** PostgreSQL 15+
- **ORM/Query Builder:** SQLC (compile-time type-safe SQL)
- **Migration Tool:** golang-migrate/migrate
- **Authentication:** JWT (golang-jwt/jwt)
- **Validation:** go-playground/validator
- **Password Hashing:** bcrypt

### Monitoring & Observability

- **Metrics:** Prometheus
- **Visualization:** Grafana
- **Alerting:** Alertmanager
- **Tracing:** Distributed request tracing with X-Trace-ID
- **Logging:** Structured JSON logging

### DevOps

- **Containerization:** Docker
- **Orchestration:** Docker Compose
- **CI/CD:** GitHub Actions
- **Reverse Proxy:** Nginx

### Development Tools

- **Code Generation:** SQLC for database queries
- **API Testing:** cURL, Postman collections
- **Linting:** golangci-lint
- **Testing:** Go testing framework + testify

---

## Directory Structure

```
DigiOrder/
├── cmd/
│   └── main.go                    # Application entry point
│
├── internal/                      # Private application code
│   ├── db/                        # Database layer
│   │   ├── connection.go          # DB connection management
│   │   ├── models.go              # Generated SQLC models
│   │   ├── *.sql.go              # Generated query functions
│   │   └── query/                 # SQL query definitions
│   │       ├── products.sql
│   │       ├── orders.sql
│   │       ├── users.sql
│   │       ├── permissions.sql
│   │       └── barcodes.sql
│   │
│   ├── middleware/                # HTTP middleware
│   │   ├── auth.go               # JWT authentication
│   │   ├── rate_limiter.go       # Rate limiting
│   │   ├── cache.go              # Response caching
│   │   ├── logging.go            # Request logging
│   │   └── observability.go      # Prometheus metrics
│   │
│   └── server/                    # HTTP server and handlers
│       ├── server.go             # Server initialization
│       ├── routes.go             # Route registration
│       ├── response.go           # Response helpers
│       ├── auth.go               # Authentication handlers
│       ├── products.go           # Product CRUD handlers
│       ├── orders.go             # Order management handlers
│       ├── users.go              # User management handlers
│       ├── permissions.go        # Permission system handlers
│       ├── audit.go              # Audit logging handlers
│       ├── barcodes.go           # Barcode management
│       ├── categories.go         # Category handlers
│       ├── dosage_forms.go       # Dosage form handlers
│       └── roles.go              # Role management
│
├── migrations/                    # Database migrations
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000002_add_features.up.sql
│   ├── 000002_add_features.down.sql
│   ├── 000003_add_permissions.up.sql
│   └── 000003_add_permissions.down.sql
│
├── monitoring/                    # Observability configuration
│   ├── prometheus/
│   │   ├── prometheus.yml        # Prometheus config
│   │   └── alerts.yml            # Alert rules
│   ├── grafana/
│   │   ├── provisioning/         # Auto-provisioning
│   │   └── dashboards/           # Dashboard definitions
│   └── alertmanager/
│       └── alertmanager.yml      # Alert routing
│
├── scripts/                       # Utility scripts
│   ├── rate_limit.sh             # Rate limit demo
│   ├── barcode_support.sh        # Barcode demo
│   └── complete_feature_demonstration.sh
│
├── docs/                          # Documentation
│   ├── API_TESTING_GUIDE.md
│   ├── AUTHENTICATION_GUIDE.md
│   ├── DEPLOYMENT_GUIDE.md
│   └── Complete_setup_and_demo_guide.md
│
├── docker-compose.yml             # Development setup
├── docker-compose.prod.yml        # Production setup
├── docker-compose.monitoring.yml  # Full observability stack
├── Dockerfile                     # Container image definition
├── Makefile                       # Build automation
├── sqlc.yaml                      # SQLC configuration
├── go.mod                         # Go dependencies
└── README.md                      # Project README
```

### Directory Responsibilities

| Directory              | Purpose                                                     |
| ---------------------- | ----------------------------------------------------------- |
| `cmd/`                 | Application entry points and main functions                 |
| `internal/`            | Private application code (not importable by other projects) |
| `internal/db/`         | Database access layer with type-safe queries                |
| `internal/middleware/` | HTTP middleware for cross-cutting concerns                  |
| `internal/server/`     | HTTP handlers and business logic                            |
| `migrations/`          | Database schema versioning                                  |
| `monitoring/`          | Observability stack configuration                           |
| `scripts/`             | Automation and demonstration scripts                        |

---

## Database Schema

### Entity Relationship Diagram

```
┌─────────────┐
│    roles    │
│─────────────│
│ id (PK)     │
│ name        │
└──────┬──────┘
       │
       │ 1:N
       │
┌──────▼──────────┐      ┌──────────────────┐
│     users       │      │   permissions    │
│─────────────────│      │──────────────────│
│ id (PK)         │      │ id (PK)          │
│ username (UK)   │      │ name (UK)        │
│ full_name       │      │ resource         │
│ password_hash   │      │ action           │
│ role_id (FK)    │      │ description      │
│ created_at      │      └────────┬─────────┘
│ deleted_at      │               │
└──────┬──────────┘               │
       │                          │
       │ 1:N              N:M     │
       │             ┌────────────┴──────────┐
       │             │  role_permissions     │
       │             │───────────────────────│
       │             │ id (PK)               │
       │             │ role_id (FK)          │
       │             │ permission_id (FK)    │
       │             └───────────────────────┘
       │
       │ 1:N
       │
┌──────▼──────────┐
│     orders      │
│─────────────────│
│ id (PK)         │───┐
│ created_by (FK) │   │
│ status          │   │
│ created_at      │   │ 1:N
│ submitted_at    │   │
│ notes           │   │
│ deleted_at      │   │
└─────────────────┘   │
                      │
                      │
        ┌─────────────▼──────────┐
        │     order_items        │
        │────────────────────────│
        │ id (PK)                │
        │ order_id (FK)          │
        │ product_id (FK)        │───┐
        │ requested_qty          │   │
        │ unit                   │   │
        │ note                   │   │
        └────────────────────────┘   │
                                     │
                                     │ N:1
┌────────────────────────────────────┘
│
├─────────────────┐
│    products     │
│─────────────────│
│ id (PK)         │
│ name            │
│ brand           │
│ dosage_form_id (FK)
│ strength        │
│ unit            │
│ category_id (FK)│
│ description     │
│ created_at      │
│ deleted_at      │
└──────┬──────────┘
       │
       │ 1:N
       │
┌──────▼──────────────┐
│  product_barcodes   │
│─────────────────────│
│ id (PK)             │
│ product_id (FK)     │
│ barcode (UK)        │
│ barcode_type        │
│ created_at          │
└─────────────────────┘

┌─────────────────┐
│   categories    │
│─────────────────│
│ id (PK)         │
│ name (UK)       │
└─────────────────┘

┌─────────────────┐
│  dosage_forms   │
│─────────────────│
│ id (PK)         │
│ name (UK)       │
└─────────────────┘

┌─────────────────────┐
│    audit_logs       │
│─────────────────────│
│ id (PK)             │
│ user_id (FK)        │
│ action              │
│ entity_type         │
│ entity_id           │
│ old_values (JSONB)  │
│ new_values (JSONB)  │
│ ip_address          │
│ user_agent          │
│ created_at          │
└─────────────────────┘
```

### Core Tables

#### users

Stores user accounts with authentication credentials.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    full_name TEXT,
    password_hash TEXT NOT NULL,
    role_id INT REFERENCES roles(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  -- Soft delete support
);
```

**Key Features:**

- UUID primary key for distributed systems
- Bcrypt hashed passwords (cost factor 10)
- Soft delete capability
- Unique username constraint

#### products

Pharmaceutical product catalog.

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    brand TEXT,
    dosage_form_id INT REFERENCES dosage_forms(id),
    strength TEXT,
    unit TEXT,
    category_id INT REFERENCES categories(id),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

**Indexes:**

- `idx_products_name` - Fast product search by name
- `idx_products_category` - Filter by category
- `idx_products_deleted_at` - Soft delete queries

#### orders

Order tracking system.

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by UUID REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    submitted_at TIMESTAMPTZ,
    notes TEXT,
    deleted_at TIMESTAMPTZ
);
```

**Status Flow:**

1. `draft` → Order being created
2. `submitted` → Order sent to supplier
3. `processing` → Order being fulfilled
4. `completed` → Order received
5. `cancelled` → Order cancelled

#### permissions

Granular access control system.

```sql
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(resource, action)
);
```

**Example Permissions:**

- `view_products` → `products:read`
- `create_orders` → `orders:create`
- `manage_permissions` → `permissions:manage`

#### audit_logs

Complete activity tracking.

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    old_values JSONB,
    new_values JSONB,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Captures:**

- Who performed the action
- What was changed (before/after JSON snapshots)
- When it happened
- From where (IP address)
- Using what client (User-Agent)

---

## Security & Authentication

### Authentication Flow

```
┌─────────┐                                  ┌─────────┐
│ Client  │                                  │  Server │
└────┬────┘                                  └────┬────┘
     │                                            │
     │  POST /api/v1/auth/login                  │
     │  {"username":"...", "password":"..."}     │
     ├──────────────────────────────────────────>│
     │                                            │
     │                                            │ Verify
     │                                            │ Password
     │                                            │ (bcrypt)
     │                                            │
     │  200 OK                                    │
     │  {"token":"eyJ...", "user":{...}}         │
     │<──────────────────────────────────────────┤
     │                                            │
     │  GET /api/v1/products                     │
     │  Authorization: Bearer eyJ...             │
     ├──────────────────────────────────────────>│
     │                                            │
     │                                            │ Validate
     │                                            │ JWT
     │                                            │ Check Role
     │                                            │
     │  200 OK                                    │
     │  {"data": [...]}                          │
     │<──────────────────────────────────────────┤
     │                                            │
```

### JWT Token Structure

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "admin",
  "role_id": 1,
  "role_name": "admin",
  "exp": 1704153600, // Expiry (24h default)
  "iat": 1704067200, // Issued at
  "nbf": 1704067200, // Not before
  "iss": "digiorder-api"
}
```

### Role-Based Access Control (RBAC)

| Role       | ID  | Can Create Users | Can Delete Products | Can View Audit |
| ---------- | --- | ---------------- | ------------------- | -------------- |
| admin      | 1   | ✅               | ✅                  | ✅             |
| pharmacist | 2   | ❌               | ❌                  | ❌             |
| clerk      | 3   | ❌               | ❌                  | ❌             |

### Permission System

**Resource-Action Model:**

```
permission = resource + action
Example: products:create, orders:read, users:delete
```

**Permission Assignment:**

```
Role → Permissions (N:M relationship)
User → Role (N:1 relationship)
Therefore: User → Permissions (derived)
```

### Security Best Practices Implemented

1. **Password Security**

   - ✅ Bcrypt hashing (cost factor 10)
   - ✅ Minimum 8 characters (configurable)
   - ✅ Password change endpoint with old password verification
   - ⚠️ No complexity requirements (recommendation: add)

2. **Authentication**

   - ✅ JWT with expiry (24h default, configurable)
   - ✅ Token refresh mechanism
   - ✅ Secure token generation with HS256
   - ✅ Per-request authentication validation

3. **Authorization**

   - ✅ Role-based access control
   - ✅ Permission-based granular control
   - ✅ Protected admin account (cannot be deleted)
   - ✅ Middleware enforcement on all protected routes

4. **Input Validation**

   - ✅ go-playground/validator for struct validation
   - ✅ UUID validation for IDs
   - ✅ SQL injection prevention via SQLC parameterized queries
   - ✅ Error message sanitization

5. **Rate Limiting**

   - ✅ Global: 100 req/sec (burst 200)
   - ✅ Authenticated: 1000 req/min
   - ✅ Per-IP tracking
   - ✅ Automatic cleanup of inactive limiters

6. **Audit Logging**
   - ✅ Complete action tracking
   - ✅ Before/after snapshots (JSONB)
   - ✅ IP and User-Agent logging
   - ✅ Entity history tracking

---

## API Design Patterns

### RESTful Conventions

```
Resource: /api/v1/products

GET    /products          → List all (paginated)
GET    /products/:id      → Get one
POST   /products          → Create new
PUT    /products/:id      → Update existing
DELETE /products/:id      → Delete
GET    /products/search   → Search with query params
```

### Response Format

**Success Response:**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Amoxicillin 500mg",
    "created_at": "2025-11-10T10:30:00Z"
  }
}
```

**Error Response:**

```json
{
  "error": "validation_error",
  "details": "The 'name' field is required."
}
```

### Pagination

All list endpoints support pagination:

```
GET /api/v1/products?limit=50&offset=0
```

**Default Values:**

- `limit`: 50
- `offset`: 0

### Filtering

```
GET /api/v1/orders?user_id=550e8400-e29b-41d4-a716-446655440000
GET /api/v1/products/search?q=aspirin
```

### Soft Deletes

Resources marked with `deleted_at` are:

- ✅ Excluded from list queries
- ✅ Return 404 on direct access
- ✅ Preserved in database for audit
- ✅ Recoverable by clearing `deleted_at`

---

## Deployment

### Docker Deployment

**Development:**

```bash
docker-compose up -d
```

**Production:**

```bash
docker-compose -f docker-compose.prod.yml up -d
```

**With Monitoring:**

```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

### Environment Variables

**Required:**

```bash
DB_HOST=postgres
DB_PORT=5432
DB_USER=digiorder_prod
DB_PASSWORD=<STRONG_PASSWORD>
DB_NAME=digiorder_production
JWT_SECRET=<64_CHAR_RANDOM_STRING>
```

**Optional:**

```bash
JWT_EXPIRY=24h
SERVER_PORT=5582
RATE_LIMIT_RPS=100
CACHE_TTL_MINUTES=5
```

### CI/CD Pipeline

**GitHub Actions Workflow:**

1. **Lint** - Code quality checks
2. **Test** - Run all tests with PostgreSQL
3. **Build** - Compile binary and Docker image
4. **Push** - Push to Docker Hub
5. **Deploy** - SSH to production server and deploy

**Trigger:** Push to `main` branch

---

## Monitoring & Observability

### Metrics Collected

**HTTP Metrics:**

- `http_requests_total` - Total requests by method, endpoint, status
- `http_request_duration_seconds` - Response time histogram
- `http_requests_in_flight` - Current active requests
- `http_request_size_bytes` - Request payload size
- `http_response_size_bytes` - Response payload size

**Database Metrics:**

- `db_connections_active` - Active DB connections
- `db_connections_idle` - Idle DB connections
- `db_queries_total` - Query count by operation
- `db_query_duration_seconds` - Query execution time

**Business Metrics:**

- `orders_created_total` - Orders created by status
- `products_created_total` - Products added to catalog
- `users_active_total` - Active user count

**Cache Metrics:**

- `cache_hits_total` - Cache hits
- `cache_misses_total` - Cache misses

### Dashboards

**System Overview:**

- API health status
- Request rate and latency
- Error rate
- Database performance

**Business Metrics:**

- Orders created (24h, 7d, 30d)
- Popular products
- User activity

### Alerting Rules

**Critical Alerts:**

- API down for >1 minute
- Database down for >1 minute
- Error rate >5% for 5 minutes

**Warning Alerts:**

- Response time p95 >1 second
- Cache hit rate <50%
- High authentication failures

---

## Development Guidelines

### Code Style

- **Follow Go conventions:** `gofmt`, `golint`
- **Error handling:** Always check and handle errors
- **Naming:** Use clear, descriptive names
- **Comments:** Document exported functions and complex logic

### Testing

```bash
# Run all tests
make test

# With coverage
go test -cover ./...

# Specific package
go test -v ./internal/server
```

### Database Migrations

**Create New Migration:**

```bash
migrate create -ext sql -dir migrations -seq add_new_feature
```

**Apply Migrations:**

```bash
make migrate-up
```

**Rollback:**

```bash
make migrate-down
```

### Adding New Endpoints

1. Define SQL queries in `internal/db/query/*.sql`
2. Run `make sqlc` to generate Go code
3. Create handler in `internal/server/*.go`
4. Register route in `internal/server/routes.go`
5. Add tests
6. Update documentation

### Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

---

## Troubleshooting

### Common Issues

**1. Database Connection Failed**

```
Error: failed to ping database
```

**Solution:** Check DB credentials in `.env`, ensure PostgreSQL is running

**2. JWT Token Invalid**

```
Error: invalid or expired token
```

**Solution:** Token expired (24h default), request refresh or re-login

**3. Rate Limit Exceeded**

```
Error: 429 Too Many Requests
```

**Solution:** Wait for rate limit window to reset (1 second)

**4. Migration Failed**

```
Error: Dirty database version
```

**Solution:** Force migration version: `migrate force <version>`

---

## Support & Resources

- **Documentation:** `/docs` directory
- **API Reference:** `API_REFERENCE.md`
- **GitHub Issues:** Report bugs and feature requests
- **Monitoring:** http://localhost:3000 (Grafana)
- **Metrics:** http://localhost:9090 (Prometheus)

---

**Version:** 3.0.0  
**Last Updated:** November 10, 2025  
**License:** MIT
