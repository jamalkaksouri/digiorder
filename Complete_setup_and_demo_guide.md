# DigiOrder v3.0 - Complete Setup Guide

## With Admin Protection, Permissions, Audit Logging & Full Observability

---

## 🚀 Quick Start (5 Minutes)

```bash
# 1. Clone and setup
git clone https://github.com/jamalkaksouri/DigiOrder.git
cd DigiOrder

# 2. Setup environment
cp .env.production .env
# Edit .env with your database credentials and JWT secret

# 3. Start monitoring stack
docker-compose -f docker-compose.monitoring.yml up -d

# 4. Wait for services to be ready
sleep 30

# 5. Access services
# - API: http://localhost:5582
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3000 (admin/admin)
# - Alertmanager: http://localhost:9093
```

---

## 📋 Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for local development)
- PostgreSQL 15+ (or use Docker)
- 8GB RAM minimum
- 20GB disk space

---

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │   API    │  │  Auth    │  │Permissions│  │  Audit   │   │
│  └────┬─────┘  └────┬─────┘  └────┬──────┘  └────┬─────┘   │
│       │             │              │              │          │
├───────┴─────────────┴──────────────┴──────────────┴─────────┤
│                    Middleware Layer                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │  Rate    │  │  Cache   │  │  Metrics │  │  Trace   │   │
│  │  Limit   │  │          │  │          │  │    ID    │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
├──────────────────────────────────────────────────────────────┤
│                    Database Layer                            │
│  ┌────────────────────────┐  ┌────────────────────────┐    │
│  │     PostgreSQL         │  │     Redis (Optional)   │    │
│  │  - Products            │  │  - Session Cache       │    │
│  │  - Orders              │  │  - Rate Limit Cache    │    │
│  │  - Users               │  └────────────────────────┘    │
│  │  - Permissions         │                                 │
│  │  - Audit Logs          │                                 │
│  └────────────────────────┘                                 │
├──────────────────────────────────────────────────────────────┤
│                 Observability Layer                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │Prometheus│  │ Grafana  │  │   Loki   │  │Alertmgr  │   │
│  │ Metrics  │  │Dashboards│  │   Logs   │  │  Alerts  │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└──────────────────────────────────────────────────────────────┘
```

---

## 🔧 Installation Steps

### 1. Create Directory Structure

```bash
mkdir -p monitoring/{prometheus,grafana/{provisioning/{datasources,dashboards},dashboards},alertmanager,loki,promtail}
```

### 2. Configuration Files

Create all configuration files from the artifacts provided:

- `docker-compose.monitoring.yml`
- `monitoring/prometheus/prometheus.yml`
- `monitoring/prometheus/alerts.yml`
- `monitoring/alertmanager/alertmanager.yml`
- `monitoring/grafana/provisioning/datasources/prometheus.yml`
- `monitoring/grafana/provisioning/dashboards/dashboard.yml`
- `monitoring/grafana/dashboards/digiorder-overview.json`
- `monitoring/grafana/dashboards/digiorder-business.json`

### 3. Database Migrations

Create new migration file:

```sql
-- migrations/000003_add_permissions.up.sql

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(resource, action)
);

-- Role-Permission junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

-- Add indexes
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);

-- Insert default permissions
INSERT INTO permissions (name, resource, action, description) VALUES
    -- Product permissions
    ('view_products', 'products', 'read', 'View products'),
    ('create_products', 'products', 'create', 'Create products'),
    ('update_products', 'products', 'update', 'Update products'),
    ('delete_products', 'products', 'delete', 'Delete products'),

    -- Order permissions
    ('view_orders', 'orders', 'read', 'View orders'),
    ('create_orders', 'orders', 'create', 'Create orders'),
    ('update_orders', 'orders', 'update', 'Update orders'),
    ('delete_orders', 'orders', 'delete', 'Delete orders'),

    -- User permissions
    ('view_users', 'users', 'read', 'View users'),
    ('create_users', 'users', 'create', 'Create users'),
    ('update_users', 'users', 'update', 'Update users'),
    ('delete_users', 'users', 'delete', 'Delete users'),

    -- Audit permissions
    ('view_audit_logs', 'audit', 'read', 'View audit logs'),

    -- System permissions
    ('manage_permissions', 'permissions', 'manage', 'Manage permissions'),
    ('manage_roles', 'roles', 'manage', 'Manage roles');

-- Assign all permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions;

-- Assign limited permissions to pharmacist
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions
WHERE name IN ('view_products', 'create_products', 'update_products',
               'view_orders', 'create_orders', 'update_orders');

-- Assign read-only permissions to clerk
INSERT INTO role_permissions (role_id, permission_id)
SELECT 3, id FROM permissions
WHERE action = 'read';

