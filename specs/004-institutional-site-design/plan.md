# Implementation Plan: Institutional Site & Design System

**Branch**: `004-institutional-site-design` | **Date**: 2026-07-31 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/004-institutional-site-design/spec.md`

## Summary

Build the public-facing institutional site for Prospecção Brasil with 6
pages (Home, Quem somos, Servicos, Nossos clientes, Fale Conosco, 404)
and a reusable design system (CSS component classes for buttons, badges,
cards, forms, nav, footer) built on the Tailwind tokens from SPEC-01.
Two new database tables (`contact_submissions`, `newsletter_subscribers`)
store PII from public forms. Forms use HTMX for async submission with
server-side fallback. All pages use a shared `base.html` template with
self-hosted HTMX and Alpine.js (no CDN).

## Technical Context

**Language/Version**: Go 1.26

**Primary Dependencies**: chi/v5 (router, from SPEC-03), html/template
(server-rendered HTML), Tailwind CSS (build-time, from SPEC-01), HTMX
1.9.12 (self-hosted), Alpine.js 3.14.1 (self-hosted), pgx/v5 (DB),
sqlc (typed queries), golang-migrate (forward-only migrations)

**Storage**: PostgreSQL 16 (existing, from SPEC-02). Two new tables:
`contact_submissions`, `newsletter_subscribers` (NOT tenant-scoped --
public institutional forms).

**Testing**: `go test -race -p 1` with integration tests against real
Postgres. 85% coverage gate (excluding internal/db, cmd/prospeccao,
scripts). Handler tests use httptest + chi router. Form validation tests
cover valid/invalid/edge cases.

**Target Platform**: Linux server (single Go binary, server-rendered HTML)

**Project Type**: Web service (server-rendered monolith, no SPA)

**Performance Goals**: All institutional pages render in under 200ms
server-side. Forms respond in under 500ms (DB insert + response).

**Constraints**: No CDN dependencies (self-host HTMX/Alpine.js). No
JavaScript-heavy rendering (server-rendered HTML with HTMX for partial
updates). LGPD-compliant PII handling (volume encryption, no plaintext
logging of PII). Portuguese (pt-BR) UI. 85% test coverage minimum.

**Scale/Scope**: 6 institutional pages, 6 design system component classes,
2 new DB tables, 2 new sqlc query files, ~15 templates, ~10 handler
methods. Single-tenant (no tenant context on institutional pages).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec-Driven | PASS | Full template, user stories, Gherkin scenarios |
| II. Security-First (LGPD) | PASS | Data contract in spec.md. PII in contact_submissions (name, email, phone) and newsletter_subscribers (email). Volume encryption at rest. No plaintext PII logging. HTML auto-escaping (XSS prevention). Rate limiting on form endpoints. |
| III. Single-Binary & Tooling | PASS | Single Go binary. Tailwind build-time. make check includes lint+test+build+ast-grep. CI mirrors Makefile. |
| IV. Test-First & 85% Coverage | PASS | Handler tests, form validation tests, integration tests. 85% gate enforced. |
| V. Observability & slog | PASS | All logging via slog. No fmt.Println in non-main code. Form submissions logged. |
| VI. Forward-Only Migrations | PASS | New tables added via forward-only migration. No destructive changes. |
| VII. Simplicity for Single-User | PASS | Institutional site is public (no auth). No multi-user UI. Forms are simple (HTMX async with fallback). |

**Post-Phase 1 Re-check**: PASS. Data model adds 2 tables (no tenant_id
-- intentionally public). No new dependencies beyond what SPEC-01/03
already added. Design system is CSS classes in input.css (no JS framework).

## Project Structure

### Documentation (this feature)

```text
specs/004-institutional-site-design/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── endpoints.md     # HTTP endpoint contracts
└── tasks.md             # Phase 2 output (NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
migrations/
└── 2_contact_newsletter.sql    # New tables: contact_submissions, newsletter_subscribers

internal/
├── db/
│   └── queries/
│       ├── contact_submissions.sql  # sqlc queries for contact form
│       └── newsletter_subscribers.sql  # sqlc queries for newsletter
├── handler/
│   ├── institutional.go      # Public page handlers (Home, Quem somos, etc.)
│   ├── institutional_test.go # Handler tests
│   ├── contact.go            # Contact form handler (POST + validation)
│   ├── contact_test.go       # Contact form tests
│   ├── newsletter.go         # Newsletter signup handler (HTMX fragment)
│   └── newsletter_test.go    # Newsletter tests
└── template/
    ├── base.html             # Shared base template (nav + footer + newsletter)
    ├── home.html             # Home page
    ├── quem-somos.html       # Quem somos page
    ├── servicos.html         # Servicos page
    ├── nossos-clientes.html  # Nossos clientes page
    ├── fale-conosco.html     # Fale Conosco page (with contact form)
    ├── 404.html              # 404 page
    ├── partials/
    │   ├── nav.html          # Navigation bar (extracted from base)
    │   └── footer.html       # Footer with newsletter form (extracted)
    └── fragments/
        ├── contact_success.html  # HTMX fragment: contact form success
        ├── contact_error.html    # HTMX fragment: contact form error
        ├── newsletter_success.html  # HTMX fragment: newsletter success
        └── newsletter_error.html    # HTMX fragment: newsletter error

input.css                     # Add @layer components for design system classes

static/
├── css/
│   └── app.css               # Compiled Tailwind (build-time)
└── js/
    ├── htmx.min.js           # Self-hosted HTMX 1.9.12
    └── alpine.min.js         # Self-hosted Alpine.js 3.14.1

cmd/prospeccao/
└── main.go                   # Add institutional routes to chi router
```

**Structure Decision**: Server-rendered monolith. Templates in
`internal/template/` (same as SPEC-03 auth templates). Handlers in
`internal/handler/` (same as SPEC-03). New DB queries in
`internal/db/queries/`. New migration in `migrations/`. Design system
component classes in `input.css` (compiled by Tailwind to
`static/css/app.css`). Self-hosted JS in `static/js/`.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
