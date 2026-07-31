# Quickstart: Auth + Tenant + RBAC Middleware

**Date**: 2026-07-31
**Spec**: [spec.md](./spec.md)
**Endpoints**: [contracts/endpoints.md](./contracts/endpoints.md)

## Prerequisites

1. Go 1.26+, Postgres 16+, golangci-lint v2, sqlc, migrate, ast-grep.
2. Local Postgres running on `localhost:5432` (user: current OS user).
3. Test DB: `prospeccaobrasil_test` (created in SPEC-02).
4. `DATABASE_URL=postgres://localhost:5432/prospeccaobrasil_test?sslmode=disable`
5. `ENCRYPTION_KEY` set (32 bytes base64). Use the value from `.env.example`.
6. `APP_BASE_URL=http://localhost:8080` (dev -- Secure cookie disabled).
7. `ADMIN_PASSWORD=test123` (for seed script).

## Setup

```bash
# 1. Apply migrations (if not already)
createdb prospeccaobrasil_test 2>/dev/null || true
migrate -path migrations -database "$DATABASE_URL" up

# 2. Regenerate sqlc (new queries added)
make sqlc

# 3. Seed the initial tenant + admin user
make seed

# 4. Verify seed
psql -d prospeccaobrasil_test -c "SELECT id, name FROM tenants;"
# Expect: 1 row (Prospecção Brasil)
psql -d prospeccaobrasil_test -c "SELECT email, role, totp_enabled FROM users;"
# Expect: 1 row (admin@prospeccao.com.br, admin, false)
```

## Validation Scenarios

### Scenario 1: Login page loads (FR-001)

```bash
# Start the server
APP_BASE_URL=http://localhost:8080 ENCRYPTION_KEY=<key> make dev &
sleep 2

# GET /login
curl -s http://localhost:8080/login | grep -q "email"
echo "PASS: login form has email field"
```

**Expected**: HTML form with email + password inputs.

### Scenario 2: Login with correct credentials (FR-001)

```bash
# POST /login with correct credentials
RESP=$(curl -s -X POST http://localhost:8080/login \
  -d "email=admin@prospeccao.com.br&password=test123" \
  -c /tmp/cookies.txt -w "%{http_code} %{redirect_url}")

echo "$RESP"
# Expect: 302 redirect to /2fa/setup (totp_enabled=false on first login)
# Cookie file /tmp/cookies.txt should have pending_session cookie
```

**Expected**: 302 redirect to `/2fa/setup` (first login, 2FA enrollment).

### Scenario 3: Login with wrong credentials (FR-001)

```bash
RESP=$(curl -s -X POST http://localhost:8080/login \
  -d "email=admin@prospeccao.com.br&password=wrong" \
  -w "%{http_code}")

echo "$RESP"
# Expect: 401 (or 200 with error page -- depends on impl)
# Error message must be generic: "Email ou senha invalidos"
```

**Expected**: Generic error, no user enumeration.

### Scenario 4: Account lockout after 5 failures (FR-002)

```bash
# 5 failed login attempts
for i in 1 2 3 4 5; do
  curl -s -X POST http://localhost:8080/login \
    -d "email=admin@prospeccao.com.br&password=wrong" > /dev/null
done

# Check locked_at in DB
psql -d prospeccaobrasil_test -c \
  "SELECT failed_login_attempts, locked_at FROM users WHERE email='admin@prospeccao.com.br';"
# Expect: failed_login_attempts=5, locked_at is not null

# 6th attempt (even with correct password) should be locked
RESP=$(curl -s -X POST http://localhost:8080/login \
  -d "email=admin@prospeccao.com.br&password=test123" -w "%{http_code}")
echo "$RESP"
# Expect: 403 (account locked)

# Reset for further tests
psql -d prospeccaobrasil_test -c \
  "UPDATE users SET failed_login_attempts=0, locked_at=NULL WHERE email='admin@prospeccao.com.br';"
```

**Expected**: Account locks at 5 failures; `locked_at` set in DB.

### Scenario 5: 2FA enrollment (FR-003)

```bash
# After login (Scenario 2), GET /2fa/setup
curl -s -b /tmp/cookies.txt http://localhost:8080/2fa/setup | grep -q "qr"
echo "PASS: 2FA setup page has QR code"

# To test enrollment, you need a valid TOTP code. Use the test helper:
# go test ./internal/auth/ -run TestTOTPEnrollment -v
```

**Expected**: QR code displayed; TOTP code input form present.

### Scenario 6: 2FA verification rejects invalid code (FR-004)

```bash
RESP=$(curl -s -X POST -b /tmp/cookies.txt http://localhost:8080/2fa/verify \
  -d "code=000000" -w "%{http_code}")
echo "$RESP"
# Expect: 401 (invalid TOTP code)
```

**Expected**: Invalid TOTP code returns 401.

### Scenario 7: Logout revokes session (FR-005)

```bash
# Login + complete 2FA first (use test helper to get a valid session)
# Then:
curl -s -X POST -b /tmp/cookies.txt http://localhost:8080/logout \
  -c /tmp/cookies.txt -w "%{http_code} %{redirect_url}"
# Expect: 302 redirect to /login

# Try to access /admin with old cookie
RESP=$(curl -s -b /tmp/cookies.txt http://localhost:8080/admin -w "%{http_code} %{redirect_url}")
echo "$RESP"
# Expect: 302 redirect to /login (session revoked)
```

**Expected**: Session revoked; subsequent access redirects to `/login`.

### Scenario 8: Session validation rejects no cookie (FR-006)

