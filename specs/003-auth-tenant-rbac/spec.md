# Feature Specification (Slim): Auth + Tenant + RBAC Middleware

**Feature Branch**: `003-auth-tenant-rbac`

**Created**: 2026-07-31

**Status**: Draft

**Template**: slim (infrastructure spec per AGENTS.md). SPEC-03 delivers
engineering value (auth encanamento, middleware, security foundation) rather
than direct product UI. The login form is minimal (no design system -- that
arrives in SPEC-04). The MVP is single-admin (Constitution principle VII),
but the middleware, RBAC, and 2FA encanamento ship now for future multi-user.

**Input**: User description: "SPEC-03: Auth + Tenant + RBAC Middleware"

## Overview

This spec implements the authentication system for Prospecção Brasil:
server-side session cookies (HttpOnly + SameSite=Strict + Secure),
email/password login with bcrypt, 2FA TOTP for the admin user, RBAC
middleware enforcing role-based access on every handler, tenant_id middleware
that sets the tenant context from the session, and instant revocation via
Postgres session deletion. The MVP is single-tenant, single-admin (Luiz
Claudio), but the RBAC roles (admin, corretor, assistente, financeiro) and
2FA encanamento ship now so future multi-user requires only UI changes, not
backend rework.

This is the security foundation for all user-facing features. Without auth,
no handler can be trusted. Every downstream spec (institutional site,
internal system) depends on authenticated, authorized requests with
tenant-scoped data access. The ast-grep rule `go-handler-missing-auth.yml`
starts firing once handlers exist in SPEC-04/SPEC-05.

## Context

**Canonical sources:**
- [Constitution](../../.specify/memory/constitution.md) principle II
  (Security-First LGPD), principle VII (Simplicity for Single-User)
- [AGENTS.md](../../AGENTS.md) -- session cookie flags, RBAC roles,
  ast-grep rules, chi router, sqlc, slog
- Reference: `pragmaos/specs/003-auth-rbac/spec.md` (adapted for
  commercial real-estate prospecting domain -- roles are admin/corretor/
  assistente/financeiro, not socio/advogado/estagiario/recepcao; 2FA
  required for admin only in MVP, not all roles; dashboard is /admin not
  /dashboard)

**Dependencies**: SPEC-02 (Database Schema & Migrations) -- complete. The
`tenants`, `users`, and `sessions` tables exist with sqlc-generated typed Go
queries in `internal/db/`.

**Gate to run**: SPEC-02 is complete (initial schema migrated, sqlc
generating typed Go in `internal/db/`, CI green).

## Goals

1. Implement session-based auth with HttpOnly + SameSite=Strict + Secure
   cookies, bcrypt password hashing, and instant revocation via Postgres
   session deletion.
2. Implement 2FA TOTP (RFC 6238) enrollment and verification for the admin
   user, with AES-256-GCM encryption of the `totp_secret` at rest.
3. Implement RBAC middleware with 4 roles (admin, corretor, assistente,
   financeiro) enforcing role-based access on every handler. MVP is
   single-admin but the middleware ships now.
4. Implement tenant_id middleware that sets the tenant context from the
   session, attaching `user_id`, `tenant_id`, and `role` to the request
   context for downstream handlers.
5. Implement login/logout flows with rate limiting (in-memory token bucket)
   and account lockout after 5 failed attempts.
6. Implement session validation middleware that checks `revoked_at IS NULL`,
   `expires_at > now()`, `users.deleted_at IS NULL`, and
   `tenants.deleted_at IS NULL` on every protected request.

## Non-Goals

The following are explicitly deferred to later specs:

- Design system / styled login form -> SPEC-04 (Institutional Site &
  Design System). SPEC-03 ships a minimal unstyled login form.
- Institutional site pages (Home, Quem somos, etc.) -> SPEC-04.
- Internal system (properties, clients, prospecting CRUD) -> SPEC-05.
- Multi-user UI (user management, role assignment screens) -> future spec.
  The middleware and schema support it; the UI does not.
