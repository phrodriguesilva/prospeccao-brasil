# Data Model: Internal System (SPEC-05)

**Date**: 2026-07-31

## Existing Entities (from SPEC-02)

All entities already exist in the database schema. No new migrations are needed. This document describes the entities as they relate to SPEC-05.

### Property

```
properties
├── id (UUID, PK)
├── tenant_id (UUID, FK tenants)
├── title (text, NOT NULL)
├── address (text, NOT NULL)
├── city (text, NOT NULL)
├── state (text, NOT NULL)
├── zip_code (text, nullable)
├── price (numeric(14,2), NOT NULL)
├── status (text, NOT NULL, CHECK: available|reserved|sold|inactive)
├── type (text, NOT NULL, CHECK: residential|commercial|land|rural)
├── bedrooms (int, nullable)
├── bathrooms (int, nullable)
├── area_sqm (numeric(10,2), nullable)
├── description (text, nullable)
├── photos (jsonb, NOT NULL DEFAULT '[]') -- array of URL strings
├── created_at (timestamptz)
├── updated_at (timestamptz)
└── deleted_at (timestamptz, nullable) -- soft delete
```

**Indexes**: idx_properties_tenant_id, idx_properties_status, idx_properties_type

**Validation rules**:
- title: min 3 chars
- address: min 5 chars
- city: min 2 chars
- state: min 2 chars
- price: > 0
- status: enum (available, reserved, sold, inactive)
- type: enum (residential, commercial, land, rural)

### Client

```
clients
├── id (UUID, PK)
├── tenant_id (UUID, FK tenants)
├── name (text, NOT NULL)
├── email (text, nullable)
├── phone (text, nullable)
├── cpf_cnpj (text, nullable) -- PII
├── address (text, nullable) -- PII
├── budget (numeric(14,2), nullable)
├── preferences (jsonb, NOT NULL DEFAULT '{}')
├── status (text, NOT NULL, CHECK: active|inactive|lead)
├── created_at (timestamptz)
├── updated_at (timestamptz)
└── deleted_at (timestamptz, nullable) -- soft delete
```

**Indexes**: idx_clients_tenant_id, idx_clients_status

**Validation rules**:
- name: min 2 chars
- email: valid email format (net/mail.ParseAddress)
- budget: >= 0
- status: enum (active, inactive, lead)

**PII**: name, email, phone, cpf_cnpj, address -- all under LGPD

### Prospection

```
prospections
├── id (UUID, PK)
├── tenant_id (UUID, FK tenants)
├── client_id (UUID, FK clients, NOT NULL)
├── property_id (UUID, FK properties, NOT NULL)
├── status (text, NOT NULL, CHECK: new|contacting|visiting|negotiating|closed_won|closed_lost)
├── notes (text, nullable)
├── contact_date (timestamptz, nullable)
├── next_action_date (timestamptz, nullable)
├── created_at (timestamptz)
├── updated_at (timestamptz)
└── deleted_at (timestamptz, nullable) -- soft delete
```

**Indexes**: idx_prospections_tenant_id, idx_prospections_client_id, idx_prospections_property_id, idx_prospections_status

**Status state machine**:
```
new -> contacting -> visiting -> negotiating -> closed_won
                                            -> closed_lost
```
Any status can transition to any other status (no strict enforcement in MVP -- the admin can correct mistakes).

### Contact

```
contacts
├── id (UUID, PK)
├── tenant_id (UUID, FK tenants)
├── client_id (UUID, FK clients, NOT NULL)
├── prospect_id (UUID, FK prospections, nullable)
├── channel (text, NOT NULL, CHECK: phone|email|whatsapp|in_person)
├── direction (text, NOT NULL, CHECK: inbound|outbound)
├── subject (text, nullable)
├── body (text, nullable)
├── contacted_at (timestamptz, NOT NULL DEFAULT now())
└── created_at (timestamptz, NOT NULL DEFAULT now())
```

**Indexes**: idx_contacts_tenant_id, idx_contacts_client_id, idx_contacts_prospect_id

**Immutability**: No `deleted_at`, no `updated_at`. Contacts are append-only (LGPD audit trail).

## New sqlc Queries (to be added)

### Dashboard

```sql
-- name: CountPropertiesByTenant :one
SELECT COUNT(*) FROM properties WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: CountClientsByTenant :one
SELECT COUNT(*) FROM clients WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: CountProspectsByTenant :one
SELECT COUNT(*) FROM prospections WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: CountProspectsByStatus :many
SELECT status, COUNT(*) as count FROM prospections
WHERE tenant_id = $1 AND deleted_at IS NULL
GROUP BY status;

-- name: ListRecentProspectsWithDetails :many
SELECT p.*, c.name as client_name, pr.title as property_title
FROM prospections p
JOIN clients c ON p.client_id = c.id
JOIN properties pr ON p.property_id = pr.id
WHERE p.tenant_id = $1 AND p.deleted_at IS NULL
ORDER BY p.created_at DESC
LIMIT $2;
```

### Filtered/Paginated Lists

```sql
-- name: ListPropertiesFiltered :many
SELECT * FROM properties
WHERE tenant_id = $1 AND deleted_at IS NULL
  AND ($2::text IS NULL OR status = $2)
  AND ($3::text IS NULL OR type = $3)
  AND ($4::text IS NULL OR title ILIKE '%' || $4 || '%' OR city ILIKE '%' || $4 || '%')
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountPropertiesFiltered :one
SELECT COUNT(*) FROM properties
WHERE tenant_id = $1 AND deleted_at IS NULL
  AND ($2::text IS NULL OR status = $2)
  AND ($3::text IS NULL OR type = $3)
  AND ($4::text IS NULL OR title ILIKE '%' || $4 || '%' OR city ILIKE '%' || $4 || '%');

-- Similar pattern for clients and prospections
```

## Entity Relationships

```
Tenant 1---N Property
Tenant 1---N Client
Tenant 1---N Prospection N---1 Client
                         N---1 Property
Tenant 1---N Contact N---1 Client
                      N---0..1 Prospection
```

## Tenant Isolation

Every query includes `WHERE tenant_id = $1`. The `tenant_id` is extracted from:
```go
user := r.Context().Value(auth.CtxUser).(*db.User)
tenantID := user.TenantID
```

Cross-tenant access returns 404 (not 403) to avoid information leakage.
