-- 000001_initial_schema.up.sql
-- SPEC-02: Database Schema & Migrations
-- Forward-only migration: CREATE TABLE for all 8 tables.
-- FK dependency order: tenants, users, sessions, properties, clients,
-- prospections, contacts, audit_log.

-- Enable pgcrypto for gen_random_uuid() (Postgres 13+ has it built-in,
-- but the extension guarantees availability).
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 1. tenants (root entity, not tenant-scoped)
CREATE TABLE tenants (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text        NOT NULL,
    cnpj        text        UNIQUE,
    plan        text        NOT NULL DEFAULT 'free',
    active      boolean     NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- 2. users (tenant-scoped, RBAC)
CREATE TABLE users (
    id                     uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id              uuid        NOT NULL REFERENCES tenants(id),
    email                  text        NOT NULL,
    full_name              text        NOT NULL,
    role                   text        NOT NULL CHECK (role IN ('admin','corretor','assistente','financeiro')),
    password_hash          text        NOT NULL,
    totp_secret            text,
    totp_enabled           boolean     NOT NULL DEFAULT false,
    failed_login_attempts  int         NOT NULL DEFAULT 0,
    locked_at              timestamptz,
    active                 boolean     NOT NULL DEFAULT true,
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    deleted_at             timestamptz,
    UNIQUE (tenant_id, email)
);

CREATE INDEX idx_users_tenant_id    ON users(tenant_id);
CREATE INDEX idx_users_tenant_email ON users(tenant_id, email);

-- 3. sessions (tenant-scoped, instant revocation)
CREATE TABLE sessions (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    user_id     uuid        NOT NULL REFERENCES users(id),
    token_hash  text        NOT NULL UNIQUE,
    expires_at  timestamptz NOT NULL,
    revoked_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_tenant_id  ON sessions(tenant_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);

-- 4. properties (tenant-scoped, imóveis)
CREATE TABLE properties (
    id          uuid           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid           NOT NULL REFERENCES tenants(id),
    title        text           NOT NULL,
    address      text           NOT NULL,
    city         text           NOT NULL,
    state        text           NOT NULL,
    zip_code     text,
    price        numeric(14,2)  NOT NULL,
    status       text           NOT NULL DEFAULT 'available' CHECK (status IN ('available','reserved','sold','inactive')),
    type         text           NOT NULL CHECK (type IN ('residential','commercial','land','rural')),
    bedrooms     int,
    bathrooms    int,
    area_sqm     numeric(10,2),
    description  text,
    photos       jsonb          NOT NULL DEFAULT '[]',
    created_at   timestamptz    NOT NULL DEFAULT now(),
    updated_at   timestamptz    NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);

CREATE INDEX idx_properties_tenant_id ON properties(tenant_id);
CREATE INDEX idx_properties_status    ON properties(status);
CREATE INDEX idx_properties_type      ON properties(type);

-- 5. clients (tenant-scoped, PII-heavy, LGPD)
CREATE TABLE clients (
    id           uuid           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid           NOT NULL REFERENCES tenants(id),
    name         text           NOT NULL,
    email        text,
    phone        text,
    cpf_cnpj     text,
    address      text,
    budget       numeric(14,2),
    preferences  jsonb          NOT NULL DEFAULT '{}',
    status       text           NOT NULL DEFAULT 'lead' CHECK (status IN ('active','inactive','lead')),
    created_at   timestamptz    NOT NULL DEFAULT now(),
    updated_at   timestamptz    NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);

CREATE INDEX idx_clients_tenant_id ON clients(tenant_id);
CREATE INDEX idx_clients_status    ON clients(status);

-- 6. prospections (tenant-scoped, links client + property)
CREATE TABLE prospections (
    id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        uuid        NOT NULL REFERENCES tenants(id),
    client_id        uuid        NOT NULL REFERENCES clients(id),
    property_id      uuid        NOT NULL REFERENCES properties(id),
    status           text        NOT NULL DEFAULT 'new' CHECK (status IN ('new','contacting','visiting','negotiating','closed_won','closed_lost')),
    notes            text,
    contact_date     timestamptz,
    next_action_date timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz
);

CREATE INDEX idx_prospections_tenant_id    ON prospections(tenant_id);
CREATE INDEX idx_prospections_client_id    ON prospections(client_id);
CREATE INDEX idx_prospections_property_id ON prospections(property_id);
CREATE INDEX idx_prospections_status       ON prospections(status);

-- 7. contacts (tenant-scoped, immutable interaction log)
-- No deleted_at: contacts are immutable (like audit_log).
-- prospect_id is nullable: a contact can be a standalone client
-- interaction OR linked to a specific prospection.
CREATE TABLE contacts (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid        NOT NULL REFERENCES tenants(id),
    client_id    uuid        NOT NULL REFERENCES clients(id),
    prospect_id  uuid        REFERENCES prospections(id),
    channel      text        NOT NULL CHECK (channel IN ('phone','email','whatsapp','in_person')),
    direction    text        NOT NULL CHECK (direction IN ('inbound','outbound')),
    subject      text,
    body         text,
    contacted_at timestamptz NOT NULL DEFAULT now(),
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_contacts_tenant_id   ON contacts(tenant_id);
CREATE INDEX idx_contacts_client_id   ON contacts(client_id);
CREATE INDEX idx_contacts_prospect_id ON contacts(prospect_id);

-- 8. audit_log (tenant-scoped, append-only)
-- No deleted_at: audit_log is never deleted (LGPD Art. 16, 5-year retention).
-- No UPDATE/DELETE in sqlc queries (FR-009).
CREATE TABLE audit_log (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    user_id     uuid        REFERENCES users(id),
    action      text        NOT NULL,
    entity_type text        NOT NULL,
    entity_id   uuid,
    metadata    jsonb,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_log_tenant_id ON audit_log(tenant_id);
CREATE INDEX idx_audit_log_entity    ON audit_log(entity_type, entity_id);
