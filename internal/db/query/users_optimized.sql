-- internal/db/query/users_optimized.sql
-- Optimized queries to fix N+1 problem

-- name: ListUsersWithRoles :many
SELECT 
    u.id,
    u.username,
    u.full_name,
    u.password_hash,
    u.role_id,
    u.created_at,
    u.deleted_at,
    r.name as role_name
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
WHERE u.deleted_at IS NULL
ORDER BY u.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetUserWithRole :one
SELECT 
    u.id,
    u.username,
    u.full_name,
    u.password_hash,
    u.role_id,
    u.created_at,
    u.deleted_at,
    r.name as role_name
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
WHERE u.id = $1 AND u.deleted_at IS NULL
LIMIT 1;

-- name: GetUserByUsernameWithRole :one
SELECT 
    u.id,
    u.username,
    u.full_name,
    u.password_hash,
    u.role_id,
    u.created_at,
    u.deleted_at,
    r.name as role_name
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
WHERE u.username = $1 AND u.deleted_at IS NULL
LIMIT 1;

-- name: CountActiveUsers :one
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL;

-- name: SearchUsers :many
SELECT 
    u.id,
    u.username,
    u.full_name,
    u.password_hash,
    u.role_id,
    u.created_at,
    u.deleted_at,
    r.name as role_name
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
WHERE u.deleted_at IS NULL
  AND (
    u.username ILIKE '%' || $1 || '%' 
    OR u.full_name ILIKE '%' || $1 || '%'
  )
ORDER BY u.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUsersByRole :many
SELECT 
    u.id,
    u.username,
    u.full_name,
    u.password_hash,
    u.role_id,
    u.created_at,
    u.deleted_at,
    r.name as role_name
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
WHERE u.deleted_at IS NULL AND u.role_id = $1
ORDER BY u.created_at DESC
LIMIT $2 OFFSET $3;