- JWT / API tokens for mobile -> future spec. Session cookies are the MVP.
- Password reset via email -> future spec. The MVP has a single admin;
  password reset is manual (direct DB update or CLI command).
- Key rotation for AES-256-GCM -> future hardening spec. MVP uses a
  single `ENCRYPTION_KEY` env var.
- argon2id password hashing -> future hardening spec. MVP uses bcrypt.

## Requirements

These requirements are the verifiable acceptance criteria. Each is
non-negotiable.

- **FR-001**: Login flow: POST `/login` with email + password validates
  against bcrypt hash in `users` table. On success, creates a session in
  Postgres and sets a cookie (HttpOnly + SameSite=Strict + Secure). On
  failure, shows generic "Email ou senha invalidos" (no user enumeration).
  Verify: `curl -X POST /login -d "email=admin@prospeccao.com.br&password=wrong"`
  returns 401 with generic error; correct credentials return 302 redirect
  to `/admin` with Set-Cookie header.
- **FR-002**: Account lockout: after 5 failed login attempts,
  `failed_login_attempts` is incremented on each failure; at 5,
  `locked_at` is set and the account is locked for 15 minutes. Verify:
  5 failed POST `/login` attempts result in `locked_at` being set in
  `users` table; subsequent login attempts return "Conta bloqueada, tente
  novamente em 15 minutos" regardless of password correctness.
- **FR-003**: 2FA TOTP enrollment: after password validation, if
  `totp_enabled` is false, redirect to `/2fa/setup` which displays a QR
  code (RFC 6238, 30-second period, 6 digits). User scans QR, enters a
  TOTP code; if valid, `totp_enabled` is set to true and `totp_secret` is
  stored encrypted (AES-256-GCM). Verify: POST `/2fa/setup` with a valid
  TOTP code sets `totp_enabled=true` in `users` table; `totp_secret` is
  not plaintext in DB.
- **FR-004**: 2FA TOTP verification: if `totp_enabled` is true, after
  password validation, redirect to `/2fa/verify`. User enters TOTP code;
  if valid, session is created and user is redirected to `/admin`. If
  invalid, access is denied. TOTP validation allows +/- 1 time step (30
  seconds) for clock drift. Verify: POST `/2fa/verify` with invalid code
  returns 401; valid code returns 302 redirect to `/admin`.
- **FR-005**: Logout: POST `/logout` revokes the session (sets
  `revoked_at` in `sessions` table), clears the cookie, and redirects to
  `/login`. Any subsequent request with the old session cookie returns 401
  and redirects to `/login`. Verify: after logout, GET `/admin` with old
  cookie returns 302 redirect to `/login`.
- **FR-006**: Session validation middleware: every request to a protected
  handler passes through middleware that reads the session cookie, looks
  up the session by `token_hash` (SHA-256 of token), checks
  `revoked_at IS NULL` and `expires_at > now()`. If invalid, returns 401
  and redirects to `/login`. Valid sessions attach `user_id`, `tenant_id`,
  `role` to request context. Verify: GET `/admin` without cookie returns
  302 to `/login`; GET `/admin` with valid cookie returns 200.
- **FR-007**: Tenant validation: session validation middleware also
  checks `users.deleted_at IS NULL` and `tenants.deleted_at IS NULL`. If
  the user or tenant is soft-deleted, the session is rejected. Verify:
  soft-delete the user in DB, GET `/admin` with valid cookie returns 302
  to `/login`.
- **FR-008**: RBAC middleware: every handler is wrapped in RBAC
  middleware. The middleware checks the user's role against the required
  role for the handler. If insufficient, returns 403 Forbidden. Roles:
  admin (highest), corretor, assistente, financeiro. MVP: only admin
  exists, but the middleware enforces the hierarchy. Verify: create a
  user with role `assistente`, login, GET an admin-only endpoint, verify
  403.
