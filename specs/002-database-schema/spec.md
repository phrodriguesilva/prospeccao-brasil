# Feature Specification (Slim): Database Schema & Migrations

**Feature Branch**: `002-database-schema`

**Created**: 2026-07-31

**Status**: Draft

**Template**: slim (for infrastructure/tooling specs). See AGENTS.md
"Spec template selection" for when to use this vs the full template.

**Input**: User description: "SPEC-02: Database Schema & Migrations"

## Overview

This spec defines the PostgreSQL schema that underpins all Prospecção Brasil
features. It creates forward-only migrations for the encanamento tables
(tenants, users, sessions, audit_log) and the domain tables (properties,
clients, prospections, contacts). sqlc generates typed Go from SQL queries --
no ORM. The schema enforces multi-tenant isolation via `tenant_id` on every
tenant-scoped table with foreign keys and indexes.

This is an infrastructure spec: it delivers engineering value (data model,
type-safe queries, migration discipline) rather than direct product value to
the prospector (Luiz Claudio). Every downstream feature spec (auth, design
system, internal system, institutional site) depends on this schema existing.

The MVP is single-tenant, single-admin (Constitution principle VII), but the
schema ships `tenant_id` on every tenant-scoped table from day one so the
encanamento is in place for future multi-tenant use without a painful
migration later.

## Context

**Canonical sources:**
- [Constitution](../../.specify/memory/constitution.md) principles II
  (Security-First LGPD), III (Single-Binary), VI (Forward-Only Migrations),
  VII (Simplicity for Single-User)
- [AGENTS.md](../../AGENTS.md) -- sqlc, golang-migrate, multi-tenant
  isolation, no ORM, SQL is source of truth
- Reference: `pragmaos/specs/002-database-schema/spec.md` (adapted for
  commercial real-estate prospecting domain -- no cases, proceedings,
  movements, deadlines, hearings, communications_log; instead properties,
  clients, prospections, contacts)

**Dependencies**: SPEC-01 (Repo Tooling & Dev Environment) -- complete.
**Gate to run**: SPEC-01 is complete (Go module initialized, Makefile
orchestrates quality gates, pre-commit hooks installed, CI green, dev
environment reproducible via `make setup`, `migrations/` and
`internal/db/queries/` directories exist with `.gitkeep`).

## Goals

1. Create a PostgreSQL schema via forward-only migrations that supports all
   Prospecção Brasil domain entities (tenants, users, sessions, properties,
   clients, prospections, contacts, audit_log) plus the encanamento for
   future RBAC and 2FA.
2. Enforce multi-tenant isolation at the schema level: `tenant_id` column
   with FK to `tenants.id` and an index on every tenant-scoped table.
3. Generate typed Go database access code via sqlc from SQL queries -- no ORM,
   SQL is the source of truth.
4. Ensure `migrate up` runs successfully against a fresh test DB in CI.

## Non-Goals

The following are explicitly deferred to later specs:

- Auth flows (login, 2FA, RBAC middleware) -> SPEC-03
- Design system and UI components -> SPEC-04
- Institutional site (Home, Quem somos, Serviços, etc.) -> SPEC-05
- Internal system (properties/clients/prospections CRUD + PDF) -> SPEC-06
- DB-level REVOKE on audit_log (append-only enforcement at DB level) ->
  future hardening spec
- Retention cron jobs (session cleanup, archival) -> future ops spec

## Requirements

These requirements are the verifiable acceptance criteria. Each is copied
from the project roadmap and is non-negotiable.

- **FR-001**: `tenants`, `users`, `sessions`, `properties`, `clients`,
  `prospections`, `contacts`, `audit_log` tables exist via forward-only
  migration in `migrations/`. Verify: `migrate up` against a fresh test DB
  succeeds; `psql -c "\dt"` lists all 8 tables.
- **FR-002**: Every tenant-scoped table has `tenant_id` column with FK to
  `tenants.id` and an index. Tenant-scoped tables are: `users`, `sessions`,
  `properties`, `clients`, `prospections`, `contacts`, `audit_log`. Verify:
  `psql -c "SELECT table_name FROM information_schema.columns WHERE
  column_name='tenant_id'"` returns all 7 tenant-scoped tables; FK and index
  existence checked via `information_schema.table_constraints` and
  `pg_indexes`.
