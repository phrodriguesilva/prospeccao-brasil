# Data Model: Database Schema & Migrations

**Date**: 2026-07-31
**Spec**: [spec.md](./spec.md)
**Research**: [research.md](./research.md)

## Entity Relationship Overview

```
tenants (root)
  ├── users (tenant-scoped)
  │     └── sessions (tenant-scoped)
  ├── properties (tenant-scoped)
  ├── clients (tenant-scoped)
  │     └── prospections (tenant-scoped)
  │           └── contacts (tenant-scoped, nullable prospect_id)
  │     └── contacts (tenant-scoped, client_id only)
  └── audit_log (tenant-scoped, append-only)
```

**Note on contacts**: The `contacts` table has a nullable `prospect_id`
foreign key. A contact can be either:
- A standalone client interaction (`prospect_id` is NULL -- e.g., a general
  inquiry call before any prospection exists), OR
- Linked to a specific prospection (`prospect_id` references a row in
  `prospections` -- e.g., a negotiation email about a specific property).
Both paths share the same `contacts` table; the `prospect_id` column
distinguishes them. This is why contacts appears under both `prospections`
and `clients` in the diagram above.

## Conventions

- **Primary keys**: `uuid` type, UUID v7 generated in Go (`uuid.NewV7()`),
  fallback `gen_random_uuid()` (v4) in SQL.
- **Tenant-scoped tables**: have `tenant_id uuid NOT NULL REFERENCES tenants(id)`
  + index `idx_<table>_tenant_id`.
- **Timestamps**: `created_at timestamptz NOT NULL DEFAULT now()`,
  `updated_at timestamptz NOT NULL DEFAULT now()`.
- **Soft-delete**: `deleted_at timestamptz` (nullable, NULL = active).
- **Naming**: snake_case for all identifiers.

## Tables

### 1. tenants

Root entity for multi-tenant isolation. Not tenant-scoped (is the tenant).

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK, DEFAULT gen_random_uuid() | |
| name | text | NOT NULL | Firm/prospector name |
| cnpj | text | nullable, UNIQUE | Brazilian tax ID (14 digits); nullable for sole proprietors |
| plan | text | NOT NULL DEFAULT 'free' | free, pro, enterprise |
| active | boolean | NOT NULL DEFAULT true | |
| created_at | timestamptz | NOT NULL DEFAULT now() | |
| updated_at | timestamptz | NOT NULL DEFAULT now() | |
| deleted_at | timestamptz | nullable | soft-delete |

### 2. users

Prospector users with RBAC. Tenant-scoped.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK, DEFAULT gen_random_uuid() | |
| tenant_id | uuid | NOT NULL, FK tenants(id), indexed | |
| email | text | NOT NULL, UNIQUE(tenant_id, email) | |
| full_name | text | NOT NULL | |
| role | text | NOT NULL, CHECK (role IN ('admin','corretor','assistente','financeiro')) | FR-003; admin for MVP |
| password_hash | text | NOT NULL | bcrypt hash |
| totp_secret | text | nullable | app-layer encrypted (SPEC-03) |
| totp_enabled | boolean | NOT NULL DEFAULT false | |
| failed_login_attempts | int | NOT NULL DEFAULT 0 | |
| locked_at | timestamptz | nullable | |
| active | boolean | NOT NULL DEFAULT true | |
| created_at | timestamptz | NOT NULL DEFAULT now() | |
| updated_at | timestamptz | NOT NULL DEFAULT now() | |
| deleted_at | timestamptz | nullable | soft-delete |

**Indexes**: `idx_users_tenant_id`, `idx_users_tenant_email` (unique)

### 3. sessions

Auth sessions with instant revocation. Tenant-scoped.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK, DEFAULT gen_random_uuid() | |
| tenant_id | uuid | NOT NULL, FK tenants(id), indexed | |
| user_id | uuid | NOT NULL, FK users(id) | |
| token_hash | text | NOT NULL, UNIQUE | SHA-256 of session token |
| expires_at | timestamptz | NOT NULL | |
| revoked_at | timestamptz | nullable | NULL = active, timestamp = revoked |
| created_at | timestamptz | NOT NULL DEFAULT now() | |

**Indexes**: `idx_sessions_tenant_id`, `idx_sessions_token_hash` (unique)

### 4. properties

Real-estate properties (imóveis). Tenant-scoped.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK, DEFAULT gen_random_uuid() | |
| tenant_id | uuid | NOT NULL, FK tenants(id), indexed | |
| title | text | NOT NULL | Short title for listing |
| address | text | NOT NULL | Street address (PII if residential) |
| city | text | NOT NULL | |
| state | text | NOT NULL | Brazilian state (2-letter) |
| zip_code | text | nullable | CEP (8 digits) |
| price | numeric(14,2) | NOT NULL | Asking price in BRL |
| status | text | NOT NULL DEFAULT 'available', CHECK (status IN ('available','reserved','sold','inactive')) | FR-005 |
| type | text | NOT NULL, CHECK (type IN ('residential','commercial','land','rural')) | FR-005 |
| bedrooms | int | nullable | |
| bathrooms | int | nullable | |
| area_sqm | numeric(10,2) | nullable | Area in square meters |
| description | text | nullable | Long description |
| photos | jsonb | NOT NULL DEFAULT '[]' | Array of {path, caption} objects (R6) |
| created_at | timestamptz | NOT NULL DEFAULT now() | |
| updated_at | timestamptz | NOT NULL DEFAULT now() | |
| deleted_at | timestamptz | nullable | soft-delete |

