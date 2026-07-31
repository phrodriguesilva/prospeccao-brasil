# Implementation Plan: Internal System (Properties, Clients, Prospecting CRUD + PDF)

**Branch**: `005-internal-system` | **Date**: 2026-07-31 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/005-internal-system/spec.md`

## Summary

The internal system provides property CRUD, client CRUD, prospecting CRUD, contact log, PDF report generation via chromedp, and an admin dashboard. All pages require authentication (SessionValidation + RequireRole(admin)) and enforce tenant_id isolation. Host-based routing separates sistema.* (internal) from prospeccaobrasil.com (public). The database schema and sqlc queries already exist from SPEC-02; this spec adds new sqlc queries for filtering/pagination and dashboard counts.

## Technical Context

**Language/Version**: Go 1.26

**Primary Dependencies**: chi/v5 (router), pgx/v5 (Postgres), sqlc (typed SQL), html/template (server-rendered), HTMX 1.9.12 (interactivity), Alpine.js 3.14.1 (micro-state), chromedp (PDF generation)

**Storage**: PostgreSQL 16 (existing schema from SPEC-02: properties, clients, prospections, contacts tables)

**Testing**: go test -race -p 1 (integration tests against real Postgres, 85% coverage gate)

**Target Platform**: Linux server (Ubuntu 24.04, Hostinger KVM 2 VPS)

**Project Type**: Web service (server-rendered Go monolith, single binary)

**Performance Goals**: Property list < 500ms for 100 properties, PDF generation < 10s, dashboard < 1s

**Constraints**: 85% test coverage, no secrets in repo, tenant_id on every query, no fmt.Println in non-main code, forward-only migrations

**Scale/Scope**: Single-admin MVP (1 tenant, 1 user). ~15 internal pages, ~6 handlers, ~10 new sqlc queries, ~15 templates.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec-Driven | PASS | Following spec lifecycle (specify -> plan -> tasks -> analyze -> implement) |
| II. Security-First (LGPD) | PASS | Client data is PII. tenant_id on every query. Session cookies HttpOnly+SameSite+Secure. 2FA required for admin. |
| III. Single-Binary & Tooling | PASS | Single Go binary. make check runs all gates. CI mirrors Makefile. |
| IV. Test-First & Continuous Quality | PASS | 85% coverage gate. -race -p 1. ast-grep rules enforce tenant_id, auth, bare errors. |
| V. Observability & Structured Logging | PASS | slog for all logging. No fmt.Println in non-main code. |
| VI. Forward-Only Migrations | PASS | No new migrations needed (schema exists from SPEC-02). New sqlc queries only. |
| VII. Simplicity for Single-User | PASS | No multi-user UI, no role management, no tenant switcher. Single admin. |

## Project Structure

### Documentation (this feature)

```text
specs/005-internal-system/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── endpoints.md     # Phase 1 output
└── tasks.md             # Phase 2 output (NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/
├── handler/
│   ├── dashboard.go         # Admin dashboard handler
│   ├── dashboard_test.go    # Dashboard tests
│   ├── property.go          # Property CRUD handler
│   ├── property_test.go     # Property CRUD tests
│   ├── client.go            # Client CRUD handler
│   ├── client_test.go       # Client CRUD tests
│   ├── prospection.go       # Prospection CRUD handler
│   ├── prospection_test.go  # Prospection CRUD tests
│   ├── contact.go           # Contact log handler (extends existing contact.go)
│   ├── contact_test.go      # Contact log tests
│   └── pdf.go               # PDF generation handler (chromedp)
├── db/
│   └── queries/
│       ├── properties.sql   # Extended with pagination/filter queries
│       ├── clients.sql      # Extended with pagination/filter queries
│       ├── prospections.sql # Extended with pagination/filter queries
│       └── dashboard.sql    # New: count queries for dashboard
├── template/
│   ├── admin/               # Internal system templates
│   │   ├── dashboard.html
│   │   ├── properties/
│   │   │   ├── list.html
│   │   │   ├── detail.html
│   │   │   ├── form.html       # Create + edit (shared)
│   │   │   └── _row.html       # HTMX row fragment
│   │   ├── clients/
│   │   │   ├── list.html
│   │   │   ├── detail.html
│   │   │   ├── form.html
│   │   │   └── _row.html
│   │   ├── prospections/
│   │   │   ├── list.html
│   │   │   ├── detail.html
│   │   │   ├── form.html
│   │   │   └── _row.html
│   │   ├── contacts/
│   │   │   ├── _form.html      # Inline form fragment
│   │   │   └── _log.html       # Contact log fragment
│   │   └── _layout.html        # Internal layout (sidebar nav)
│   └── partials/
│       └── internal_nav.html   # Sidebar navigation for internal system
cmd/prospeccao/
└── main.go                  # Updated: internal router with CRUD routes
```

**Structure Decision**: Server-rendered Go monolith. All internal templates live under `internal/template/admin/` with subdirectories per entity. Handlers in `internal/handler/` follow the existing pattern (struct with queries, tmpl, log). No separate frontend -- HTMX handles interactivity, Alpine.js handles micro-state (modals, dropdowns).

## Complexity Tracking

> No constitution violations. No complexity justifications needed.