- **FR-003**: `users` table includes `role` (admin for MVP; CHECK
  constraint allows admin, corretor, assistente, financeiro for
  future-proofing), `password_hash`, `totp_secret`,
  `totp_enabled`, `failed_login_attempts`, `locked_at`. Verify:
  `psql -c "\d users"` shows all columns; `role` is a CHECK constraint with
  the 4 values.
- **FR-004**: `sessions` table includes `token_hash`, `user_id`,
  `tenant_id`, `expires_at`, `revoked_at` (for instant revocation). Verify:
  `psql -c "\d sessions"` shows all columns; `revoked_at` is nullable
  (active sessions have NULL).
- **FR-005**: `properties` table (imóveis) includes `title`, `address`,
  `city`, `state`, `zip_code`, `price`, `status` (available, reserved,
  sold, inactive), `type` (residential, commercial, land, rural),
  `bedrooms`, `bathrooms`, `area_sqm`, `description`, `photos` (jsonb array
  of URLs/paths). Verify: `psql -c "\d properties"` shows all columns;
  `status` and `type` are CHECK constraints.
- **FR-006**: `clients` table (clientes) includes `name`, `email`, `phone`,
  `cpf_cnpj`, `budget`, `preferences` (jsonb), `status` (active, inactive,
  lead). Verify: `psql -c "\d clients"` shows all columns; `cpf_cnpj` and
  `email` are PII (LGPD).
- **FR-007**: `prospections` table (prospecções) links `client_id` to
  `property_id` with `status` (new, contacting, visiting, negotiating,
  closed_won, closed_lost), `notes`, `contact_date`, `next_action_date`.
  Verify: `psql -c "\d prospections"` shows FKs to `clients(id)` and
  `properties(id)`, both tenant-scoped.
- **FR-008**: `contacts` table (contatos) logs interactions with a client:
  `client_id`, `prospect_id` (nullable), `channel` (phone, email, whatsapp,
  in_person), `direction` (inbound, outbound), `subject`, `body`,
  `contacted_at`. Verify: `psql -c "\d contacts"` shows FK to `clients(id)`,
  nullable FK to `prospections(id)`.
- **FR-009**: `audit_log` is append-only (no UPDATE/DELETE in sqlc queries;
  DB-level REVOKE deferred to future hardening). Verify: `grep -r
  "UPDATE audit_log\|DELETE FROM audit_log" internal/db/queries/` returns
  no matches; sqlc generated code has no Update/Delete methods for
  audit_log.
- **FR-010**: `migrate up` runs successfully against a fresh test DB (CI
  enforces). Verify: CI step "Migrate (test DB)" passes; `make migrate`
  locally exits 0.
- **FR-011**: sqlc generates typed Go in `internal/db/` from queries in
  `internal/db/queries/`. Verify: `make sqlc` exits 0; `internal/db/`
  contains generated `.go` files; `go build ./internal/db/...` succeeds.
- **FR-012**: 85% test coverage on new Go code (excluding sqlc-generated
  `internal/db` which is "DO NOT EDIT"). Verify: `go test -coverprofile=
  coverage.out ./... && go tool cover -func=coverage.out | grep total`
  shows >= 85% on non-excluded packages.

## Constraints

1. Single Go binary + Postgres. No Docker/K8s required for the MVP dev
   environment (Constitution principle III).
2. No secrets in the repo; `.env.example` for samples; gitleaks pre-commit
   hook (Constitution principle II).
3. No emojis anywhere -- code, UI, comments, docs, commits (project rule).
4. No conventional-commit prefixes (rejected by CI).
5. Structured logging via slog (Constitution principle V); no `fmt.Println`
   in non-main code (enforced by ast-grep rule `go-slog-fmt.yml`).
6. Forward-only migrations via golang-migrate (Constitution principle VI); no
   destructive migrations without an ADR and dual-write period.
7. sqlc is the source of truth for typed Go from SQL; no ORM (AGENTS.md).
8. Multi-tenant isolation: `tenant_id` filter on every tenant-scoped table;
   cross-tenant access is a critical bug (Constitution principle II, LGPD).
9. Client and property data is LGPD-sensitive: schema must support audit log
   of all access to client/property data (Constitution principle II).
10. PII fields (cpf_cnpj, phone, email, address) must be designed for
    app-layer encryption (columns typed to store encrypted blobs; encryption
    enforced in SPEC-03).

## Definition of Done

Database Schema & Migrations is done when ALL of the following are verifiable:

