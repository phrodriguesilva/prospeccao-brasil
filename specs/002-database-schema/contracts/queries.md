# Query Contracts: Database Schema & Migrations

**Date**: 2026-07-31
**Spec**: [spec.md](./spec.md)
**Data Model**: [data-model.md](./data-model.md)

## Convention

Every tenant-scoped query MUST include `AND tenant_id = $X` in the WHERE
clause. This is enforced by:
1. Code review (PR gate)
2. ast-grep rule `go-missing-tenant-filter` (structural scan)
3. Integration tests (cross-tenant access returns empty)

Queries are written as sqlc SQL in `internal/db/queries/<entity>.sql`. sqlc
generates typed Go functions in `internal/db/<entity>.sql.go`.

## tenants

```sql
-- name: GetTenant :one
SELECT * FROM tenants WHERE id = $1;

-- name: CreateTenant :one
INSERT INTO tenants (id, name, cnpj, plan) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateTenant :one
UPDATE tenants SET name = $2, cnpj = $3, updated_at = now() WHERE id = $1 RETURNING *;
```

## users

```sql
-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (id, tenant_id, email, full_name, role, password_hash)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdateUserTOTP :one
UPDATE users SET totp_secret = $2, totp_enabled = $3, updated_at = now()
WHERE id = $1 AND tenant_id = $4 RETURNING *;

-- name: UpdateUserLoginAttempts :one
UPDATE users SET failed_login_attempts = $2, locked_at = $3, updated_at = now()
WHERE id = $1 AND tenant_id = $4 RETURNING *;

-- name: ListUsersByTenant :many
SELECT * FROM users WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC;
```

## sessions

```sql
-- name: GetSessionByTokenHash :one
SELECT * FROM sessions WHERE token_hash = $1 AND tenant_id = $2 AND revoked_at IS NULL;

-- name: CreateSession :one
INSERT INTO sessions (id, tenant_id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: RevokeSession :exec
UPDATE sessions SET revoked_at = now()
WHERE id = $1 AND tenant_id = $2;

-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions WHERE expires_at < now();
```

## properties

```sql
-- name: GetPropertyByID :one
SELECT * FROM properties WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: ListPropertiesByTenant :many
SELECT * FROM properties WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListPropertiesByStatus :many
SELECT * FROM properties WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateProperty :one
INSERT INTO properties (id, tenant_id, title, address, city, state, zip_code, price, status, type, bedrooms, bathrooms, area_sqm, description, photos)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING *;

-- name: UpdateProperty :one
UPDATE properties SET title = $3, address = $4, city = $5, state = $6, zip_code = $7, price = $8, status = $9, type = $10, bedrooms = $11, bathrooms = $12, area_sqm = $13, description = $14, photos = $15, updated_at = now()
WHERE id = $1 AND tenant_id = $2 RETURNING *;

-- name: SoftDeleteProperty :exec
UPDATE properties SET deleted_at = now() WHERE id = $1 AND tenant_id = $2;
```

## clients

```sql
-- name: GetClientByID :one
SELECT * FROM clients WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: ListClientsByTenant :many
SELECT * FROM clients WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListClientsByStatus :many
SELECT * FROM clients WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateClient :one
INSERT INTO clients (id, tenant_id, name, email, phone, cpf_cnpj, address, budget, preferences, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: UpdateClient :one
UPDATE clients SET name = $3, email = $4, phone = $5, cpf_cnpj = $6, address = $7, budget = $8, preferences = $9, status = $10, updated_at = now()
WHERE id = $1 AND tenant_id = $2 RETURNING *;

-- name: SoftDeleteClient :exec
UPDATE clients SET deleted_at = now() WHERE id = $1 AND tenant_id = $2;
```

## prospections

```sql
-- name: GetProspectByID :one
SELECT * FROM prospections WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: ListProspectsByTenant :many
SELECT * FROM prospections WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListProspectsByClient :many
SELECT * FROM prospections WHERE client_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListProspectsByProperty :many
SELECT * FROM prospections WHERE property_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListProspectsByStatus :many
SELECT * FROM prospections WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateProspect :one
INSERT INTO prospections (id, tenant_id, client_id, property_id, status, notes, contact_date, next_action_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: UpdateProspect :one
UPDATE prospections SET status = $3, notes = $4, next_action_date = $5, updated_at = now()
WHERE id = $1 AND tenant_id = $2 RETURNING *;

-- name: SoftDeleteProspect :exec
UPDATE prospections SET deleted_at = now() WHERE id = $1 AND tenant_id = $2;
```

## contacts

```sql
-- name: GetContactByID :one
SELECT * FROM contacts WHERE id = $1 AND tenant_id = $2;

-- name: ListContactsByClient :many
SELECT * FROM contacts WHERE client_id = $1 AND tenant_id = $2
ORDER BY contacted_at DESC;

-- name: ListContactsByProspect :many
SELECT * FROM contacts WHERE prospect_id = $1 AND tenant_id = $2
ORDER BY contacted_at DESC;

-- name: CreateContact :one
INSERT INTO contacts (id, tenant_id, client_id, prospect_id, channel, direction, subject, body, contacted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;
```

## audit_log

```sql
-- name: CreateAuditLog :one
INSERT INTO audit_log (id, tenant_id, user_id, action, entity_type, entity_id, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: ListAuditLogByTenant :many
SELECT * FROM audit_log WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListAuditLogByEntity :many
SELECT * FROM audit_log WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
ORDER BY created_at DESC;
```

**Append-only enforcement**: No `UpdateAuditLog` or `DeleteAuditLog` queries
exist (FR-009). The ast-grep rule and integration tests verify this.
