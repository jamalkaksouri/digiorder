-- name: CreateBarcode :one
INSERT INTO product_barcodes (
    product_id, barcode, barcode_type
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetBarcode :one
SELECT * FROM product_barcodes
WHERE id = $1 LIMIT 1;

-- name: GetBarcodesByProduct :many
SELECT * FROM product_barcodes
WHERE product_id = $1
ORDER BY created_at DESC;

-- name: GetProductByBarcode :one
SELECT p.* FROM products p
INNER JOIN product_barcodes pb ON p.id = pb.product_id
WHERE pb.barcode = $1
LIMIT 1;

-- name: UpdateBarcode :one
UPDATE product_barcodes
SET 
    barcode = COALESCE($2, barcode),
    barcode_type = COALESCE($3, barcode_type)
WHERE id = $1
RETURNING *;

-- name: DeleteBarcode :exec
DELETE FROM product_barcodes WHERE id = $1;

-- name: SearchBarcodes :many
SELECT * FROM product_barcodes
WHERE barcode ILIKE '%' || $1 || '%'
ORDER BY barcode
LIMIT $2 OFFSET $3;