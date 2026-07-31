# Data Contract: SPEC-05 Internal System

**Created**: 2026-07-31
**Status**: Active

## Entities

### Property (imóvel)

| Field | Type | Constraints | PII |
|-------|------|-------------|-----|
| id | UUID | PK, gen_random_uuid() | No |
| tenant_id | UUID | FK tenants(id), NOT NULL | No |
| title | text | NOT NULL, min 3 chars | No |
| address | text | NOT NULL, min 5 chars | No |
| city | text | NOT NULL | No |
| state | text | NOT NULL | No |
| zip_code | text | nullable | No |
| price | numeric(14,2) | NOT NULL, > 0 | No |
| status | text | NOT NULL, CHECK IN ('available','reserved','sold','inactive') | No |
| type | text | NOT NULL, CHECK IN ('residential','commercial','land','rural') | No |
| bedrooms | int | nullable | No |
| bathrooms | int | nullable | No |
| area_sqm | numeric(10,2) | nullable | No |
| description | text | nullable | No |
| photos | jsonb | NOT NULL DEFAULT '[]' (array of URL strings) | No |
| created_at | timestamptz | NOT NULL DEFAULT now() | No |
| updated_at | timestamptz | NOT NULL DEFAULT now() | No |
| deleted_at | timestamptz | nullable (soft delete) | No |

### Client (cliente)

| Field | Type | Constraints | PII |
|-------|------|-------------|-----|
| id | UUID | PK, gen_random_uuid() | No |
| tenant_id | UUID | FK tenants(id), NOT NULL | No |
| name | text | NOT NULL, min 2 chars | Yes (name) |
| email | text | nullable, valid email format | Yes (email) |
| phone | text | nullable | Yes (phone) |
| cpf_cnpj | text | nullable | Yes (CPF/CNPJ) |
| address | text | nullable | Yes (address) |
| budget | numeric(14,2) | nullable, >= 0 | No |
| preferences | jsonb | NOT NULL DEFAULT '{}' | No |
| status | text | NOT NULL, CHECK IN ('active','inactive','lead') | No |
| created_at | timestamptz | NOT NULL DEFAULT now() | No |
| updated_at | timestamptz | NOT NULL DEFAULT now() | No |
| deleted_at | timestamptz | nullable (soft delete) | No |

### Prospection (prospecção)

| Field | Type | Constraints | PII |
|-------|------|-------------|-----|
| id | UUID | PK, gen_random_uuid() | No |
| tenant_id | UUID | FK tenants(id), NOT NULL | No |
| client_id | UUID | FK clients(id), NOT NULL | No (references PII) |
| property_id | UUID | FK properties(id), NOT NULL | No |
| status | text | NOT NULL, CHECK IN ('new','contacting','visiting','negotiating','closed_won','closed_lost') | No |
| notes | text | nullable | No |
| contact_date | timestamptz | nullable | No |
| next_action_date | timestamptz | nullable | No |
| created_at | timestamptz | NOT NULL DEFAULT now() | No |
| updated_at | timestamptz | NOT NULL DEFAULT now() | No |
| deleted_at | timestamptz | nullable (soft delete) | No |

### Contact (interaction log)

| Field | Type | Constraints | PII |
|-------|------|-------------|-----|
| id | UUID | PK, gen_random_uuid() | No |
| tenant_id | UUID | FK tenants(id), NOT NULL | No |
| client_id | UUID | FK clients(id), NOT NULL | No (references PII) |
| prospect_id | UUID | FK prospections(id), nullable | No |
| channel | text | NOT NULL, CHECK IN ('phone','email','whatsapp','in_person') | No |
| direction | text | NOT NULL, CHECK IN ('inbound','outbound') | No |
| subject | text | nullable | No |
| body | text | nullable | No |
| contacted_at | timestamptz | NOT NULL DEFAULT now() | No |
| created_at | timestamptz | NOT NULL DEFAULT now() | No |

**Note**: Contacts have NO `deleted_at` column -- they are immutable (LGPD audit trail).

## Data Flow

```
Tenant (1) ---> (N) Property
Tenant (1) ---> (N) Client
Tenant (1) ---> (N) Prospection ---> (1) Client
                                ---> (1) Property
Tenant (1) ---> (N) Contact ---> (1) Client
                             ---> (0..1) Prospection
```

## Tenant Isolation

Every query MUST include `WHERE tenant_id = $1`. The `tenant_id` is extracted from the authenticated user's session context (`r.Context().Value(auth.CtxUser).TenantID`). Cross-tenant access returns 404 (not 403, to avoid information leakage).

## LGPD Considerations

- **Client data** (name, email, phone, cpf_cnpj, address) is PII under LGPD.
- **Contact log** is immutable and serves as an audit trail of client interactions.
- **Soft delete** (deleted_at) is used for clients/properties/prospections to preserve referential integrity and audit history. Hard delete is not available in the UI.
- **Data retention**: 5 years per LGPD Art. 16 (enforced at the database level, not in this spec).
