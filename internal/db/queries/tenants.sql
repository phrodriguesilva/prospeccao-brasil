-- name: GetTenant :one
SELECT * FROM tenants WHERE id = $1;

-- name: CreateTenant :one
INSERT INTO tenants (id, name, cnpj, plan)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateTenant :one
UPDATE tenants
SET name = $2, cnpj = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListTenantsByActive :many
SELECT * FROM tenants WHERE active = $1 AND deleted_at IS NULL ORDER BY created_at;
