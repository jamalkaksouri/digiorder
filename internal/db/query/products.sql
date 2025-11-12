-- name: CreateProduct :one
INSERT INTO products (
    name, brand, dosage_form_id, strength, unit, category_id, description
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetProduct :one
SELECT * FROM products
WHERE id = $1 LIMIT 1;

-- name: ListProducts :many
SELECT * FROM products
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateProduct :one
UPDATE products
SET 
    name = COALESCE($2, name),
    brand = COALESCE($3, brand),
    dosage_form_id = COALESCE($4, dosage_form_id),
    strength = COALESCE($5, strength),
    unit = COALESCE($6, unit),
    category_id = COALESCE($7, category_id),
    description = COALESCE($8, description)
WHERE id = $1
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;

-- name: SearchProducts :many
SELECT * FROM products
WHERE 
    name ILIKE '%' || $1 || '%' 
    OR brand ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;