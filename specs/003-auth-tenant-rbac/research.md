# Research: Auth + Tenant + RBAC Middleware

**Date**: 2026-07-31
**Spec**: [spec.md](./spec.md)

## Research Tasks

### R1: chi vs net/http vs gorilla/mux for HTTP routing

**Decision**: chi (`github.com/go-chi/chi/v5`).

**Rationale**:
- chi is lightweight, stdlib-compatible (uses `http.Handler` interface), and
  has first-class middleware chaining via `r.With(middleware)`.
- Middleware groups: `r.Group(func(r chi.Router) { r.Use(authmw, rbacmw) })`
  makes it easy to separate public routes (login, health) from protected
  routes (admin).
- chi is the most popular Go router for server-rendered monoliths (no SPA).
- No external dependencies beyond chi itself.

**Alternatives considered**:
- `net/http` (stdlib): no middleware chaining out of the box. Would require
  manual chaining. Rejected because: verbose, error-prone, no route groups.
- `gorilla/mux`: popular but in maintenance mode (archived). Rejected.
- `gin`: performance-focused but adds its own context type (not stdlib
  compatible). Rejected because: adds complexity, not needed for a monolith.

### R2: bcrypt vs argon2id for password hashing

**Decision**: bcrypt (cost 10) for MVP. argon2id deferred to future hardening.

**Rationale**:
- bcrypt is battle-tested, widely supported, and simple (`golang.org/x/crypto/
  bcrypt`). Cost 10 = ~100ms per hash (acceptable for login).
- argon2id is the modern recommendation (memory-hard, resistant to GPU
  attacks) but requires tuning parameters (memory, iterations, parallelism).
  For a single-admin MVP, bcrypt is sufficient.
- Migration path: when argon2id is added, re-hash on next login (check
  bcrypt first, if valid, re-hash with argon2id and update).

**Alternatives considered**:
- argon2id: better security but more complex. Deferred.
- scrypt: good but less common in Go ecosystems. Rejected.

### R3: TOTP library -- pquerna/otp vs go-authy/totp

**Decision**: `github.com/pquerna/otp/totp`.

**Rationale**:
- `pquerna/otp` is the most popular Go TOTP library. Implements RFC 6238
  (TOTP) and RFC 4226 (HOTP).
- Generates QR codes as PNG (base64-encoded) for easy embedding in HTML.
- Validates TOTP codes with configurable skew (time step window).
- Well-maintained, no external dependencies.

**Alternatives considered**:
- `go-authy/totp`: less popular, less documentation. Rejected.
- Manual implementation: TOTP is just HMAC-SHA1 over a time counter. But QR
  code generation is non-trivial. Rejected because: reinventing the wheel.

### R4: AES-256-GCM encryption for totp_secret -- stdlib vs external

**Decision**: Go stdlib `crypto/aes` + `crypto/cipher` (GCM mode).

**Rationale**:
- Go stdlib has excellent crypto support. AES-GCM is the recommended
  authenticated encryption mode (NIST SP 800-38D).
- No external dependencies. `crypto/aes` + `crypto/cipher.NewGCM` is all
  that's needed.
- Key: 32 bytes (AES-256) from `ENCRYPTION_KEY` env var (base64-decoded).
- Nonce: 12 bytes, randomly generated per encryption (stored prepended to
  ciphertext).

**Alternatives considered**:
- `tink-go`: Google's crypto library. Overkill for a single field.
  Rejected.
- pgcrypto: DB-side encryption. Rejected (SPEC-02 research R2: app-layer
  encryption keeps keys out of DB).

### R5: Session token generation -- crypto/rand vs uuid

**Decision**: 256-bit random token via `crypto/rand`, SHA-256 hashed for
storage.

**Rationale**:
- 256-bit random token from `crypto/rand` is cryptographically secure and
  has enough entropy to be unguessable.
- SHA-256 hash for storage: if the DB is compromised, the attacker cannot
  reconstruct the original token from the hash.
- The raw token is stored in the cookie (client-side); the hash is stored
  in `sessions.token_hash` (DB-side).
- UUID v4 is 122 bits of entropy -- sufficient but less than 256 bits.
  Using `crypto/rand` gives us maximum entropy with minimal code.

**Alternatives considered**:
- UUID v4: 122 bits, sufficient but less entropy. Rejected because: we want
  256-bit tokens for session security.
- JWT: stateless, no DB lookup needed. Rejected because: no instant
  revocation (JWT is valid until expiry). Constitution requires instant
  revocation.

### R6: Rate limiting -- in-memory vs Redis

**Decision**: In-memory token bucket (`golang.org/x/time/rate`).

**Rationale**:
- MVP is a single binary on a single server. In-memory rate limiting is
  sufficient.
- `golang.org/x/time/rate` is a stdlib-adjacent library (golang.org/x).
  Token bucket algorithm: `rate.NewLimiter(rate.Every(3*time.Second), 5)`
  allows 5 bursts, refilling 1 token per 3 seconds.
- Per-IP and per-email limiters: two maps keyed by IP and email,
  respectively. Each entry is a `*rate.Limiter`.
- For multi-instance deployment (future), a Redis-backed rate limiter
  would be needed. Deferred.

**Alternatives considered**:
- Redis-backed: needed for multi-instance but adds a dependency. Deferred.
- No rate limiting: rejected because: brute-force attacks are a real threat
  (T2 in threat model).

### R7: Cookie Secure flag -- always vs conditional

**Decision**: Conditional on `APP_BASE_URL` env var.

**Rationale**:
- In dev, the app runs on `http://localhost:8080`. If the cookie has
  `Secure=true`, the browser will not send it over HTTP -- the user cannot
  log in locally.
- In production, the app runs on `https://prospeccao.com.br`. The cookie
  MUST have `Secure=true` to prevent the browser from sending it over HTTP
  (downgrade attacks).
- `APP_BASE_URL` env var: if it starts with `https://`, `Secure=true`;
  otherwise `Secure=false`.
- This is a common pattern in Go web apps.

**Alternatives considered**:
- Always `Secure=true`: breaks local dev. Rejected.
- Always `Secure=false`: insecure in production. Rejected.
- Separate `COOKIE_SECURE` env var: more explicit but another env var to
  manage. `APP_BASE_URL` already conveys the protocol.

### R8: Seed data -- migration vs CLI script

**Decision**: CLI script (`scripts/seed.go`), run via `make seed`.

**Rationale**:
- Migrations should be schema-only (DDL), not data (DML). Mixing data into
  migrations makes them non-idempotent and harder to test.
- A CLI script is explicit: `make seed` creates the tenant + admin user.
  It can be re-run safely (checks if tenant/user already exist).
- The admin password is read from `ADMIN_PASSWORD` env var (not hardcoded).
  If not set, a default is used with a warning to change it.
- The script uses the same `internal/db` queries as the app (no raw SQL).

**Alternatives considered**:
- Migration-based seed: rejected because: DML in migrations is an anti-
  pattern; non-idempotent; hard to re-run.
- Manual SQL: rejected because: error-prone, not reproducible.
