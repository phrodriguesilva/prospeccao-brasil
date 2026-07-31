-- name: GetClientByID :one
SELECT * FROM clients
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: ListClientsByTenant :many
SELECT * FROM clients
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListClientsByStatus :many
SELECT * FROM clients
WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateClient :one
INSERT INTO clients (
    id, tenant_id, name, email, phone, cpf_cnpj, address,
    budget, preferences, status
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10
)
RETURNING *;

-- name: UpdateClient :one
UPDATE clients
SET
    name = $3, email = $4, phone = $5, cpf_cnpj = $6, address = $7,
    budget = $8, preferences = $9, status = $10, updated_at = now()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: SoftDeleteClient :exec
UPDATE clients
SET deleted_at = now()
WHERE id = $1 AND tenant_id = $2;

-- name: ListClientsFiltered :many
SELECT * FROM clients
WHERE tenant_id = $1 AND deleted_at IS NULL
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR name ILIKE '%' || $3 || '%' OR email ILIKE '%' || $3 || '%')
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountClientsFiltered :one
SELECT COUNT(*) FROM clients
WHERE tenant_id = $1 AND deleted_at IS NULL
  AND ($2 = '' OR status = $2)
  AND ($3 = '' OR name ILIKE '%' || $3 || '%' OR email ILIKE '%' || $3 || '%');
