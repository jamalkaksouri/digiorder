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