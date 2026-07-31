-- name: GetSessionByTokenHash :one
SELECT * FROM sessions
WHERE token_hash = $1 AND tenant_id = $2 AND revoked_at IS NULL;

-- name: CreateSession :one
INSERT INTO sessions (id, tenant_id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: RevokeSession :exec
UPDATE sessions
SET revoked_at = now()
WHERE id = $1 AND tenant_id = $2;

-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions WHERE expires_at < now();

-- name: GetSessionWithUser :one
-- Joins sessions + users + tenants for validation middleware.
-- Returns session + user role + soft-delete flags for middleware checks.
SELECT
    s.id AS session_id,
    s.tenant_id,
    s.user_id,
    s.token_hash,
    s.expires_at,
    s.revoked_at,
    s.created_at,
    u.email,
    u.full_name,
    u.role,
    u.totp_enabled,
    u.deleted_at AS user_deleted_at,
    t.deleted_at AS tenant_deleted_at
FROM sessions s
JOIN users u ON s.user_id = u.id
JOIN tenants t ON s.tenant_id = t.id
WHERE s.token_hash = $1
  AND s.tenant_id = $2
  AND s.revoked_at IS NULL
  AND s.expires_at > now();

-- name: RevokeSessionByID :exec
UPDATE sessions
SET revoked_at = now()
WHERE id = $1 AND tenant_id = $2;
