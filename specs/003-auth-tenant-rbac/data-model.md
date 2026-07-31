# Data Model: Auth + Tenant + RBAC Middleware

**Date**: 2026-07-31
**Spec**: [spec.md](./spec.md)
**Research**: [research.md](./research.md)

## Entity Overview

No new tables. SPEC-02 schema is sufficient. This document describes the
auth-specific data flows and state transitions over the existing tables.

## Auth State Machine

### Login Flow

```
[User enters email + password]
        |
        v
[GetUserForAuth by email] --not found--> [Generic error: "Email ou senha invalidos"]
        |
     found
        |
        v
[Check locked_at] --locked & <15min--> [Error: "Conta bloqueada, tente novamente em 15 minutos"]
        |
     not locked
        |
        v
[bcrypt.Compare] --mismatch--> [Increment failed_login_attempts]
        |                          |--attempts >= 5--> [Set locked_at, log account_locked]
        |                          |
        match                     v
        |                   [Generic error: "Email ou senha invalidos"]
        v
[Reset failed_login_attempts to 0]
        |
        v
[Check totp_enabled] --false--> [Redirect to /2fa/setup (enrollment)]
        |
     true
        |
        v
[Redirect to /2fa/verify] --valid TOTP--> [Create session, set cookie, redirect to /admin]
        |
     invalid TOTP
        |
        v
[Error: "Codigo TOTP invalido"]
```

### Session Lifecycle

```
[CreateSession] --> [Active: revoked_at IS NULL, expires_at > now()]
                        |
                        |--POST /logout--> [RevokeSession: set revoked_at]
                        |--expires_at < now()--> [Expired: auto-cleanup by DeleteExpiredSessions]
                        |--user soft-deleted--> [Invalid: middleware rejects]
                        |--tenant soft-deleted--> [Invalid: middleware rejects]
                        |
                        v
                    [Inactive: revoked_at IS NOT NULL OR expires_at < now()]
```

### 2FA Enrollment Flow

```
[totp_enabled = false]
        |
        v
[Generate TOTP secret] --> [Encrypt with AES-256-GCM] --> [Store encrypted secret in users.totp_secret]
        |
        v
[Generate QR code (PNG, base64)] --> [Display to user]
        |
        v
[User scans QR, enters TOTP code]
        |
        v
[totp.Validate(code, decrypted_secret)] --valid--> [Set totp_enabled = true, redirect to /admin]
        |
     invalid
        |
        v
[Error: "Codigo TOTP invalido"]
```

## RBAC Role Hierarchy

```
admin (highest)
  |-- full access to all features
  |-- can manage users, properties, clients, prospections, contacts
  |-- can view audit_log
  |
corretor
  |-- manages properties and prospections
  |-- can CRUD properties, clients, prospections, contacts
  |-- cannot manage users
  |-- cannot view audit_log
  |
assistente
  |-- supports data entry and client contacts
  |-- can CRUD clients, contacts
  |-- can read properties, prospections
  |-- cannot create/update properties or prospections
  |-- cannot manage users
  |
financeiro (lowest)
  |-- billing and financial reports
  |-- can read clients, prospections (for billing context)
  |-- cannot CRUD any entity
  |-- cannot manage users
```

**MVP**: Only `admin` role exists (single-admin). The hierarchy ships as
encanamento for future multi-user.

## Context Values

The session validation middleware attaches the following to the request
context (`r.Context()`):

| Key | Type | Source | Used By |
|-----|------|--------|---------|
| `user` | `*db.User` | `GetSessionWithUser` query | All protected handlers |
| `tenant_id` | `pgtype.UUID` | `session.tenant_id` | All tenant-scoped queries |
| `user_id` | `pgtype.UUID` | `session.user_id` | Audit logging |
| `role` | `string` | `user.role` | RBAC middleware |

Handlers access these via `r.Context().Value("user").(*db.User)` etc.
The ast-grep rule `go-handler-missing-auth` checks for `r.Context().Value("user")`
in every handler (except health endpoints).

## New sqlc Queries

Added to `internal/db/queries/` (existing files updated):

### users.sql (new queries)

```sql
-- name: GetUserForAuth :one
-- Does NOT filter deleted_at so middleware can check separately.
SELECT * FROM users WHERE email = $1 AND tenant_id = $2;

-- name: ResetFailedLoginAttempts :exec
UPDATE users
SET failed_login_attempts = 0, locked_at = NULL, updated_at = now()
WHERE id = $1 AND tenant_id = $2;
```

### sessions.sql (new queries)

```sql
-- name: GetSessionWithUser :one
-- Joins sessions + users + tenants for validation middleware.
-- Returns session, user, and tenant in one query (reduces round trips).
SELECT
    s.*,
    u.deleted_at AS user_deleted_at,
    u.role,
    t.deleted_at AS tenant_deleted_at
FROM sessions s
JOIN users u ON s.user_id = u.id
JOIN tenants t ON s.tenant_id = t.id
WHERE s.token_hash = $1
  AND s.tenant_id = $2
  AND s.revoked_at IS NULL
  AND s.expires_at > now();
```

Note: `GetSessionWithUser` returns a custom struct (not a table model).
sqlc generates a `GetSessionWithUserRow` struct with the joined columns.
