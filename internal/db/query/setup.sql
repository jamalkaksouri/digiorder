-- internal/db/query/setup.sql

-- name: GetSystemSetupStatus :one
SELECT * FROM system_setup
ORDER BY created_at DESC
LIMIT 1;

-- name: CompleteSystemSetup :one
UPDATE system_setup
SET 
    admin_created = $1,
    setup_completed_at = NOW(),
    setup_by_ip = $2
WHERE id = (SELECT id FROM system_setup ORDER BY created_at DESC LIMIT 1)
RETURNING *;

-- name: CreateAdminUser :one
INSERT INTO users (id, username, full_name, password_hash, role_id, created_at)
VALUES ($1, $2, $3, $4, $5, NOW())
RETURNING *;

-- name: HasAdminUser :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE role_id = 1 AND deleted_at IS NULL 
    LIMIT 1
) as has_admin;