-- Create primary admin user (protected from deletion)
INSERT INTO users (id, username, full_name, password_hash, role_id)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin',
    'Primary Administrator',
    '$2a$10$Zu7yVNJ0e9Fn9vwUy9vRbO5CqPQZMB8l5k8hEWnGvhkrFUKqj9iEW',
    1
) ON CONFLICT (id) DO NOTHING;
```

```sql
-- migrations/000003_add_permissions.down.sql

DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
```

### 4. Update SQLC Queries

Create `internal/db/query/permissions.sql`:

```sql
-- name: CreatePermission :one
INSERT INTO permissions (name, resource, action, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPermission :one
SELECT * FROM permissions WHERE id = $1;

-- name: ListPermissions :many
SELECT * FROM permissions
ORDER BY resource, action
LIMIT $1 OFFSET $2;

-- name: ListPermissionsByResource :many
SELECT * FROM permissions
WHERE resource = $1
ORDER BY action
LIMIT $2 OFFSET $3;

-- name: UpdatePermission :one
UPDATE permissions
SET
    name = COALESCE(NULLIF($2, ''), name),
    resource = COALESCE(NULLIF($3, ''), resource),
    action = COALESCE(NULLIF($4, ''), action),
    description = COALESCE(NULLIF($5, ''), description)
WHERE id = $1
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1;

-- name: AssignPermissionToRole :one
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
RETURNING *;

-- name: RevokePermissionFromRole :exec
DELETE FROM role_permissions
WHERE role_id = $1 AND permission_id = $2;

-- name: GetRolePermissions :many
SELECT p.* FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1
ORDER BY p.resource, p.action;

-- name: CheckRolePermission :one
SELECT EXISTS(
    SELECT 1 FROM role_permissions rp
    JOIN permissions p ON rp.permission_id = p.id
    WHERE rp.role_id = $1
    AND p.resource = $2
    AND p.action = $3
) as has_permission;

-- name: ListActiveUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = NOW()
WHERE id = $1;

-- name: CountAdminUsers :one
SELECT COUNT(*) FROM users
WHERE role_id = 1 AND deleted_at IS NULL;

-- name: CreateAuditLog :one
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetAuditLog :one
SELECT * FROM audit_logs WHERE id = $1;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetAuditLogsByUser :many
SELECT * FROM audit_logs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAuditLogsByEntity :many
SELECT * FROM audit_logs
WHERE entity_type = $1 AND entity_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetAuditLogsByAction :many
SELECT * FROM audit_logs
WHERE action = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAuditLogStats :one
SELECT
    COUNT(*) as total_logs,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT entity_type) as unique_entities
FROM audit_logs
WHERE created_at >= NOW() - INTERVAL '24 hours';
```

### 5. Generate SQLC Code

```bash
sqlc generate
```

### 6. Update Routes

Add new routes in `internal/server/routes.go`:

```go
// Permission routes (admin only)
permissions := protected.Group("/permissions")
permissions.Use(middleware.RequireRole("admin"))
{
    permissions.POST("", s.CreatePermission)
    permissions.GET("", s.ListPermissions)
    permissions.GET("/:id", s.GetPermission)
    permissions.PUT("/:id", s.UpdatePermission)
    permissions.DELETE("/:id", s.DeletePermission)
}

// Role permission management
{
    protected.POST("/roles/:role_id/permissions", s.AssignPermissionToRole, middleware.RequireRole("admin"))
    protected.GET("/roles/:role_id/permissions", s.GetRolePermissions)
    protected.DELETE("/roles/:role_id/permissions/:permission_id", s.RevokePermissionFromRole, middleware.RequireRole("admin"))
}

// Audit log routes (admin only)
auditLogs := protected.Group("/audit-logs")
auditLogs.Use(middleware.RequireRole("admin"))
{
    auditLogs.GET("", s.GetAuditLogs)
    auditLogs.GET("/:id", s.GetAuditLog)
    auditLogs.GET("/entity/:type/:id", s.GetEntityHistory)
    auditLogs.GET("/stats", s.GetAuditStats)
}

// User activity
protected.GET("/users/:user_id/activity", s.GetUserActivity, middleware.RequireRole("admin"))

// Permission check
protected.GET("/auth/check-permission", s.CheckUserPermission)

