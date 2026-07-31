-- name: GetContactByID :one
SELECT * FROM contacts
WHERE id = $1 AND tenant_id = $2;

-- name: ListContactsByClient :many
SELECT * FROM contacts
WHERE client_id = $1 AND tenant_id = $2
ORDER BY contacted_at DESC;

-- name: ListContactsByProspect :many
SELECT * FROM contacts
WHERE prospect_id = $1 AND tenant_id = $2
ORDER BY contacted_at DESC;

-- name: CreateContact :one
INSERT INTO contacts (
    id, tenant_id, client_id, prospect_id,
    channel, direction, subject, body, contacted_at
)
VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8, $9
)
RETURNING *;