| # | Acceptance Criterion | Verification Command | FR | Status |
|---|----------------------|----------------------|----|--------|
| 1 | All 8 tables exist via forward-only migration | `migrate up` against fresh test DB; `psql -c "\dt"` | FR-001 | [ ] |
| 2 | Every tenant-scoped table has `tenant_id` with FK + index | `psql -c "SELECT ... FROM information_schema..."` | FR-002 | [ ] |
| 3 | `users` table has role, password_hash, totp_secret, totp_enabled, failed_login_attempts, locked_at | `psql -c "\d users"` | FR-003 | [ ] |
| 4 | `sessions` table has token_hash, user_id, tenant_id, expires_at, revoked_at | `psql -c "\d sessions"` | FR-004 | [ ] |
| 5 | `properties` table has all real-estate fields with status/type CHECK | `psql -c "\d properties"` | FR-005 | [ ] |
| 6 | `clients` table has name, email, phone, cpf_cnpj, budget, preferences, status | `psql -c "\d clients"` | FR-006 | [ ] |
| 7 | `prospections` table links client_id to property_id with FKs + status | `psql -c "\d prospections"` | FR-007 | [ ] |
| 8 | `contacts` table logs interactions with client_id + nullable prospect_id | `psql -c "\d contacts"` | FR-008 | [ ] |
| 9 | `audit_log` is append-only (no UPDATE/DELETE in sqlc queries) | `grep -r "UPDATE audit_log\|DELETE FROM audit_log" internal/db/queries/` | FR-009 | [ ] |
| 10 | `migrate up` succeeds on fresh test DB (CI enforces) | CI "Migrate (test DB)" step passes; `make migrate` exits 0 | FR-010 | [ ] |
| 11 | sqlc generates typed Go in `internal/db/` from queries | `make sqlc` exits 0; `go build ./internal/db/...` succeeds | FR-011 | [ ] |
| 12 | 85% test coverage on new Go code | `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out \| grep total` | FR-012 | [ ] |

**Spec is ready for `/speckit-plan` when all rows are checked.**

## Data Contract

Generated by `speckit-tekimax-security-data-contract` hook (mandatory for
SPEC-02: has data entities, PII, tenant_id). Adapted for Go + Postgres +
sqlc (no Zod/TypeScript -- schema is SQL migrations, typed Go via sqlc).

### Sources

| Name | Origin | Trust | Schema Location | PII? |
|------|--------|-------|-----------------|------|
| tenants | DB table (migration) | vetted | `migrations/` + sqlc | name, cnpj |
| users | DB table (migration) | vetted | `migrations/` + sqlc | email, full_name, password_hash, totp_secret |
| sessions | DB table (migration) | vetted | `migrations/` + sqlc | token_hash |
| properties | DB table (migration) | vetted | `migrations/` + sqlc | address, city, zip_code, price |
| clients | DB table (migration) | vetted | `migrations/` + sqlc | name, email, phone, cpf_cnpj, address |
| prospections | DB table (migration) | vetted | `migrations/` + sqlc | client_id, property_id, notes |
| contacts | DB table (migration) | vetted | `migrations/` + sqlc | client_id, subject, body |
| audit_log | DB table (migration) | vetted | `migrations/` + sqlc | user_id, action, entity_id (access trail) |

### Schema Definition

Schemas are defined in forward-only SQL migrations in `migrations/` and
materialized as typed Go via sqlc in `internal/db/`. No Zod/TypeScript
schemas -- this is a Go project. The SQL migration is the source of truth;
sqlc generates `internal/db/models.go` and query-specific `*.sql.go` files.

Key schema constraints (enforced at DB level):
- `tenant_id` FK + index on every tenant-scoped table (FR-002)
- `users.role` CHECK constraint with 4 values: admin, corretor,
  assistente, financeiro (FR-003)
- `sessions.revoked_at` nullable (active = NULL, revoked = timestamp) (FR-004)
- `properties.status` CHECK: available, reserved, sold, inactive (FR-005)
- `properties.type` CHECK: residential, commercial, land, rural (FR-005)
- `clients.status` CHECK: active, inactive, lead (FR-006)
- `prospections.status` CHECK: new, contacting, visiting, negotiating,
  closed_won, closed_lost (FR-007)
