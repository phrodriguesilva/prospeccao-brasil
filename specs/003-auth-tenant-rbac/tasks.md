# Tasks: Auth + Tenant + RBAC Middleware

**Input**: Design documents from `/specs/003-auth-tenant-rbac/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/endpoints.md, quickstart.md

**Tests**: Tests are MANDATORY for SPEC-03 (FR-013 requires 85% coverage). Every source file has a corresponding `_test.go` file.

**Organization**: Tasks are grouped by phase (Setup, Auth Core, HTTP Layer, Seed, Polish). Tasks marked `[P]` can run in parallel (different files, no dependencies).

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)

## Path Conventions

- Go monolith: `internal/` for packages, `cmd/prospeccao/` for entry point.
- Paths are relative to repository root.

---

## Phase 1: Setup (T001-T003)

**Purpose**: Add Go dependencies and update sqlc queries.

**Checkpoint**: All deps installed. `make sqlc` succeeds. `go build ./...` succeeds.

- [ ] T001 Add Go dependencies: `github.com/go-chi/chi/v5`, `github.com/pquerna/otp`, `golang.org/x/crypto/bcrypt`, `golang.org/x/time/rate`. Run `go get github.com/go-chi/chi/v5@latest && go get github.com/pquerna/otp@latest && go get golang.org/x/crypto/bcrypt@latest && go get golang.org/x/time/rate@latest`. Verify `go.mod` updated. (research.md R1, R3, R6)
- [ ] T002 Add new sqlc queries to `internal/db/queries/users.sql`: `GetUserForAuth` (by email + tenant_id, does NOT filter deleted_at -- middleware checks separately), `ResetFailedLoginAttempts` (sets failed_login_attempts=0, locked_at=NULL). Add to `internal/db/queries/sessions.sql`: `GetSessionWithUser` (joins sessions + users + tenants, checks revoked_at IS NULL + expires_at > now(), returns user.deleted_at, user.role, tenant.deleted_at), `RevokeSessionByID` (by id + tenant_id, sets revoked_at=now()). Follow data-model.md. (FR-001, FR-006, FR-007, data-model.md)
- [ ] T003 Run `make sqlc` and verify generated Go code. Verify `go build ./internal/db/...` succeeds. Verify `GetUserForAuth`, `ResetFailedLoginAttempts`, `GetSessionWithUser`, `RevokeSessionByID` are generated. (FR-001)

## Phase 2: Auth Core (T004-T013)

**Purpose**: Implement auth service, password hashing, TOTP, session management, RBAC, rate limiting.

**Checkpoint**: All auth core packages compile. Unit tests pass. 85% coverage on `internal/auth/`.

- [ ] T004 [P] Create `internal/auth/password.go` -- `HashPassword(password string) (string, error)` using bcrypt with cost 10. `VerifyPassword(hashedPassword, password string) error` using bcrypt.CompareHashAndPassword. (FR-001, research.md R2)
- [ ] T005 [P] Create `internal/auth/password_test.go` -- test HashPassword returns a `$2a$` hash; test VerifyPassword succeeds on correct password, fails on wrong password; test HashPassword is non-deterministic (different salts). (FR-001, FR-013)
- [ ] T006 [P] Create `internal/auth/totp.go` -- `GenerateTOTPSecret(email string) (secret string, qrPNG []byte, err error)` using pquerna/otp/totp.Generate with issuer "Prospeccao Brasil". `ValidateTOTP(code, encryptedSecret string, key []byte) bool` -- decrypts secret with AES-256-GCM, validates with totp.Validate (skew=1 for clock drift). `EncryptSecret(secret string, key []byte) (string, error)` -- AES-GCM encrypt, return base64. `DecryptSecret(encrypted string, key []byte) (string, error)` -- base64 decode, AES-GCM decrypt. (FR-003, FR-004, research.md R3, R4)
- [ ] T007 [P] Create `internal/auth/totp_test.go` -- test GenerateTOTPSecret returns non-empty secret + QR PNG; test EncryptSecret + DecryptSecret round-trip; test ValidateTOTP with a valid code (generate secret, generate code, validate); test ValidateTOTP with invalid code returns false; test DecryptSecret with wrong key returns error. (FR-003, FR-004, FR-013)
- [ ] T008 [P] Create `internal/auth/session.go` -- `GenerateSessionToken() (raw string, hash string)` -- 256-bit random via crypto/rand, SHA-256 hash for storage. `SessionCookie(name, token string, secure bool) *http.Cookie` -- HttpOnly + SameSite=Strict + Secure + Path=/ + MaxAge=86400. `ClearCookie(name string, secure bool) *http.Cookie` -- MaxAge=0. `PendingSessionCookie(userID, tenantID string, key []byte) *http.Cookie` -- HMAC-signed, MaxAge=300. `VerifyPendingSession(cookie string, key []byte) (userID, tenantID string, err error)`. (FR-001, FR-005, FR-010, research.md R5, R7)
- [ ] T009 [P] Create `internal/auth/session_test.go` -- test GenerateSessionToken returns 256-bit token + different hash; test SessionCookie has HttpOnly + SameSite=Strict; test SessionCookie Secure flag (secure=true vs false); test ClearCookie has MaxAge=0; test PendingSessionCookie + VerifyPendingSession round-trip; test VerifyPendingSession with tampered cookie returns error. (FR-001, FR-005, FR-010, FR-013)
- [ ] T010 [P] Create `internal/auth/rbac.go` -- `Role` type with constants: `RoleAdmin`, `RoleCorretor`, `RoleAssistente`, `RoleFinanceiro`. `RoleLevel(role string) int` -- admin=4, corretor=3, assistente=2, financeiro=1. `RequireRole(required string) func(http.Handler) http.Handler` -- middleware that checks `r.Context().Value("role")` >= required level, returns 403 if insufficient. `ContextKey` type for context values ("user", "tenant_id", "user_id", "role"). (FR-008, data-model.md)
- [ ] T011 [P] Create `internal/auth/rbac_test.go` -- test RoleLevel for all 4 roles; test RequireRole with admin accessing admin-only (200); test RequireRole with assistente accessing admin-only (403); test RequireRole with no role in context (403). (FR-008, FR-013)
- [ ] T012 [P] Create `internal/auth/ratelimit.go` -- `RateLimiter` struct with per-IP and per-email `*rate.Limiter` maps (sync.Map). `NewRateLimiter() *RateLimiter`. `Allow(ip string) bool` -- 5 per 15 seconds. `AllowEmail(email string) bool` -- 5 per 15 seconds. `Middleware` function for chi that checks IP rate limit and returns 429 if exceeded. (FR-009, research.md R6)
- [ ] T013 [P] Create `internal/auth/ratelimit_test.go` -- test Allow returns true for first 5, false on 6th; test AllowEmail independently; test concurrent access (sync.Map safety); test rate recovery after window. (FR-009, FR-013)

## Phase 3: Auth Service + HTTP Layer (T014-T022)

**Purpose**: Implement the auth service (login/logout logic), HTTP handlers, and middleware.

**Checkpoint**: All handlers compile. Integration tests pass. 85% coverage on `internal/handler/`.

- [ ] T014 Create `internal/auth/auth.go` -- `Service` struct with `*db.Queries`, `*pgxpool.Pool`, `[]byte` (encryption key), `*RateLimiter`, `*slog.Logger`. `Login(ctx, email, password, tenantID string) (*LoginResult, error)` -- fetches user via GetUserForAuth, checks locked_at, verifies bcrypt, increments failed_login_attempts on failure, resets on success, logs audit events. Returns LoginResult{User, Need2FASetup, Need2FAVerify}. `Complete2FA(ctx, userID, totpCode string) (bool, error)` -- validates TOTP, sets totp_enabled if enrollment. `CreateSession(ctx, userID, tenantID string) (rawToken, hash string, err error)` -- generates token, inserts session. `Logout(ctx, sessionID, tenantID string) error` -- revokes session. (FR-001, FR-002, FR-003, FR-004, FR-005, FR-011)
- [ ] T015 Create `internal/auth/auth_test.go` -- integration tests against real Postgres: test Login with correct credentials (returns LoginResult with Need2FASetup=true for new user); test Login with wrong password (increments failed_login_attempts); test Login with locked account (returns locked error); test Login resets failed_login_attempts on success; test Complete2FA with valid code (sets totp_enabled); test Complete2FA with invalid code (returns false); test CreateSession inserts row in sessions; test Logout sets revoked_at. Use setupTestDB helper from SPEC-02 (truncate between tests). (FR-001, FR-002, FR-003, FR-004, FR-005, FR-013)
- [ ] T016 Create `internal/handler/middleware.go` -- `SessionValidation(queries *db.Queries, log *slog.Logger) func(http.Handler) http.Handler` -- reads session cookie, looks up via GetSessionWithUser, checks user.deleted_at IS NULL + tenant.deleted_at IS NULL, attaches user/tenant_id/user_id/role to context, returns 401 + redirect to /login if invalid. `TenantContext` helper to extract tenant_id from context. `UserContext` helper to extract user from context. (FR-006, FR-007, data-model.md)
- [ ] T017 Create `internal/handler/middleware_test.go` -- integration tests: test SessionValidation with no cookie (302 to /login); test with valid cookie (handler called, context has user); test with expired session (302 to /login); test with revoked session (302 to /login); test with soft-deleted user (302 to /login); test with soft-deleted tenant (302 to /login). Use httptest.NewRecorder + chi router. (FR-006, FR-007, FR-013)
- [ ] T018 Create `internal/handler/auth_handler.go` -- `AuthHandler` struct with `*auth.Service`, `*template.Template`, `*slog.Logger`, `secureCookies bool`. `LoginGET` handler (renders login.html). `LoginPOST` handler (parses form, rate-limits, calls Service.Login, sets pending cookie, redirects to /2fa/setup or /2fa/verify). `TotpSetupGET` (generates QR if not enrolled, renders totp_setup.html). `TotpSetupPOST` (validates TOTP, completes enrollment, creates session, sets cookie, redirects to /admin). `TotpVerifyGET` (renders totp_verify.html). `TotpVerifyPOST` (validates TOTP, creates session, sets cookie, redirects to /admin). `LogoutPOST` (revokes session, clears cookie, redirects to /login). `AdminGET` (placeholder: "Authenticated as <email>"). All handlers access `r.Context().Value("user")` (except LoginGET/LoginPOST which are public). (FR-001, FR-003, FR-004, FR-005, FR-012)
- [ ] T019 Create `internal/handler/auth_handler_test.go` -- integration tests: test LoginGET returns 200 with form; test LoginPOST with correct creds returns 302 to /2fa/setup; test LoginPOST with wrong creds returns 401; test LoginPOST with locked account returns 403; test TotpSetupPOST with valid code sets totp_enabled + creates session; test TotpVerifyPOST with invalid code returns 401; test LogoutPOST clears cookie + redirects; test AdminGET without auth redirects to /login; test AdminGET with auth returns 200. Use httptest + chi router + real Postgres. (FR-001, FR-003, FR-004, FR-005, FR-006, FR-013)
- [ ] T020 [P] Create `internal/template/login.html` -- minimal unstyled HTML form with email + password inputs, POST to /login, error message display ({{.Error}}). No design system (SPEC-04). (FR-001)
- [ ] T021 [P] Create `internal/template/totp_setup.html` -- minimal HTML with QR code image ({{.QRCode}}), TOTP code input form, POST to /2fa/setup. (FR-003)
- [ ] T022 [P] Create `internal/template/totp_verify.html` -- minimal HTML with TOTP code input form, POST to /2fa/verify. (FR-004)

## Phase 4: Router + Seed (T023-T026)

**Purpose**: Wire the chi router in main.go, create the seed script, update Makefile.

**Checkpoint**: `make dev` starts the server. `make seed` creates tenant + admin. All endpoints respond.

- [ ] T023 Update `cmd/prospeccao/main.go` -- replace net/http mux with chi router. Wire middleware chain: public group (/healthz, /login, /2fa/*) and protected group (/logout, /admin with SessionValidation + RequireRole(admin)). Parse templates from internal/template/. Initialize pgxpool, auth.Service, handler.AuthHandler. Read APP_BASE_URL for cookie Secure flag. Read ENCRYPTION_KEY for AES-GCM. Use slog for logging. (FR-001, FR-006, FR-008, FR-010, contracts/endpoints.md)
- [ ] T024 Create `scripts/seed.go` -- standalone Go program (package main). Connects to DATABASE_URL. Creates 1 tenant ("Prospeccao Brasil", plan="free"). Creates 1 admin user (email from ADMIN_EMAIL env or default "admin@prospeccao.com.br", password from ADMIN_PASSWORD env, bcrypt-hashed, role="admin", totp_enabled=false). Idempotent: checks if tenant/user already exist before inserting. Logs via slog. (FR-014, research.md R8)
- [ ] T025 Add `seed` target to Makefile: `go run ./scripts/seed.go`. Add `dev` target if not present: `go run ./cmd/prospeccao`. Verify `make seed` creates tenant + admin. Verify `make dev` starts server on :8080. (FR-014)
- [ ] T026 Update `.env.example` -- add `ADMIN_EMAIL=admin@prospeccao.com.br`, `ADMIN_PASSWORD=changeme`, `APP_BASE_URL=http://localhost:8080`. ENCRYPTION_KEY already present from SPEC-01. (FR-010, FR-014)

