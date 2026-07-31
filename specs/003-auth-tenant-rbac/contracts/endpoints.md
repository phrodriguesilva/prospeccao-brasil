# HTTP Endpoint Contracts: Auth + Tenant + RBAC Middleware

**Date**: 2026-07-31
**Spec**: [spec.md](./spec.md)

## Public Endpoints (no auth)

These endpoints do NOT require authentication. They are mounted on the chi
router without the session validation or RBAC middleware.

### GET /login

**Purpose**: Display the login form.

**Response**: 200 OK, HTML (`internal/template/login.html`).

**Form fields**:
- `email` (text input, required, type="email")
- `password` (password input, required)
- CSRF token (hidden input, future hardening -- not in MVP)

### POST /login

**Purpose**: Authenticate user with email + password.

**Request**:
- Content-Type: `application/x-www-form-urlencoded`
- Body: `email=<email>&password=<password>`

**Responses**:
- 302 Found: Redirect to `/2fa/setup` (if `totp_enabled=false`), `/2fa/verify`
  (if `totp_enabled=true`), or `/admin` (if 2FA not required for this role --
  not applicable in MVP since admin requires 2FA).
  - Set-Cookie: `session=<token>; HttpOnly; SameSite=Strict; [Secure]; Path=/; Max-Age=86400`
- 401 Unauthorized: HTML with error "Email ou senha invalidos" (generic, no
  user enumeration).
- 429 Too Many Requests: Rate limit exceeded (5 attempts per 15 seconds per
  IP or per email).
- 403 Forbidden: Account locked. HTML with error "Conta bloqueada, tente
  novamente em 15 minutos".

**Rate limiting**: 5 attempts per 15 seconds per IP + per email (R6).

**Audit**: Logs `login_attempt` (always), `login_success` (on success),
`login_failed` (on failure), `account_locked` (on lockout).

### GET /healthz

**Purpose**: Liveness probe (no auth, per Constitution V).

**Response**: 200 OK, `{"status":"ok"}`.

## 2FA Endpoints (session-pending, not fully authenticated)

These endpoints require a "pending" session (password validated but 2FA not
yet completed). The pending session is stored in a short-lived cookie
(separate from the main session cookie) or in a context value.

### GET /2fa/setup

**Purpose**: Display QR code for TOTP enrollment.

**Response**: 200 OK, HTML (`internal/template/totp_setup.html`) with:
- QR code image (base64 PNG, `data:image/png;base64,...`)
- TOTP code input form (POST to `/2fa/setup`)

### POST /2fa/setup

**Purpose**: Verify TOTP code and complete enrollment.

**Request**:
- Content-Type: `application/x-www-form-urlencoded`
- Body: `code=<6-digit TOTP code>`

**Responses**:
- 302 Found: Redirect to `/admin`. Set-Cookie with main session token.
  `totp_enabled` set to true in DB.
- 401 Unauthorized: HTML with error "Codigo TOTP invalido".

### GET /2fa/verify

**Purpose**: Display TOTP code input for existing 2FA users.

**Response**: 200 OK, HTML (`internal/template/totp_verify.html`) with:
- TOTP code input form (POST to `/2fa/verify`)

### POST /2fa/verify

**Purpose**: Verify TOTP code for existing 2FA users.

**Request**:
- Content-Type: `application/x-www-form-urlencoded`
- Body: `code=<6-digit TOTP code>`

**Responses**:
- 302 Found: Redirect to `/admin`. Set-Cookie with main session token.
- 401 Unauthorized: HTML with error "Codigo TOTP invalido".

**Rate limiting**: 5 attempts per 15 seconds per IP.

## Protected Endpoints (auth + RBAC required)

These endpoints require a valid session cookie and the appropriate RBAC role.
They are mounted on the chi router with the session validation + RBAC
middleware chain.

### POST /logout

**Purpose**: Revoke session and clear cookie.

**Request**: Session cookie required.

**Response**: 302 Found, redirect to `/login`.
- Set-Cookie: `session=; HttpOnly; SameSite=Strict; [Secure]; Path=/; Max-Age=0`
  (clears the cookie).

**Audit**: Logs `logout` and `session_revoked`.

### GET /admin

**Purpose**: Placeholder for the internal system dashboard (SPEC-05).

**RBAC**: admin only.

**Response**: 200 OK, HTML with "Authenticated as <user.email>" (placeholder).

## Middleware Chain

```text
chi router
├── Public group (no middleware)
│   ├── GET  /healthz
│   ├── GET  /login
│   ├── POST /login
│   ├── GET  /2fa/setup    (pending-session middleware)
│   ├── POST /2fa/setup    (pending-session middleware)
│   ├── GET  /2fa/verify   (pending-session middleware)
│   └── POST /2fa/verify   (pending-session middleware)
└── Protected group (session validation + RBAC)
    ├── POST /logout
    └── GET  /admin        (RBAC: admin)
```

## Cookie Specification

| Attribute | Value | Notes |
|-----------|-------|-------|
| Name | `session` | Main session cookie |
| Value | 256-bit random token (base64) | Raw token; DB stores SHA-256 hash |
| HttpOnly | true | JS cannot read (XSS protection) |
| SameSite | Strict | Not sent on cross-site requests (CSRF protection) |
| Secure | conditional | true if `APP_BASE_URL` starts with `https://` |
| Path | `/` | Sent on all paths |
| Max-Age | 86400 (24 hours) | Matches `expires_at` in DB |

**Pending session cookie** (for 2FA flow):
| Attribute | Value | Notes |
|-----------|-------|-------|
| Name | `pending_session` | Short-lived, for 2FA enrollment/verification |
| Value | user_id + tenant_id (signed) | HMAC-signed to prevent tampering |
| Max-Age | 300 (5 minutes) | Short window for 2FA completion |
