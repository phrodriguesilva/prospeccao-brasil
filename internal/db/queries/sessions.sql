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
