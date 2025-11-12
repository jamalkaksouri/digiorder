-- name: CreatePermission :one
INSERT INTO permissions (name, resource, action, description)
VALUES (sqlc.arg('name'), sqlc.arg('resource'), sqlc.arg('action'), sqlc.arg('description'))
RETURNING *;

-- name: GetPermission :one
SELECT * FROM permissions WHERE id = sqlc.arg('id');

-- name: ListPermissions :many
SELECT * FROM permissions
ORDER BY resource, action
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListPermissionsByResource :many
SELECT * FROM permissions
WHERE resource = sqlc.arg('resource')
ORDER BY action
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: UpdatePermission :one
UPDATE permissions
SET
    name = COALESCE(NULLIF(sqlc.narg('name'), ''), name),
    resource = COALESCE(NULLIF(sqlc.narg('resource'), ''), resource),
    action = COALESCE(NULLIF(sqlc.narg('action'), ''), action),
    description = COALESCE(NULLIF(sqlc.narg('description'), ''), description)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = sqlc.arg('id');

-- name: AssignPermissionToRole :one
INSERT INTO role_permissions (role_id, permission_id)
VALUES (sqlc.arg('role_id'), sqlc.arg('permission_id'))
RETURNING *;

-- name: RevokePermissionFromRole :exec
DELETE FROM role_permissions
WHERE role_id = sqlc.arg('role_id') AND permission_id = sqlc.arg('permission_id');

-- name: GetRolePermissions :many
SELECT p.* FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = sqlc.arg('role_id')
ORDER BY p.resource, p.action;

-- name: CheckRolePermission :one
SELECT EXISTS(
    SELECT 1 FROM role_permissions rp
    JOIN permissions p ON rp.permission_id = p.id
    WHERE rp.role_id = sqlc.arg('role_id')
    AND p.resource = sqlc.arg('resource')
    AND p.action = sqlc.arg('action')
) as has_permission;

-- name: ListActiveUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = sqlc.arg('id');

-- name: CountAdminUsers :one
SELECT COUNT(*) FROM users
WHERE role_id = 1 AND deleted_at IS NULL;

-- name: CreateAuditLog :one
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent)
VALUES (
    sqlc.arg('user_id'),
    sqlc.arg('action'),
    sqlc.arg('entity_type'),
    sqlc.arg('entity_id'),
    sqlc.arg('old_values'),
    sqlc.arg('new_values'),
    sqlc.arg('ip_address'),
    sqlc.arg('user_agent')
)
RETURNING *;

-- name: GetAuditLog :one
SELECT * FROM audit_logs WHERE id = sqlc.arg('id');

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetAuditLogsByUser :many
SELECT * FROM audit_logs
WHERE user_id = sqlc.arg('user_id')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetAuditLogsByEntity :many
SELECT * FROM audit_logs
WHERE entity_type = sqlc.arg('entity_type')
  AND entity_id = sqlc.arg('entity_id')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetAuditLogsByAction :many
SELECT * FROM audit_logs
WHERE action = sqlc.arg('action')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetAuditLogStats :one
SELECT
    COUNT(*) as total_logs,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT entity_type) as unique_entities
FROM audit_logs
WHERE created_at >= NOW() - INTERVAL '24 hours';
