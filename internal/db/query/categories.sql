-- name: ListCategories :many
SELECT * FROM categories
ORDER BY name;

-- name: GetCategory :one
SELECT * FROM categories
WHERE id = $1 LIMIT 1;

-- name: CreateCategory :one
INSERT INTO categories (name) 
VALUES ($1)
RETURNING *;

-- name: ListDosageForms :many
SELECT * FROM dosage_forms
ORDER BY name;

-- name: GetDosageForm :one
SELECT * FROM dosage_forms
WHERE id = $1 LIMIT 1;

-- name: CreateDosageForm :one
INSERT INTO dosage_forms (name) 
VALUES ($1)
RETURNING *;

-- name: ListRoles :many
SELECT * FROM roles
ORDER BY name;

-- name: GetRole :one
SELECT * FROM roles
WHERE id = $1 LIMIT 1;