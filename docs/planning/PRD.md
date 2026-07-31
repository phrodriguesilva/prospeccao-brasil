# Product Requirements Document: Prospecção Brasil

**Status**: Draft
**Last Updated**: 2026-07-31

## Vision

Prospecção Brasil is a commercial real-estate prospecting platform that
reduces the cognitive load on the prospector. The software manages
properties, clients, and prospecting workflows; the prospector focuses on
relationships and deals.

## Problem

Commercial real-estate prospectors in Brazil manage a high volume of
properties, clients, and ongoing prospecting interactions across scattered
tools (spreadsheets, WhatsApp, paper notes). There is no centralized system
to:

1. Track which properties are available, reserved, or sold.
2. Match clients to properties based on their preferences and budget.
3. Log every client interaction (calls, emails, visits) for follow-up.
4. Generate professional PDF reports for clients and properties.
5. Maintain an LGPD-compliant audit trail of all data access.

The result is lost opportunities, forgotten follow-ups, and legal risk from
unmanaged PII.

## Ideal Customer Profile (ICP)

**Primary ICP**: Independent commercial real-estate prospectors and small
prospecting firms in Brazil.

- **Size**: 1-5 prospectors (MVP targets a single prospector).
- **Location**: Brazil (Portuguese-language UI, BRL currency, Brazilian
  address format, CEP, CPF/CNPJ).
- **Pain**: Managing properties and clients across scattered tools; losing
  track of follow-ups; no professional client-facing materials.
- **Compliance**: Subject to LGPD (Lei Geral de Proteção de Dados) -- must
  handle client PII with audit trails and retention policies.

**MVP user**: Luiz Claudio, a single prospector who needs a simple internal
system to manage his properties, clients, and prospecting pipeline.

## MVP Scope

**Single-tenant, single-admin**. The backend is future-proofed for
multi-tenant (tenant_id on every table, RBAC roles, session + 2FA
encanamento) but the UI is minimal -- one admin manages everything.

MVP deliverables:
1. **Institutional site & design system** (SPEC-04): Home, Quem somos,
   Servicos, Nossos clientes, Fale Conosco, Newsletter. Premium front-end
   for credibility. Design system component classes (buttons, badges,
   cards, forms, nav, footer) built on Tailwind tokens from SPEC-01.
2. **Internal system** (SPEC-05): Property CRUD, client CRUD, prospecting
   pipeline (status tracking), contact log, PDF report generation via
   chromedp.

Non-goals for MVP:
- Multi-tenant SaaS (schema supports it, UI does not).
- Multi-user collaboration (RBAC roles exist in schema, UI is single-admin).
- WhatsApp integration.
- Mobile app.
- Payment processing.

## Future SaaS Multi-Tenant Roadmap

The schema (SPEC-02) ships `tenant_id` on every tenant-scoped table from day
one. The auth middleware (SPEC-03) will support RBAC roles (admin, corretor,
assistente, financeiro) and 2FA. The path from MVP to SaaS:

1. **MVP (now)**: Single-tenant, single-admin. One tenant row, one admin
   user. All features work for this single tenant.
2. **Multi-user (near future)**: Add multiple users per tenant with RBAC.
   The schema and auth middleware already support this -- only UI changes
   needed (login screen, user management).
3. **Multi-tenant SaaS (future)**: Onboarding flow creates a new tenant row,
   seeds an admin user, and isolates all data via tenant_id. The schema is
   ready; the work is in onboarding, billing (SPEC-11 equivalent), and
   tenant management UI.

## Stack Summary

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Language | Go 1.26 | Single binary, fast compilation, strong typing |
| Frontend | HTMX + Alpine.js + Tailwind | Server-rendered, no SPA, self-hosted JS (no CDN SPOF) |
| Database | PostgreSQL 16 | Relational, JSON for flexible fields, LGPD compliance |
| SQL layer | sqlc | Typed Go from SQL, no ORM, SQL is source of truth |
| Migrations | golang-migrate | Forward-only, no destructive changes without ADR |
| PDF | chromedp | Headless Chrome for client/property PDF reports |
| Logging | slog | Structured logging, no fmt.Println in non-main code |
| Testing | go test -race | 85% coverage gate, integration tests against real Postgres |
| Linting | golangci-lint v2 + ast-grep | Deterministic quality gates |
| CI | GitHub Actions | Caches for tools, Postgres service container |

## Specs (Roadmap)

| Spec | Title | Status |
|------|-------|--------|
| SPEC-01 | Repo Tooling & Dev Environment | DONE (CI green) |
| SPEC-02 | Database Schema & Migrations | DONE (CI green) |
| SPEC-03 | Auth + Tenant + RBAC Middleware | DONE (CI green) |
| SPEC-04 | Institutional Site & Design System | Pending |
| SPEC-05 | Internal System (properties, clients, prospecting CRUD + PDF) | Pending |

## References

- [AGENTS.md](../../AGENTS.md) -- AI agent rules, code conventions, quality gates.
- [Constitution](../../.specify/memory/constitution.md) -- 7 non-negotiable principles.
- [SPEC-02 spec](../../specs/002-database-schema/spec.md) -- Database schema with data contract and threat model.
- [Architecture Decisions](../architecture/DECISIONS.md) -- ADRs for key technical choices.