- `contacts.channel` CHECK: phone, email, whatsapp, in_person (FR-008)
- `contacts.direction` CHECK: inbound, outbound (FR-008)
- `audit_log` append-only: no UPDATE/DELETE in sqlc queries (FR-009)
- `created_at` / `updated_at` timestamps on all tables
- UUID primary keys (gen_random_uuid() default; UUID v7 in Go preferred)

### PII Handling

LGPD-sensitive fields (Constitution principle II). PII must be encrypted at
rest (app-layer) and masked before any external display.

| Field | Table | Strategy | Implementation |
|-------|-------|----------|----------------|
| email | users, clients | hash (lookup) | pgcrypto `digest()` for token_hash; email stored for app use (encrypted at rest via Postgres volume encryption) |
| password_hash | users | hash | bcrypt (never plaintext, never reversible) |
| totp_secret | users | field-encrypt | encrypted at rest; decrypted only in auth flow (SPEC-03) |
| token_hash | sessions | hash | SHA-256 of session token; token never stored |
| cpf_cnpj | clients | field-encrypt | encrypted at rest; masked in UI (***.***.***-**) |
| phone | clients | field-encrypt | encrypted at rest; masked in UI |
| address | properties, clients | field-encrypt | encrypted at rest |
| notes | prospections | omit (for external) | PII masked before any external rendering |
| body | contacts | omit (for external) | PII masked before any external rendering |

Note: "field-encrypt" means Postgres column-level encryption or
application-layer encryption before insert. The concrete provider is
configured in SPEC-03 (auth). For SPEC-02, the schema defines the columns;
encryption strategy is documented here and enforced in SPEC-03.

### Bias Audit

- **Segments that must be represented**: N/A for SPEC-02 (schema only, no
  filtering or scoring logic). Bias audit applies to prospection features.
- **Known bias risks**: None at schema level. Schema is neutral -- `tenant_id`
  isolation prevents cross-tenant data leakage but does not introduce bias.
- **Mitigation**: [DEFERRED: SPEC-06] Bias audit will run when prospection
  features consume this schema.

### Drift Monitoring

- **Baseline**: N/A for SPEC-02 (no data pipeline or scoring at schema level).
- **Threshold**: N/A.
- **Detection**: [DEFERRED: future ops spec] Drift monitoring applies to
  features that consume this data. Schema migrations are version-controlled
  and reviewed in PRs (Constitution principle VI).

### Retention

| Data | TTL | Deletion Path | LGPD Basis |
|------|-----|---------------|------------|
| clients | indefinite (while tenant active) | soft-delete (`deleted_at` column); hard delete on tenant termination | Art. 15 (data retention for legal obligation) |
| properties | indefinite (while tenant active) | soft-delete (`deleted_at` column); hard delete on tenant termination | Art. 15 |
| prospections | indefinite (while client/property active) | soft-delete on closure; hard delete on tenant termination | Art. 15 |
| contacts | indefinite (while client active) | hard delete on tenant termination (immutable log, no soft-delete) | Art. 15 |
| audit_log | 5 years (LGPD Art. 16) | archival job (future ops spec); no in-app DELETE | Art. 16 (legal retention period) |
| sessions | 30 days after `expires_at` | cron job (future ops spec) deletes expired sessions | Art. 15 |

Note: Retention policies are documented here for schema design (e.g.,
`deleted_at` columns). Enforcement is deferred to SPEC-03 (auth) and a
future ops spec.

## Security / Threat Model

**Generated**: 2026-07-31
**Scope**: SPEC-02 -- Database Schema & Migrations
**Attack surface**: PostgreSQL schema (8 tables), migration files,
sqlc-generated queries, DATABASE_URL connection. No HTTP endpoints in this
spec.

### Threats

