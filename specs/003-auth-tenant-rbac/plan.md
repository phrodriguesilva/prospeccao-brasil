# Implementation Plan: Auth + Tenant + RBAC Middleware

**Branch**: `003-auth-tenant-rbac` | **Date**: 2026-07-31 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-auth-tenant-rbac/spec.md`

## Summary

Implement server-side session cookie authentication (HttpOnly + SameSite=Strict +
Secure), email/password login with bcrypt, 2FA TOTP (RFC 6238) for the admin
user, RBAC middleware enforcing role-based access on every handler, tenant_id
middleware, and instant revocation via Postgres session deletion. Uses chi
router for middleware chaining, `pquerna/otp` for TOTP, AES-GCM (stdlib) for
totp_secret encryption, and in-memory token bucket rate limiting. No new
migrations needed (SPEC-02 schema is sufficient). MVP is single-admin but the
RBAC encanamento ships now.

## Technical Context

**Language/Version**: Go 1.26 (go.mod: `go 1.26.0`)

**Primary Dependencies** (new):
- `github.com/go-chi/chi/v5` -- HTTP router with middleware support
- `github.com/pquerna/otp/totp` -- TOTP RFC 6238 (QR code + validation)
- `golang.org/x/crypto/bcrypt` -- Password hashing
- `github.com/google/uuid` -- UUID generation (already from SPEC-02)

**Primary Dependencies** (existing):
- `github.com/jackc/pgx/v5` + `pgxpool` -- Postgres driver (SPEC-02)
- `internal/db` -- sqlc-generated typed queries (SPEC-02)

**Storage**: PostgreSQL 16 (existing `users`, `sessions`, `tenants`, `audit_log`
tables from SPEC-02). No new migrations.

**Testing**: `go test -race -coverprofile` with integration tests against local
Postgres (DATABASE_URL). 85% coverage gate.

**Target Platform**: Linux server (production), macOS (local dev). Single Go
binary.

**Project Type**: Web service (server-rendered monolith, Go + HTMX + Postgres).

**Performance Goals**: Session validation middleware < 5ms overhead per request
(single indexed Postgres lookup by token_hash). Login flow < 10s end-to-end
(including 2FA).

**Constraints**:
- No secrets in repo (gitleaks). ENCRYPTION_KEY in .env.local.
- No emojis (project rule).
- slog for logging (no fmt.Println in non-main code, ast-grep enforced).
- No new migrations (SPEC-02 schema is sufficient).
- Cookie Secure flag: disabled in dev (HTTP), enabled in prod (HTTPS).
  Controlled by `APP_BASE_URL` env var.
- Context propagation: ctx as first param on all DB/HTTP work.
- ast-grep rule `go-handler-missing-auth` enforces every handler accesses
  `r.Context().Value("user")` (except health/ready/live/ping).

**Scale/Scope**: MVP single-tenant, single-admin. Auth is per-request,
stateless middleware + stateful session in Postgres. ~10 source files, ~15
test functions.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec-Driven | PASS | spec.md exists at specs/003-auth-tenant-rbac/, plan.md being generated |
| II. Security-First (LGPD) | PASS | Session cookie HttpOnly+SameSite=Strict+Secure, bcrypt, AES-GCM totp_secret, RBAC on every handler, rate limiting, instant revocation, audit_log. Data contract + threat model in spec.md. |
| III. Single-Binary & Tooling | PASS | Single Go binary, no npm/Docker. New deps via go.mod (chi, otp, bcrypt). All gates via make check. |
| IV. Test-First & 85% Coverage | PASS | FR-013 requires 85% coverage. Tests cover happy path + failure modes (wrong password, invalid TOTP, expired session, locked account, RBAC denial, rate limit). |
| V. Observability & slog | PASS | All auth events logged via slog (login_attempt, login_success, login_failed, logout, session_revoked, totp_enrolled, account_locked, rbac_denied). |
| VI. Forward-Only Migrations | PASS | No new migrations. SPEC-02 schema is sufficient. |
| VII. Simplicity for Single-User | PASS | MVP is single-admin. RBAC middleware ships as encanamento but only admin role exists. No multi-user UI. Minimal login form (no design system). |

**Gate Result**: PASS. No violations. No complexity tracking needed.

## Project Structure

### Documentation (this feature)

```text
specs/003-auth-tenant-rbac/
├── plan.md              # This file
├── research.md          # Phase 0: design decisions
├── data-model.md        # Phase 1: auth state machine + entity model
├── quickstart.md        # Phase 1: validation scenarios
└── contracts/
    └── endpoints.md     # Phase 1: HTTP endpoint contracts
```

### Source Code (repository root)

```text
internal/
├── auth/
│   ├── auth.go          # Auth service: login, logout, session management
│   ├── auth_test.go     # Integration tests for auth service
│   ├── password.go      # bcrypt hash/verify helpers
│   ├── password_test.go # Unit tests for password hashing
│   ├── totp.go          # TOTP generation, validation, QR code, AES-GCM encrypt/decrypt
│   ├── totp_test.go     # Unit tests for TOTP
│   ├── session.go       # Session token generation, cookie management, session validation
│   ├── session_test.go  # Unit + integration tests for sessions
│   ├── rbac.go          # RBAC role definitions, middleware, role hierarchy
│   ├── rbac_test.go     # Unit tests for RBAC
│   ├── ratelimit.go     # Token bucket rate limiter (in-memory, per IP + per email)
│   └── ratelimit_test.go # Unit tests for rate limiting
├── handler/
│   ├── auth_handler.go  # HTTP handlers: /login, /logout, /2fa/setup, /2fa/verify
│   ├── auth_handler_test.go # Handler integration tests
│   ├── middleware.go    # Session validation middleware + RBAC middleware + tenant context
│   └── middleware_test.go # Middleware tests
├── db/                  # SPEC-02: sqlc-generated queries (updated with new queries)
│   ├── queries/
│   │   ├── users.sql    # Updated: GetUserForAuth
│   │   ├── sessions.sql # Updated: RevokeSessionByID, GetSessionWithUser
│   │   └── audit_log.sql # Existing: CreateAuditLog
│   └── *.sql.go         # Generated (gitignored)
├── template/
│   ├── login.html       # Minimal login form (no design system)
│   ├── totp_setup.html  # QR code + TOTP code input
│   └── totp_verify.html # TOTP code input
cmd/prospeccao/
    └── main.go          # Updated: chi router, middleware chain, DB pool, template parsing
scripts/
    └── seed.go          # Seed script: creates initial tenant + admin user
```

**Structure Decision**: Single Go binary layout. Auth logic in `internal/auth/`
(separable, testable). HTTP handlers in `internal/handler/`. Middleware in
`internal/handler/middleware.go`. Templates in `internal/template/` (minimal,
unstyled). Seed script in `scripts/seed.go` (run via `make seed`). chi router
in `cmd/prospeccao/main.go` wires the middleware chain.

## Complexity Tracking

> No constitution violations. Table empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| -- | -- | -- |