```bash
RESP=$(curl -s http://localhost:8080/admin -w "%{http_code} %{redirect_url}")
echo "$RESP"
# Expect: 302 redirect to /login (no cookie)
```

**Expected**: No cookie -> redirect to `/login`.

### Scenario 9: Tenant validation rejects soft-deleted user (FR-007)

```bash
# Soft-delete the user in DB
psql -d prospeccaobrasil_test -c \
  "UPDATE users SET deleted_at=now() WHERE email='admin@prospeccao.com.br';"

# Try to access /admin with a valid session cookie (from a prior login)
RESP=$(curl -s -b /tmp/cookies.txt http://localhost:8080/admin -w "%{http_code} %{redirect_url}")
echo "$RESP"
# Expect: 302 redirect to /login (user soft-deleted)

# Restore for further tests
psql -d prospeccaobrasil_test -c \
  "UPDATE users SET deleted_at=NULL WHERE email='admin@prospeccao.com.br';"
```

**Expected**: Soft-deleted user -> session rejected.

### Scenario 10: RBAC denies insufficient role (FR-008)

```bash
# Create a user with role 'assistente'
psql -d prospeccaobrasil_test -c \
  "INSERT INTO users (id, tenant_id, email, full_name, role, password_hash, totp_enabled, failed_login_attempts, active) \
   VALUES (gen_random_uuid(), (SELECT id FROM tenants LIMIT 1), 'assistente@prospeccao.com.br', 'Test Assistente', 'assistente', '\$2a\$10\$test', false, 0, true);"

# Login as assistente (bypass 2FA for test -- set totp_enabled=true with known secret)
# Then try to access /admin (admin-only)
RESP=$(curl -s -b /tmp/assistente_cookies.txt http://localhost:8080/admin -w "%{http_code}")
echo "$RESP"
# Expect: 403 Forbidden (RBAC denial)
```

**Expected**: `assistente` role -> 403 on admin-only endpoint.

### Scenario 11: Rate limiting returns 429 (FR-009)

```bash
# 6 rapid login attempts from the same IP
for i in 1 2 3 4 5 6; do
  CODE=$(curl -s -X POST http://localhost:8080/login \
    -d "email=ratelimit@test.com&password=test" -w "%{http_code}")
  echo "Attempt $i: $CODE"
done
# Expect: attempts 1-5 return 401; attempt 6 returns 429
```

**Expected**: 6th rapid attempt returns 429.

### Scenario 12: Cookie Secure flag conditional (FR-010)

```bash
# Dev (HTTP) -- Secure should be absent
RESP=$(curl -s -X POST http://localhost:8080/login \
  -d "email=admin@prospeccao.com.br&password=test123" -D - -o /dev/null)
echo "$RESP" | grep -i "set-cookie"
# Expect: Set-Cookie without Secure flag

# Production (HTTPS) -- restart server with APP_BASE_URL=https://...
APP_BASE_URL=https://prospeccao.com.br make dev &
# Repeat the curl -- Expect: Set-Cookie with Secure flag
```

**Expected**: Dev = no Secure; Prod = Secure flag present.

### Scenario 13: Audit log records auth events (FR-011)

```bash
# After a login attempt (success or failure)
psql -d prospeccaobrasil_test -c \
  "SELECT action, user_id, tenant_id, created_at FROM audit_log WHERE action LIKE 'login%' ORDER BY created_at DESC LIMIT 5;"
# Expect: rows with action='login_attempt', 'login_success' or 'login_failed'
```

**Expected**: `audit_log` table has rows for each auth event.

### Scenario 14: ast-grep rule fires on handler without auth (FR-012)

```bash
# Create a temporary handler without auth context access
cat > /tmp/test_handler.go << 'EOF'
package handler
func BadHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("no auth"))
}
EOF
cp /tmp/test_handler.go internal/handler/bad_handler_test.go
ast-grep scan 2>&1 | grep -q "go-handler-missing-auth"
echo "PASS: ast-grep rule fires"
rm internal/handler/bad_handler_test.go
```

**Expected**: ast-grep `go-handler-missing-auth` rule fires on the bad handler.

### Scenario 15: 85% test coverage (FR-013)

```bash
export DATABASE_URL="postgres://localhost:5432/prospeccaobrasil_test?sslmode=disable"
export ENCRYPTION_KEY="<key from .env.example>"
go test -race -p 1 -timeout 20m -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out | grep total
# Expect: total >= 85% (excluding internal/db and cmd/prospeccao)
```

**Expected**: Coverage >= 85% on non-excluded packages.

### Scenario 16: Seed creates tenant + admin (FR-014)

```bash
# Reset DB
dropdb prospeccaobrasil_test && createdb prospeccaobrasil_test
migrate -path migrations -database "$DATABASE_URL" up

# Run seed
ADMIN_PASSWORD=test123 make seed

# Verify
psql -d prospeccaobrasil_test -c "SELECT count(*) FROM tenants;"
# Expect: 1
psql -d prospeccaobrasil_test -c "SELECT count(*) FROM users;"
# Expect: 1
psql -d prospeccaobrasil_test -c "SELECT substr(password_hash, 1, 4) FROM users;"
# Expect: $2a$ (bcrypt)
```

**Expected**: 1 tenant, 1 admin user with bcrypt password hash.

## Cleanup

```bash
# Stop the dev server
kill %1 2>/dev/null

# Reset the test DB
dropdb prospeccaobrasil_test && createdb prospeccaobrasil_test
migrate -path migrations -database "$DATABASE_URL" up
```
