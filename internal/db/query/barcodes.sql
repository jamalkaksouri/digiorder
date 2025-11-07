-- name: CreateProductBarcode :one
INSERT INTO product_barcodes (
    product_id, barcode, barcode_type
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetProductByBarcode :one
SELECT p.* FROM products p
JOIN product_barcodes pb ON p.id = pb.product_id
WHERE pb.barcode = $1 LIMIT 1;

-- name: GetProductBarcodes :many
SELECT * FROM product_barcodes
WHERE product_id = $1;

-- name: DeleteProductBarcode :exec
DELETE FROM product_barcodes WHERE id = $1;

-- name: SearchProductsByBarcode :many
SELECT p.* FROM products p
JOIN product_barcodes pb ON p.id = pb.product_id
WHERE pb.barcode ILIKE '%' || $1 || '%'
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;