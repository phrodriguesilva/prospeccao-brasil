-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (id, tenant_id, email, full_name, role, password_hash)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateUserTOTP :one
UPDATE users
SET totp_secret = $2, totp_enabled = $3, updated_at = now()
WHERE id = $1 AND tenant_id = $4
RETURNING *;

-- name: UpdateUserLoginAttempts :one
UPDATE users
SET failed_login_attempts = $2, locked_at = $3, updated_at = now()
WHERE id = $1 AND tenant_id = $4
RETURNING *;

-- name: ListUsersByTenant :many
SELECT * FROM users
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetUserForAuth :one
-- Does NOT filter deleted_at so middleware can check separately.
SELECT * FROM users
WHERE email = $1 AND tenant_id = $2;

-- name: ResetFailedLoginAttempts :exec
UPDATE users
SET failed_login_attempts = 0, locked_at = NULL, updated_at = now()
WHERE id = $1 AND tenant_id = $2;
