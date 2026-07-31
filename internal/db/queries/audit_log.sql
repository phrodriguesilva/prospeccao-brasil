-- name: CreateAuditLog :one
INSERT INTO audit_log (id, tenant_id, user_id, action, entity_type, entity_id, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAuditLogByTenant :many
SELECT * FROM audit_log
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListAuditLogByEntity :many
SELECT * FROM audit_log
WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
ORDER BY created_at DESC;