// Prometheus metrics
s.router.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
```

### 7. Update Main Server

Add Prometheus middleware in `internal/server/routes.go`:

```go
import (
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// In registerRoutes() function, add:
s.router.Use(middleware.PrometheusMiddleware())
s.router.Use(middleware.TracingMiddleware())
```

---

## 🧪 Testing All Features

### 1. Start Services

```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

### 2. Run Feature Demos

```bash
# Rate Limiting Demo
chmod +x scripts/rate_limit_demo.sh
./scripts/rate_limit_demo.sh

# Barcode Demo
chmod +x scripts/barcode_demo.sh
./scripts/barcode_demo.sh

# Audit Logging Demo
chmod +x scripts/audit_demo.sh
./scripts/audit_demo.sh
```

### 3. Test Admin Protection

```bash
# Try to delete primary admin (should fail)
TOKEN=$(curl -s -X POST http://localhost:5582/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123456"}' \
  | jq -r '.data.token')

curl -X DELETE http://localhost:5582/api/v1/users/00000000-0000-0000-0000-000000000001 \
  -H "Authorization: Bearer $TOKEN"

# Expected: 403 Forbidden - "The primary administrator account cannot be deleted"
```

### 4. Test Permissions

```bash
# Check user permission
curl http://localhost:5582/api/v1/auth/check-permission?resource=products&action=create \
  -H "Authorization: Bearer $TOKEN"

# List role permissions
curl http://localhost:5582/api/v1/roles/2/permissions \
  -H "Authorization: Bearer $TOKEN"
```

### 5. View Metrics

```bash
# Prometheus metrics
curl http://localhost:5582/metrics

# View in Prometheus UI
open http://localhost:9090

# Sample queries:
# - rate(http_requests_total[5m])
# - histogram_quantile(0.95, http_request_duration_seconds_bucket)
# - db_connections_active
```

### 6. Access Grafana

```bash
# Open Grafana
open http://localhost:3000

# Login: admin/admin
# Navigate to Dashboards → DigiOrder - System Overview
```

---

## 📊 Available Grafana Dashboards

1. **System Overview**

   - API health status
   - Request rates
   - Response times
   - Error rates
   - Database connections

2. **Business Metrics**

   - Orders created
   - Products created
   - Active users
   - Order status distribution

3. **Performance Metrics**

   - P50/P95/P99 latencies
   - Throughput
   - Cache hit rates

4. **Infrastructure Metrics**
   - CPU usage
   - Memory usage
   - Disk space
   - Network I/O

---

## 🔔 Alerts Configuration

Alerts are configured in `monitoring/prometheus/alerts.yml`:

- **Critical**: API down, Database down
- **Warning**: High error rate, High response time, Low cache hit rate
- **Info**: No orders created, Rate limit exceeded

Configure email/Slack notifications in `monitoring/alertmanager/alertmanager.yml`.

---

## 🎯 Feature Summary

### ✅ Admin Protection

- Primary admin (UUID: 00000000...001) cannot be deleted
- Last admin in system cannot be deleted
- Only admins can create users

### ✅ Permission System

- CRUD operations for permissions
- Role-permission assignment
- Dynamic permission checking
- Resource-action based permissions

### ✅ Audit Logging

- Complete audit trail for all actions
- User activity tracking
- Entity history
- IP and User-Agent tracking

### ✅ Rate Limiting

- Global: 100 req/sec (burst 200)
- Authenticated: 1000 req/min
- Per-IP tracking
- Automatic recovery

### ✅ Barcode Support

- Multiple barcode types (EAN-13, UPC-A, Code128)
- Quick product lookup
- Barcode CRUD operations
- Scanner integration ready

### ✅ Soft Deletes

- Users, Orders, Products
- Recovery possible
- Audit trail preserved

### ✅ Full Observability

- Prometheus metrics
- Grafana dashboards
- Request tracing
- Alert management
- Log aggregation

---

## 📈 Performance Metrics

Monitor these key metrics in Grafana:

- **Request Rate**: Target < 100 req/sec
- **P95 Latency**: Target < 500ms
- **Error Rate**: Target < 1%
- **Cache Hit Rate**: Target > 80%
- **DB Connections**: Target < 20
- **CPU Usage**: Target < 70%
- **Memory Usage**: Target < 80%

---

## 🎓 Next Steps

1. **Customize Dashboards**: Add business-specific metrics
2. **Configure Alerts**: Set up email/Slack notifications
3. **Enable HTTPS**: Add SSL certificates
4. **Scale Horizontally**: Add more API instances
5. **Backup Strategy**: Configure automated backups
6. **Security Hardening**: Enable firewall, rate limits
7. **CI/CD Integration**: Automate deployments

---

## 📞 Support

- Documentation: `/docs`
- Metrics: http://localhost:9090
- Dashboards: http://localhost:3000
- Alerts: http://localhost:9093

---

**Congratulations! 🎉 You now have a production-ready, enterprise-grade pharmacy management system with complete observability!**
