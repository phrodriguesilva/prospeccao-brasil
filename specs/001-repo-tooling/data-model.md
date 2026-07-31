# Data Model: SPEC-01 -- Repo Tooling & Dev Environment

**Date**: 2026-07-31
**Spec**: [spec.md](./spec.md)

## Overview

SPEC-01 is a tooling/infrastructure spec. It introduces **no domain
entities** -- no tenants, users, sessions, properties, clients, prospections,
or contacts. Those begin in SPEC-02 (Database Schema & Migrations).

The only "data" in SPEC-01 is:

1. **`/healthz` response body**: `{"status":"ok"}` (a fixed JSON string,
   not a persisted entity).
2. **`migrations/.gitkeep`**: an empty file to track the directory in git.
   No schema content.
3. **`sqlc.yaml`**: a config file referencing directories that SPEC-02
   will populate. No queries, no schema.
4. **`static/css/app.css`**: the Tailwind build output (generated, not
   hand-authored). Contains the compiled token set.
5. **`.env.example`**: placeholder environment variables. No real values.

## Entities

None.

## State Transitions

None.

## Validation Rules

None (no data to validate).

## Deferred to SPEC-02

The following entities will be defined in SPEC-02 (Database Schema &
Migrations) and are listed here only to confirm they are out of scope:

- `tenants` -- the multi-tenant encanamento (1 row in MVP: Prospecção Brasil).
- `users` -- the single admin user (Luiz Claudio) with 2FA.
- `sessions` -- session cookies with TOTP verification state.
- `properties` -- commercial real-estate listings (address, type, area,
  value, status, photos, geolocation).
- `clients` -- companies/contacts the prospections target.
- `prospections` -- the link between a property and a client with status
  (prospectando / apresentado / fechado / perdido) and dates.
- `contacts` -- Fale Conosco form submissions and newsletter sign-ups
  (from the public site).

SPEC-01 creates the `migrations/` directory and the `sqlc.yaml` config so
that SPEC-02 can add the first migration and the first queries without
touching tooling.
