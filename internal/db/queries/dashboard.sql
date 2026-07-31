-- name: CountPropertiesByTenant :one
SELECT COUNT(*) FROM properties
WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: CountClientsByTenant :one
SELECT COUNT(*) FROM clients
WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: CountProspectsByTenant :one
SELECT COUNT(*) FROM prospections
WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: CountProspectsByStatus :many
SELECT status, COUNT(*) as count FROM prospections
WHERE tenant_id = $1 AND deleted_at IS NULL
GROUP BY status;

-- name: ListRecentProspectsWithDetails :many
SELECT
    p.id, p.tenant_id, p.client_id, p.property_id, p.status,
    p.notes, p.contact_date, p.next_action_date,
    p.created_at, p.updated_at, p.deleted_at,
    c.name AS client_name,
    pr.title AS property_title
FROM prospections p
JOIN clients c ON p.client_id = c.id
JOIN properties pr ON p.property_id = pr.id
WHERE p.tenant_id = $1 AND p.deleted_at IS NULL
ORDER BY p.created_at DESC
LIMIT $2;