- **FR-009**: Rate limiting: in-memory token bucket rate limiter on
  `/login` and `/2fa/verify` endpoints (per IP + per email). 5 attempts
  per 15 seconds per IP; 5 attempts per 15 seconds per email. Verify:
  6 rapid POST `/login` attempts from the same IP return 429 Too Many
  Requests.
- **FR-010**: Cookie security: cookie is HttpOnly + SameSite=Strict +
  Secure. Secure flag is disabled in dev (HTTP localhost) and enabled in
  production (HTTPS). Controlled by `APP_BASE_URL` env var (if starts
  with `https://`, Secure=true). Verify: in dev (APP_BASE_URL=http://...),
  Set-Cookie header has `Secure` absent; in production
  (APP_BASE_URL=https://...), Set-Cookie header has `Secure`.
- **FR-011**: Audit logging: all auth events (login_attempt,
  login_success, login_failed, logout, session_revoked, totp_enrolled,
  account_locked, rbac_denied) are logged via slog and recorded in
  `audit_log` table. Verify: after a successful login, `audit_log` table
  has a row with `action='login_success'` and the correct `user_id` and
  `tenant_id`.
- **FR-012**: ast-grep rule `go-handler-missing-auth.yml` fires on any
  handler that does not access `r.Context().Value("user")` (except
  health/ready/live/ping endpoints). Verify: create a handler without
  auth context access, run `ast-grep scan`, verify the rule fires.
- **FR-013**: 85% test coverage on new Go code (excluding sqlc-generated
  `internal/db` and `cmd/prospeccao`). Verify: `go test -coverprofile=
  coverage.out ./... && go tool cover -func=coverage.out | grep total`
  shows >= 85% on non-excluded packages.
- **FR-014**: Seed data: a CLI command or migration seed creates the
  initial tenant and admin user (Luiz Claudio) with a bcrypt-hashed
  password. The admin user has `totp_enabled=false` (enrolled on first
  login). Verify: `make seed` creates 1 tenant and 1 admin user in the
  DB; the admin user's password hash is bcrypt (starts with `$2a$`).

## Constraints

1. Single Go binary + Postgres. No Docker/K8s for dev (Constitution III).
2. No secrets in repo; `ENCRYPTION_KEY` in `.env.local` (gitignored,
   gitleaks-enforced). Used for AES-256-GCM encryption of `totp_secret`.
3. No emojis anywhere -- code, UI, comments, docs, commits (project rule).
4. No conventional-commit prefixes (rejected by CI).
5. Structured logging via slog (Constitution V); no `fmt.Println` in
   non-main code (ast-grep rule `go-slog-fmt.yml`).
6. No new migrations (SPEC-02 schema is sufficient -- `users`, `sessions`,
   `tenants` tables already have all needed columns).
7. Cookie Secure flag conditional on `APP_BASE_URL` (dev=HTTP, prod=HTTPS).
8. RBAC roles: admin, corretor, assistente, financeiro (from SPEC-02
   CHECK constraint). No other roles allowed.
9. 2FA TOTP required for admin (MVP single-admin). Future: required for
   all roles touching client data (LGPD).
10. Session token: 256-bit random, SHA-256 hashed before storage. Token
    never stored in plaintext. Cookie value is the raw token.
11. Context propagation: `ctx context.Context` as first param on all
    DB/HTTP work (AGENTS.md).
12. Error wrapping: `fmt.Errorf("...: %w", err)` -- no bare `return err`
    (ast-grep rule `go-bare-error.yml`).

## Definition of Done

Auth + Tenant + RBAC Middleware is done when ALL of the following are
verifiable:

| # | Acceptance Criterion | Verification Command | FR | Status |
|---|----------------------|----------------------|----|--------|
| 1 | Login with correct credentials creates session + cookie | `curl -X POST /login -d "email=admin@prospeccao.com.br&password=test123" -v` shows 302 + Set-Cookie | FR-001 | [ ] |
| 2 | Login with wrong credentials returns generic error | `curl -X POST /login -d "email=admin@prospeccao.com.br&password=wrong"` returns 401 | FR-001 | [ ] |
| 3 | Account locks after 5 failed attempts | 5 failed POSTs, then `psql -c "SELECT locked_at FROM users WHERE email='admin@prospeccao.com.br'"` shows non-null | FR-002 | [ ] |
| 4 | 2FA enrollment sets totp_enabled + encrypts secret | POST `/2fa/setup` with valid TOTP, then `psql -c "SELECT totp_enabled, totp_secret FROM users"` shows true + non-plaintext | FR-003 | [ ] |
| 5 | 2FA verification rejects invalid code | POST `/2fa/verify` with invalid code returns 401 | FR-004 | [ ] |
| 6 | Logout revokes session + clears cookie | POST `/logout`, then GET `/admin` returns 302 to `/login` | FR-005 | [ ] |
| 7 | Session validation rejects no cookie | GET `/admin` without cookie returns 302 to `/login` | FR-006 | [ ] |
| 8 | Session validation rejects expired session | Expire session in DB, GET `/admin` returns 302 to `/login` | FR-006 | [ ] |
| 9 | Tenant validation rejects soft-deleted user | Soft-delete user in DB, GET `/admin` returns 302 to `/login` | FR-007 | [ ] |
| 10 | RBAC denies insufficient role | Login as `assistente`, GET admin-only endpoint, verify 403 | FR-008 | [ ] |
| 11 | Rate limiting returns 429 after 5 attempts | 6 rapid POST `/login` from same IP, 6th returns 429 | FR-009 | [ ] |
| 12 | Cookie Secure flag conditional on APP_BASE_URL | Dev: Set-Cookie lacks Secure; Prod: Set-Cookie has Secure | FR-010 | [ ] |
| 13 | Audit log records auth events | After login, `psql -c "SELECT action FROM audit_log WHERE action='login_success'"` returns row | FR-011 | [ ] |
| 14 | ast-grep rule fires on handler without auth | Create handler without `r.Context().Value("user")`, `ast-grep scan` fires | FR-012 | [ ] |
| 15 | 85% test coverage on new code | `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out \| grep total` >= 85% | FR-013 | [ ] |
| 16 | Seed creates tenant + admin user | `make seed` then `psql -c "SELECT count(*) FROM tenants"` = 1, `psql -c "SELECT count(*) FROM users"` = 1 | FR-014 | [ ] |

**Spec is ready for `/speckit-plan` when all rows are checked.**

## Data Contract

Generated by `speckit-tekimax-security-data-contract` hook (mandatory for
SPEC-03: has auth, sessions, PII (password_hash, totp_secret), HTTP
endpoints). Adapted for Go + Postgres + sqlc (no Zod/TypeScript).

### Sources

| Name | Origin | Trust | Schema Location | PII? |
|------|--------|-------|-----------------|------|
| users | DB table (SPEC-02) | vetted | `migrations/` + sqlc | email, full_name, password_hash, totp_secret |
| sessions | DB table (SPEC-02) | vetted | `migrations/` + sqlc | token_hash |
| tenants | DB table (SPEC-02) | vetted | `migrations/` + sqlc | name, cnpj |
| audit_log | DB table (SPEC-02) | vetted | `migrations/` + sqlc | user_id, action, entity_id |

### Schema Definition

No new migrations. SPEC-02 schema is sufficient:
- `users`: id, tenant_id, email, full_name, role, password_hash,
  totp_secret, totp_enabled, failed_login_attempts, locked_at, active,
  created_at, updated_at, deleted_at.
- `sessions`: id, tenant_id, user_id, token_hash, expires_at, revoked_at,
  created_at.
- `tenants`: id, name, cnpj, plan, active, created_at, updated_at,
  deleted_at.
- `audit_log`: id, tenant_id, user_id, action, entity_type, entity_id,
  metadata, created_at.

New sqlc queries needed (added to `internal/db/queries/`):
- `users.sql`: `GetUserForAuth` (by email, includes password_hash, locked_at,
  failed_login_attempts, totp_secret, totp_enabled -- does NOT filter
  deleted_at so middleware can check).
- `sessions.sql`: `RevokeSessionByID` (by id + tenant_id), `GetSessionWithUser`
  (joins sessions + users + tenants for validation middleware).

### PII Handling

| Field | Table | Strategy | Implementation |
|-------|-------|----------|----------------|
| email | users | stored (login lookup) | Used for authentication; not encrypted (needed for lookup). Protected by tenant_id isolation + RBAC. |
| password_hash | users | hash (bcrypt) | One-way bcrypt hash. Never plaintext, never reversible. Cost factor 10 (default). |
| totp_secret | users | field-encrypt (AES-256-GCM) | Encrypted at rest with `ENCRYPTION_KEY` env var. Decrypted only in 2FA flow. Never logged. |
| token_hash | sessions | hash (SHA-256) | SHA-256 of session token. One-way. Token never stored. |
| full_name | users | stored | Used for display. Protected by tenant_id isolation + RBAC. |

### Bias Audit

- **Segments that must be represented**: N/A for SPEC-03 (auth middleware,
  no filtering or scoring logic). Bias audit applies to prospecting features.
- **Known bias risks**: None at auth level. Auth is neutral -- email +
  password + TOTP. RBAC is role-based, not demographic.
- **Mitigation**: [DEFERRED: SPEC-05] Bias audit will run when prospecting
  features consume auth.

### Drift Monitoring

- **Baseline**: N/A for SPEC-03 (no data pipeline or scoring at auth level).
- **Threshold**: N/A.
- **Detection**: [DEFERRED: future ops spec] Auth events are logged to
  `audit_log` for retrospective analysis.

### Retention

| Data | TTL | Deletion Path | LGPD Basis |
|------|-----|---------------|------------|
| sessions | 30 days after `expires_at` | `DeleteExpiredSessions` query (SPEC-02) + cron (future ops spec) | Art. 15 |
| audit_log (auth events) | 5 years | archival job (future ops spec); no in-app DELETE | Art. 16 |
| users | indefinite (while tenant active) | soft-delete (`deleted_at`); hard delete on tenant termination | Art. 15 |
| tenants | indefinite | soft-delete; hard delete on service termination | Art. 15 |

## Security / Threat Model

**Generated**: 2026-07-31
**Scope**: SPEC-03 -- Auth + Tenant + RBAC Middleware
**Attack surface**: HTTP endpoints (`/login`, `/logout`, `/2fa/setup`,
`/2fa/verify`, `/admin`), session cookies, bcrypt password hashes,
AES-GCM encrypted TOTP secrets, Postgres session table, rate limiter.

### Threats

| ID | Category | Threat | Severity | Likelihood | Mitigation | Status |
|----|----------|--------|----------|------------|------------|--------|
| T1 | Spoofing | Attacker steals a session cookie via XSS and impersonates the user | High | Low | Cookie is HttpOnly (JS cannot read it) + SameSite=Strict (not sent on cross-site requests) + Secure (only over HTTPS in prod). No inline scripts in templates (CSP future hardening). | Mitigated |
| T2 | Spoofing | Attacker brute-forces the login password | High | Medium | bcrypt password hashing (cost 10, ~100ms per hash). Rate limiting: 5 attempts per 15 seconds per IP + per email. Account lockout after 5 failed attempts (15-minute lock). | Mitigated |
| T3 | Spoofing | Attacker bypasses 2FA by replaying an old TOTP code | Medium | Low | TOTP codes are time-based (30-second window). `pquerna/otp` validates with +/- 1 step window. Replayed codes are rejected (each code valid only once per step). | Mitigated |
| T4 | Tampering | Attacker forges a session cookie with a known token_hash | Critical | Low | Session token is 256-bit random (unguessable). Cookie stores the raw token; DB stores SHA-256 hash. Attacker cannot forge a cookie that matches a DB hash without knowing the original token. | Mitigated |
| T5 | Tampering | Attacker modifies the RBAC middleware to skip role checks | High | Low | RBAC middleware is applied at the router level (chi middleware chain). Handlers cannot bypass it. ast-grep rule `go-handler-missing-auth` catches handlers without auth context access. | Mitigated |
| T6 | Repudiation | User denies performing an action; no audit trail | High | Medium | All auth events (login, logout, 2FA, RBAC denial) are logged to `audit_log` with `user_id`, `tenant_id`, `action`, `created_at`. Append-only (FR-009 from SPEC-02). | Mitigated |
| T7 | Information Disclosure | Attacker reads `totp_secret` from DB dump and bypasses 2FA | Critical | Low | `totp_secret` is encrypted at rest with AES-256-GCM. `ENCRYPTION_KEY` is in `.env.local` (gitignored, gitleaks-enforced). DB dump does not contain the key. | Mitigated |
| T8 | Information Disclosure | Login error message reveals whether email exists ("user not found" vs "wrong password") | Medium | Medium | Generic error message "Email ou senha invalidos" for all failures. No user enumeration. `failed_login_attempts` is incremented only when the email exists (to avoid enumeration via timing). | Mitigated |
| T9 | Information Disclosure | Attacker reads session token from URL or referrer | Medium | Low | Session token is in a cookie (not URL). SameSite=Strict prevents cross-site referrer leakage. No token in query params. | Mitigated |
| T10 | Denial of Service | Attacker floods `/login` with requests, exhausting Postgres connections | Medium | Medium | Rate limiting (5 per 15 seconds per IP). pgxpool connection pooling with max connections. Session lookup is a single indexed query by `token_hash` (O(log n)). | Mitigated |
| T11 | Elevation of Privilege | User with role `assistente` accesses an admin-only endpoint | High | Low | RBAC middleware checks role on every handler. admin > corretor > assistente > financeiro hierarchy. Insufficient role returns 403. ast-grep catches handlers without RBAC. | Mitigated |
| T12 | Elevation of Privilege | User from tenant A accesses tenant B's data via session | Critical | Low | Session validation middleware attaches `tenant_id` from the session (not from user input). All downstream queries use `tenant_id` from context. Cross-tenant access requires forging the session token (see T4). | Mitigated |

## Assumptions

1. The MVP has a single tenant and a single admin user (Luiz Claudio).
   The seed command (`make seed`) creates this tenant and admin. The
   admin's initial password is set via `ADMIN_PASSWORD` env var (or a
   default that must be changed on first login -- deferred to future
   hardening).
2. 2FA is required for the admin user. The admin enrolls on first login
   (redirected to `/2fa/setup`). Future: 2FA required for all roles
   touching client data (corretor, assistente).
3. The `ENCRYPTION_KEY` env var (32 bytes, base64-encoded) is used for
   AES-256-GCM encryption of `totp_secret`. It is already in `.env.example`
   from SPEC-01. The production key must be different from the dev key.
4. The chi router (`github.com/go-chi/chi/v5`) is used for HTTP routing
   and middleware chaining. It is stdlib-compatible and lightweight.
5. The `pquerna/otp/totp` library is used for TOTP (RFC 6238). It handles
   QR code generation (PNG) and TOTP code validation.
6. Rate limiting is in-memory (token bucket). This is sufficient for a
   single-binary deployment. For multi-instance deployment (future), a
   Redis-backed rate limiter would be needed -- deferred.
7. The login form is minimal (unstyled HTML). The design system arrives in
   SPEC-04. SPEC-03 focuses on the auth logic, not the UI.
8. The `/admin` route is a placeholder (returns a simple "authenticated"
   message). The real internal system UI arrives in SPEC-05.
