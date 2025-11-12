-- name: CreateOrder :one
INSERT INTO orders (
    created_by, status, notes
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetOrder :one
SELECT * FROM orders
WHERE id = $1 LIMIT 1;

-- name: ListOrders :many
SELECT * FROM orders
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListOrdersByUser :many
SELECT * FROM orders
WHERE created_by = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateOrderStatus :one
UPDATE orders
SET 
    status = $2,
    submitted_at = CASE WHEN $2 = 'submitted' THEN NOW() ELSE submitted_at END
WHERE id = $1
RETURNING *;

-- name: DeleteOrder :exec
DELETE FROM orders WHERE id = $1;

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id, product_id, requested_qty, unit, note
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetOrderItems :many
SELECT * FROM order_items
WHERE order_id = $1
ORDER BY id;

-- name: UpdateOrderItem :one
UPDATE order_items
SET 
    requested_qty = COALESCE($2, requested_qty),
    unit = COALESCE($3, unit),
    note = COALESCE($4, note)
WHERE id = $1
RETURNING *;

-- name: DeleteOrderItem :exec
DELETE FROM order_items WHERE id = $1;