## Phase 5: Polish (T027-T031)

**Purpose**: Verify ast-grep rule, run make check, validate quickstart scenarios, commit and push.

**Checkpoint**: CI green. All 14 FRs verified. SPEC-03 complete.

- [ ] T027 Create or update ast-grep rule `.ast-grep/rules/go-handler-missing-auth.yml` -- rule that fires on any function in `internal/handler/` that has signature `func(http.ResponseWriter, *http.Request)` but does NOT contain `r.Context().Value("user")` or `r.Context().Value("role")`. Exclude health/ready/live/ping handlers. Verify the rule fires on a test handler without auth context access. (FR-012)
- [ ] T028 Run `make check` and verify it passes: golangci-lint, go test (with coverage >= 85% on non-excluded packages), build-css, go build, ast-grep scan. If any gate fails, fix the underlying issue. (FR-013, quickstart.md Scenario 15)
- [ ] T029 Run all quickstart.md validation scenarios (1-16) and verify each passes. Document any failures and fix. (All FRs, quickstart.md)
- [ ] T030 Verify `make seed` on a fresh test DB creates 1 tenant + 1 admin user with bcrypt password hash. Verify `make dev` starts the server and /login responds. (FR-014, quickstart.md Scenario 16)
- [ ] T031 Commit all changes. Push to main. Verify CI passes via `gh run watch`. Check that the "Test" step runs integration tests against Postgres service container and the coverage gate passes. (FR-013, FR-014)
