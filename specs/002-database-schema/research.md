# Research: Database Schema & Migrations

**Date**: 2026-07-31
**Spec**: [spec.md](./spec.md)

## Research Tasks

### R1: UUID v7 vs bigserial for primary keys

**Decision**: UUID v7 for all primary keys.

**Rationale**:
- UUID v7 provides temporal ordering (timestamp prefix) + global uniqueness.
- Enables future sharding without key migration (no sequential ID collisions).
- Safe for client-side generation (future offline-first features).
- Postgres 16 has native `uuid` type; `pgcrypto` or `uuid-ossp` extension
  provides `uuid_generate_v7()` (or generate in Go via `github.com/google/uuid`).
- Go generation preferred: `uuid.NewV7()` in application layer before insert,
  so we control the version and don't depend on a Postgres extension.

**Alternatives considered**:
- `bigserial` (auto-increment): simpler, smaller (8 bytes vs 16 bytes), faster
  inserts. Rejected because: sequential IDs leak row counts (security concern
  for multi-tenant -- tenant A can infer tenant B's row count); sharding
  requires additional coordination; no client-side generation.
- UUID v4 (random): globally unique but no temporal ordering. Rejected because:
  B-tree index fragmentation (random insertion pattern); UUID v7 gives the same
  uniqueness with ordered insertion (better index locality).

**Implementation**:
- PK column type: `uuid` (Postgres native).
- Default: `gen_random_uuid()` as fallback (Postgres 13+ built-in, produces v4).
  Application layer generates v7 via `uuid.NewV7()` and passes explicitly.
- This gives us v7 ordering when app generates IDs, v4 fallback if SQL inserts
  without explicit ID (e.g., seed data).

### R2: pgcrypto vs app-layer encryption for field-encrypt (cpf_cnpj, phone, address)

**Decision**: App-layer encryption (Go) for SPEC-02, with schema columns typed
as `text` (encrypted blob).

**Rationale**:
- App-layer encryption keeps keys out of the database (defense in depth). If the
  Postgres volume is compromised, encrypted fields remain unreadable without the
  Go application's key.
- pgcrypto (`pgp_sym_encrypt`/`pgp_sym_decrypt`) requires the encryption key to
  be sent to Postgres in the query -- the key transits the connection and lives
  in Postgres memory. This is less secure.
- App-layer encryption allows key rotation without DB downtime (re-encrypt on
  read, write back with new key).
- Go has excellent crypto libraries: `crypto/aes` (GCM mode) in stdlib.

**Alternatives considered**:
- `pgcrypto` column-level encryption: simpler SQL queries (encrypt/decrypt in
  SQL). Rejected because: key exposure to DB, harder key rotation, Postgres
  extension dependency.
- Postgres TDE (Transparent Data Encryption): not available in standard
  Postgres (only in commercial forks). Rejected.
- No field-level encryption (rely on volume encryption only): simplest. Rejected
  because: LGPD requires defense in depth for PII; volume encryption alone
  doesn't protect against SQL injection or DB access.

**Implementation for SPEC-02**:
- Schema defines PII columns as `text` (storing base64-encoded encrypted blobs).
- Encryption/decryption logic is deferred to SPEC-03 (auth) -- SPEC-02 only
  defines the column types and documents the strategy in the data contract.
- `password_hash` and `token_hash` are hashes (bcrypt, SHA-256) -- not
  reversible, no encryption needed.
- `totp_secret` is `text` -- will be app-layer encrypted in SPEC-03.

### R3: soft-delete columns (deleted_at) -- in SPEC-02 or later?

**Decision**: Include `deleted_at` columns in SPEC-02 schema for all
tenant-scoped domain tables (properties, clients, prospections, contacts) and
tenants/users.

**Rationale**:
- The data contract (spec.md) references `deleted_at` in the retention table.
- Adding `deleted_at` now is a forward-only migration (Constitution VI). Adding
  it later would require an ALTER TABLE migration -- not destructive, but
  avoidable by including it now.
- `deleted_at` is nullable (NULL = active, timestamp = soft-deleted).
- sqlc queries filter `WHERE deleted_at IS NULL` for active records.
- Hard delete (physical DELETE) is deferred to a future ops spec for retention
  enforcement.

**Alternatives considered**:
- Defer to SPEC-03: would require ALTER TABLE migration on every table.
  Rejected because: avoidable migration churn; `deleted_at` is a schema
  concern, not an auth concern.
- No soft-delete (hard delete only): rejected because: LGPD requires audit
  trail of deleted data; prospections may need to reference soft-deleted
  clients/properties; accidental deletes are recoverable.

**Tables with `deleted_at`**: tenants, users, properties, clients,
prospections, contacts.
**Tables WITHOUT `deleted_at`**: sessions (hard delete on expiry), audit_log
(never deleted), contacts (soft-delete on client closure -- has deleted_at).

### R4: sqlc configuration (pgx vs database/sql)

**Decision**: Use sqlc pgx recipe (pgx/v5).

**Rationale**:
- pgx/v5 is the recommended Postgres driver for Go (performance, native types,
  prepared statement caching).
- sqlc pgx recipe generates code using `pgx.Tx` and `*pgxpool.Pool` instead of
  `database/sql`.
- SPEC-01 already has `sqlc.yaml` configured with the pgx recipe -- no changes
  needed.

**Implementation**: `sqlc.yaml` (already from SPEC-01):
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/db/queries/"
    schema: "migrations/"
    gen:
      go:
        package: "db"
        out: "internal/db/"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true
        emit_pointers_for_null_types: true
```

### R5: Test strategy for sqlc-generated code

**Decision**: Integration tests against real Postgres (CI service container).

**Rationale**:
- sqlc generates typed Go from SQL -- the generated code is only as correct as
  the SQL. Unit tests on generated code test the generator, not our SQL.
- Integration tests against real Postgres validate: schema correctness, query
  correctness, constraint enforcement, tenant_id isolation.
- CI already has a Postgres 16 service container (from SPEC-01).
- `testcontainers-go` is optional for local dev (requires Docker). CI uses the
  service container directly.

**Implementation**:
- `internal/db/db_test.go` connects to `DATABASE_URL` (test DB).
- Test setup: `migrate up` against test DB, insert seed data (1 tenant, 1 admin
  user).
- Tests: CRUD for each entity, tenant_id isolation (cross-tenant access returns
  empty), audit_log append-only (INSERT only, no UPDATE/DELETE).
- Coverage: generated code is excluded from coverage (it's generated); tests
  cover the SQL queries and schema constraints.

### R6: properties.photos as jsonb vs separate table

**Decision**: `properties.photos` as `jsonb` array of relative file paths.

**Rationale**:
- MVP has simple photo storage (file paths, not binary blobs in DB).
- `jsonb` allows flexible metadata (caption, order, size) without a separate
  table.
- A separate `property_photos` table is overkill for MVP (YAGNI -- Constitution
  principle VII).
- If photo management grows complex (reordering, captions, multiple sizes), a
  future migration can extract to a separate table.

**Alternatives considered**:
- Separate `property_photos` table: more normalized, allows per-photo metadata
  queries. Rejected for MVP because: adds a table + queries + joins for a
  simple use case; jsonb is sufficient for a path array.
- `text[]` (array of paths): simpler than jsonb but no metadata. Rejected
  because: jsonb is more flexible for future caption/order metadata.

**Implementation**:
- Column: `photos jsonb NOT NULL DEFAULT '[]'`.
- Content: `[{"path": "/uploads/prop-1/photo-1.jpg", "caption": "Fachada"}, ...]`.
- Actual file upload/storage is deferred to SPEC-06 (internal system).

### R7: clients.preferences as jsonb vs columns

**Decision**: `clients.preferences` as `jsonb` blob.

**Rationale**:
- Client preferences for real-estate prospecting are flexible: budget range,
  property type, city, neighborhood, number of bedrooms, etc.
- Hardcoding these as columns would require a migration every time a new
  preference field is added -- violates forward-only migration discipline
  (Constitution VI) and adds churn.
- `jsonb` allows flexible schema evolution without migrations.
- Query patterns (filter clients by preferences) use jsonb operators
  (`@>`, `?`, `->>`) which are indexed via GIN if needed (deferred to SPEC-06
  when query patterns are known).

**Alternatives considered**:
- Dedicated columns (budget_min, budget_max, preferred_city, etc.): typed,
  indexable. Rejected because: rigid; every new preference field requires a
  migration; prospector may want custom fields.
- Separate `client_preferences` table (EAV): flexible but complex queries.
  Rejected for MVP because: jsonb is simpler and sufficient.

**Implementation**:
- Column: `preferences jsonb NOT NULL DEFAULT '{}'`.
- Content: `{"budget_min": 500000, "budget_max": 1000000, "type": "residential", "city": "São Paulo"}`.
- Query patterns defined in SPEC-06 (internal system).