**Indexes**: `idx_properties_tenant_id`, `idx_properties_status`, `idx_properties_type`

### 5. clients

Prospector's clients (clientes). Tenant-scoped. PII-heavy (LGPD).

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK, DEFAULT gen_random_uuid() | |
| tenant_id | uuid | NOT NULL, FK tenants(id), indexed | |
| name | text | NOT NULL | |
| email | text | nullable | PII (field-encrypt in SPEC-03) |
| phone | text | nullable | PII (field-encrypt in SPEC-03) |
| cpf_cnpj | text | nullable | PII (field-encrypt in SPEC-03); CPF (11) or CNPJ (14) |
| address | text | nullable | PII (field-encrypt in SPEC-03) |
| budget | numeric(14,2) | nullable | Max budget in BRL |
| preferences | jsonb | NOT NULL DEFAULT '{}' | Flexible search criteria (R7) |
| status | text | NOT NULL DEFAULT 'lead', CHECK (status IN ('active','inactive','lead')) | FR-006 |
| created_at | timestamptz | NOT NULL DEFAULT now() | |
| updated_at | timestamptz | NOT NULL DEFAULT now() | |
| deleted_at | timestamptz | nullable | soft-delete |

**Indexes**: `idx_clients_tenant_id`, `idx_clients_status`

### 6. prospections

Links a client to a property with a prospection status. Tenant-scoped.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK, DEFAULT gen_random_uuid() | |
| tenant_id | uuid | NOT NULL, FK tenants(id), indexed | |
| client_id | uuid | NOT NULL, FK clients(id) | |
| property_id | uuid | NOT NULL, FK properties(id) | |
| status | text | NOT NULL DEFAULT 'new', CHECK (status IN ('new','contacting','visiting','negotiating','closed_won','closed_lost')) | FR-007 |
| notes | text | nullable | PII masked before external rendering |
| contact_date | timestamptz | nullable | First contact date |
| next_action_date | timestamptz | nullable | Next scheduled action |
| created_at | timestamptz | NOT NULL DEFAULT now() | |
| updated_at | timestamptz | NOT NULL DEFAULT now() | |
| deleted_at | timestamptz | nullable | soft-delete |

**Indexes**: `idx_prospections_tenant_id`, `idx_prospections_client_id`, `idx_prospections_property_id`, `idx_prospections_status`

### 7. contacts

Interaction log for client communications. Tenant-scoped.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK, DEFAULT gen_random_uuid() | |
| tenant_id | uuid | NOT NULL, FK tenants(id), indexed | |
| client_id | uuid | NOT NULL, FK clients(id) | |
| prospect_id | uuid | nullable, FK prospections(id) | Nullable; contact may be pre-prospection |
| channel | text | NOT NULL, CHECK (channel IN ('phone','email','whatsapp','in_person')) | FR-008 |
| direction | text | NOT NULL, CHECK (direction IN ('inbound','outbound')) | FR-008 |
| subject | text | nullable | |
| body | text | nullable | PII masked before external rendering |
| contacted_at | timestamptz | NOT NULL DEFAULT now() | When the contact occurred |
| created_at | timestamptz | NOT NULL DEFAULT now() | |

**Indexes**: `idx_contacts_tenant_id`, `idx_contacts_client_id`, `idx_contacts_prospect_id`

### 8. audit_log

Append-only audit trail for LGPD compliance. Tenant-scoped.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK, DEFAULT gen_random_uuid() | |
| tenant_id | uuid | NOT NULL, FK tenants(id), indexed | |
| user_id | uuid | nullable, FK users(id) | Nullable for system actions |
| action | text | NOT NULL | e.g., 'create', 'update', 'delete', 'view' |
| entity_type | text | NOT NULL | e.g., 'client', 'property', 'prospect' |
| entity_id | uuid | nullable | The affected entity's ID |
| metadata | jsonb | nullable | Additional context (field changes, etc.) |
| created_at | timestamptz | NOT NULL DEFAULT now() | |

**Indexes**: `idx_audit_log_tenant_id`, `idx_audit_log_entity`

**Append-only enforcement**: No UPDATE/DELETE in sqlc queries (FR-009).
DB-level REVOKE deferred to future hardening spec.

## Migration File

Single initial migration: `migrations/000001_initial_schema.up.sql` (CREATE
TABLE for all 8 tables) and `migrations/000001_initial_schema.down.sql`
(DROP TABLE in reverse dependency order).

Table creation order (respecting FK dependencies):
1. tenants (no FK)
2. users (FK tenants)
3. sessions (FK tenants, users)
4. properties (FK tenants)
5. clients (FK tenants)
6. prospections (FK tenants, clients, properties)
7. contacts (FK tenants, clients, prospections)
8. audit_log (FK tenants, users)

Drop order (reverse): audit_log, contacts, prospections, clients,
properties, sessions, users, tenants.