| ID | Category | Threat | Severity | Likelihood | Mitigation | Status |
|----|----------|--------|----------|------------|------------|--------|
| T1 | Spoofing | Attacker connects to Postgres using leaked DATABASE_URL and impersonates the app to read/modify all tenant data | High | Low | DATABASE_URL in .env.local (gitignored, gitleaks-enforced); Postgres listens on localhost only in dev; production uses TLS 1.2+ + SCRAM-SHA-256 auth (Constitution II) | Mitigated (SPEC-02: gitleaks + .gitignore; prod hardening deferred) |
| T2 | Spoofing | Session token theft: attacker reads `sessions.token_hash` from DB and attempts to reconstruct the original token | High | Low | `token_hash` stores SHA-256 hash, not the raw token. Raw token is never persisted. Hash is one-way. Token entropy >= 256 bits (Constitution II) | Mitigated (schema design: hash-only storage) |
| T3 | Tampering | Attacker modifies migration files in `migrations/` to drop tables or weaken constraints (e.g., remove tenant_id FK) | High | Low | Forward-only migrations (Constitution VI); migration files reviewed in PR (CI runs `migrate up` against fresh test DB); pre-commit hooks (gitleaks, ast-grep); branch protection on main | Mitigated (PR review + CI + forward-only) |
| T4 | Tampering | Cross-tenant data modification: sqlc query missing `tenant_id` WHERE clause allows tenant A to update tenant B's records | Critical | Medium | Every tenant-scoped query in `internal/db/queries/` includes `AND tenant_id = $X`; ast-grep rule `go-missing-tenant-filter` catches missing tenant_id in Go code; integration tests verify cross-tenant access returns empty | Mitigated (query contracts + ast-grep + tests) |
| T5 | Repudiation | User performs action on client/property data but audit_log entry is missing or incomplete, preventing LGPD accountability | High | Medium | `audit_log` table with `user_id`, `action`, `entity_type`, `entity_id`, `created_at`; append-only (FR-009); every access to client/property data must log (Constitution II) | Mitigated (schema design: audit_log table + append-only queries) |
| T6 | Repudiation | Attacker deletes audit_log entries to cover their tracks | Critical | Low | `audit_log` has no DELETE in sqlc queries (FR-009); DB-level REVOKE on audit_log deferred to future hardening; soft-delete columns excluded from audit_log | Partially mitigated (query-level enforced; DB-level REVOKE deferred) |
| T7 | Information Disclosure | PII fields (cpf_cnpj, phone, address, email) stored in plaintext in DB, exposed via SQL injection or DB dump | Critical | Medium | App-layer encryption for PII fields (AES-GCM in Go, keys out of DB); Postgres volume encryption (Constitution II); PII masked before external rendering | Mitigated (app-layer encryption in SPEC-03; schema defines columns now) |
| T8 | Information Disclosure | `users.totp_secret` stored in plaintext allows attacker to bypass 2FA | High | Low | `totp_secret` column typed as `text` (will store encrypted blob); app-layer encryption in SPEC-03 (auth); `totp_enabled` flag defaults to false | Mitigated (schema design: column exists; encryption in SPEC-03) |
| T9 | Information Disclosure | Error messages from Postgres reveal schema details (table names, column names, constraint names) to attacker | Medium | Medium | sqlc generates typed Go errors (no raw SQL in error messages); application layer wraps errors with `fmt.Errorf("...: %w", err)` (AGENTS.md); no `fmt.Println` in non-main code (ast-grep go-slog-fmt.yml) | Mitigated (sqlc typed errors + error wrapping) |
| T10 | Denial of Service | Attacker floods `sessions` table with INSERT, causing table bloat and slow queries | Medium | Low | `sessions.token_hash` has UNIQUE constraint (prevents duplicate floods); `DeleteExpiredSessions` query cleans expired sessions; rate limiting deferred to SPEC-03 (auth) | Partially mitigated (schema: unique constraint + cleanup query; rate limiting in SPEC-03) |

## Assumptions

1. The MVP is single-tenant, single-admin (Luiz Claudio). The schema ships
   `tenant_id` on all tenant-scoped tables for future-proofing, but only one
   tenant row will exist in the MVP (seeded in SPEC-03).
2. UUID v4 (`gen_random_uuid()`) is used as the SQL default for primary keys.
   UUID v7 (time-ordered) may be generated in Go for better index locality;
   this is a plan.md decision, not a spec constraint.
3. `properties.photos` is a `jsonb` array of relative file paths or URLs.
   Actual file upload/storage is deferred to SPEC-06.
4. `clients.preferences` is a `jsonb` blob for flexible search criteria
   (budget range, property type, city, etc.). Schema is flexible; query
   patterns are defined in SPEC-06.
5. App-layer encryption for PII fields uses AES-GCM with keys from
   `ENCRYPTION_KEY` env var (already in `.env.example` from SPEC-01). The
   schema defines columns as `text` (storing encrypted blobs); encryption
   is applied in SPEC-03 (auth) before insert.
6. The `audit_log` table is tenant-scoped (has `tenant_id`) so each tenant
   sees only its own audit trail. Global audit (cross-tenant) is a future
   hardening concern.
