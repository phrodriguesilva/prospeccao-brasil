# Quickstart: Database Schema & Migrations

**Date**: 2026-07-31
**Spec**: [spec.md](./spec.md)
**Data Model**: [data-model.md](./data-model.md)
**Query Contracts**: [contracts/queries.md](./contracts/queries.md)

## Prerequisites

- SPEC-01 complete: `make setup` works, `make check` passes, CI green.
- Postgres 16+ running locally (`pg_isready` exits 0).
- `DATABASE_URL` set in `.env.local` (default:
  `postgres://postgres:postgres@localhost:5432/prospeccaobrasil?sslmode=disable`).
- Tools installed: `migrate`, `sqlc` (verified by `make setup`).

## Setup

```bash
# 1. Create the dev database (if not exists)
createdb prospeccaobrasil

# 2. Run migrations
make migrate
# Expected: "1/u initial_schema" applied, exit 0

# 3. Generate sqlc code
make sqlc
# Expected: internal/db/models.go, *.sql.go, db.go created, exit 0

# 4. Verify build
go build ./internal/db/...
# Expected: exit 0
```

## Validation Scenarios

### Scenario 1: All 8 tables exist

```bash
psql -d prospeccaobrasil -c "\dt"
```

Expected output lists 8 tables: `tenants`, `users`, `sessions`, `properties`,
`clients`, `prospections`, `contacts`, `audit_log` (plus `schema_migrations`
from golang-migrate).

### Scenario 2: tenant_id on all tenant-scoped tables

```bash
psql -d prospeccaobrasil -c "SELECT table_name FROM information_schema.columns WHERE column_name='tenant_id' ORDER BY table_name;"
```

Expected: 7 rows -- `audit_log`, `clients`, `contacts`, `properties`,
`prospections`, `sessions`, `users`.

### Scenario 3: FK and index on tenant_id

```bash
psql -d prospeccaobrasil -c "SELECT conname, conrelid::regclass FROM pg_constraint WHERE contype='f' AND conname LIKE '%tenant_id%';"
psql -d prospeccaobrasil -c "SELECT indexname FROM pg_indexes WHERE indexname LIKE 'idx_%_tenant_id';"
```

Expected: 7 FK constraints and 7 indexes on `tenant_id`.

### Scenario 4: users table has all required columns

```bash
psql -d prospeccaobrasil -c "\d users"
```

Expected columns: `id`, `tenant_id`, `email`, `full_name`, `role`,
`password_hash`, `totp_secret`, `totp_enabled`, `failed_login_attempts`,
`locked_at`, `active`, `created_at`, `updated_at`, `deleted_at`. The `role`
column has a CHECK constraint with 6 values: admin, socio, advogado,
estagiario, financeiro, recepcao.

### Scenario 5: sessions table has revocation support

```bash
psql -d prospeccaobrasil -c "\d sessions"
```

Expected columns: `id`, `tenant_id`, `user_id`, `token_hash`, `expires_at`,
`revoked_at` (nullable), `created_at`. `token_hash` has a UNIQUE constraint.

### Scenario 6: properties table has real-estate fields

```bash
psql -d prospeccaobrasil -c "\d properties"
```

Expected columns: `title`, `address`, `city`, `state`, `zip_code`, `price`,
`status` (CHECK: available, reserved, sold, inactive), `type` (CHECK:
residential, commercial, land, rural), `bedrooms`, `bathrooms`, `area_sqm`,
`description`, `photos` (jsonb).

### Scenario 7: prospections links client and property

```bash
psql -d prospeccaobrasil -c "\d prospections"
```

Expected: FKs to `clients(id)` and `properties(id)`, both with `tenant_id`.
`status` CHECK: new, contacting, visiting, negotiating, closed_won, closed_lost.

### Scenario 8: audit_log is append-only

```bash
grep -r "UPDATE audit_log\|DELETE FROM audit_log" internal/db/queries/
# Expected: no matches (exit 1 from grep = no matches = pass)
```

### Scenario 9: migrate up on fresh test DB (CI parity)

```bash
createdb prospeccaobrasil_test
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/prospeccaobrasil_test?sslmode=disable" up
# Expected: "1/u initial_schema" applied, exit 0
dropdb prospeccaobrasil_test
```

CI runs this automatically in the "Migrate (test DB)" step.

### Scenario 10: sqlc generates typed Go

```bash
make sqlc
ls internal/db/*.go
# Expected: models.go, db.go, tenants.sql.go, users.sql.go, sessions.sql.go,
# properties.sql.go, clients.sql.go, prospections.sql.go, contacts.sql.go,
# audit_log.sql.go
go build ./internal/db/...
# Expected: exit 0
```

### Scenario 11: 85% test coverage

```bash
go test -race -p 1 -timeout 20m -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out | grep total
# Expected: total >= 85% (excluding internal/db generated code and
# cmd/prospeccao entry point per coverage gate)
```

### Scenario 12: tenant_id isolation (integration test)

```bash
go test ./internal/db/ -run TestTenantIsolation -v
```

Expected: test inserts data for tenant A and tenant B, queries with tenant A's
ID return only tenant A's data; cross-tenant access returns empty.

### Scenario 13: migrate down is reversible (dev only)

```bash
make migrate-down
psql -d prospeccaobrasil -c "\dt"
# Expected: no domain tables (only schema_migrations)
make migrate
# Expected: tables recreated
```

Note: `migrate down` is for dev only. Production is forward-only (Constitution
principle VI).
