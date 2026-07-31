-- name: GetProspectByID :one
SELECT * FROM prospections
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: ListProspectsByTenant :many
SELECT * FROM prospections
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListProspectsByClient :many
SELECT * FROM prospections
WHERE client_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListProspectsByProperty :many
SELECT * FROM prospections
WHERE property_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListProspectsByStatus :many
SELECT * FROM prospections
WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateProspect :one
INSERT INTO prospections (
    id, tenant_id, client_id, property_id, status,
    notes, contact_date, next_action_date
)
VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8
)
RETURNING *;

-- name: UpdateProspect :one
UPDATE prospections
SET
    status = $3, notes = $4, next_action_date = $5, updated_at = now()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: SoftDeleteProspect :exec
UPDATE prospections
SET deleted_at = now()
WHERE id = $1 AND tenant_id = $2;

-- name: ListProspectsFiltered :many
SELECT * FROM prospections
WHERE tenant_id = $1 AND deleted_at IS NULL
  AND ($2 = '' OR status = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountProspectsFiltered :one
SELECT COUNT(*) FROM prospections
WHERE tenant_id = $1 AND deleted_at IS NULL
  AND ($2 = '' OR status = $2);

-- name: ListProspectsByClientWithDetails :many
SELECT
    p.id, p.tenant_id, p.client_id, p.property_id, p.status,
    p.notes, p.contact_date, p.next_action_date,
    p.created_at, p.updated_at, p.deleted_at,
    pr.title AS property_title
FROM prospections p
JOIN properties pr ON p.property_id = pr.id
WHERE p.client_id = $1 AND p.tenant_id = $2 AND p.deleted_at IS NULL
ORDER BY p.created_at DESC;
