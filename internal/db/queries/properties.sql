-- name: GetPropertyByID :one
SELECT * FROM properties
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: ListPropertiesByTenant :many
SELECT * FROM properties
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListPropertiesByStatus :many
SELECT * FROM properties
WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateProperty :one
INSERT INTO properties (
    id, tenant_id, title, address, city, state, zip_code, price,
    status, type, bedrooms, bathrooms, area_sqm, description, photos
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14, $15
)
RETURNING *;

-- name: UpdateProperty :one
UPDATE properties
SET
    title = $3, address = $4, city = $5, state = $6, zip_code = $7,
    price = $8, status = $9, type = $10, bedrooms = $11, bathrooms = $12,
    area_sqm = $13, description = $14, photos = $15, updated_at = now()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: SoftDeleteProperty :exec
UPDATE properties
SET deleted_at = now()
WHERE id = $1 AND tenant_id = $2;

-- name: ListPropertiesFiltered :many
SELECT * FROM properties
WHERE tenant_id = $1 AND deleted_at IS NULL
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR type = $3)
  AND ($4 = '' OR title ILIKE '%' || $4 || '%' OR city ILIKE '%' || $4 || '%')
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountPropertiesFiltered :one
SELECT COUNT(*) FROM properties
WHERE tenant_id = $1 AND deleted_at IS NULL
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR type = $3)
  AND ($4 = '' OR title ILIKE '%' || $4 || '%' OR city ILIKE '%' || $4 || '%');
