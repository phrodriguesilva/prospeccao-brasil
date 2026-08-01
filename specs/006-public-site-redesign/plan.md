# Implementation Plan: Public Site Redesign

**Branch**: `006-public-site-redesign` | **Date**: 2026-07-31 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/006-public-site-redesign/spec.md`

## Summary

Redesignar o site institucional publico da Prospecção Brasil para
vender o SERVICO de prospeccao imobiliaria comercial -- nao o software.
O redesign inclui: hero full-bleed com fotografia, paginas profundas de
servicos (5 servicos com `/servicos/{slug}`), historia do fundador com
autoridade de mercado, prova social (depoimentos + metricas), melhoradia
visual do formulario de contato, e correcao CSS dos templates de
autenticacao. Conteudo estatico em Go (maps/slices), sem novas tabelas
exceto uma coluna `company` em `contact_submissions`. Pencil designs
em `designs/prospeccao.pen` antes de HTML/CSS. Brand tokens (Deep Navy,
Sóbrio Gold, Montserrat, Inter) permanecem.

## Technical Context

**Language/Version**: Go 1.26

**Primary Dependencies**: chi/v5 (router, from SPEC-03), html/template
(server-rendered HTML), Tailwind CSS (build-time, from SPEC-01), HTMX
1.9.12 (self-hosted), Alpine.js 3.14.1 (self-hosted), pgx/v5 (DB),
sqlc (typed queries), golang-migrate (forward-only migrations),
Pencil MCP server (design files)

**Storage**: PostgreSQL 16 (existing). Uma migration nova
(`000003_add_company_to_contact_submissions`) adiciona coluna
`company` (VARCHAR(255), NULL) a `contact_submissions`. Nenhuma nova
tabela. Servicos, depoimentos e metricas sao estaticos em Go.

**Testing**: `go test -race -p 1` with integration tests against real
Postgres. 70% coverage gate for app code (85% for `internal/auth`).
Handler tests use httptest + chi router. Testes existentes do SPEC-04
atualizados para novo copy. Novos testes para `/servicos/{slug}` e
auth CSS.

**Target Platform**: Linux server (single Go binary, server-rendered HTML)

**Project Type**: Web service (server-rendered monolith, no SPA)

**Performance Goals**: All institutional pages render in under 200ms
server-side (static content, no DB queries except contact form).

**Constraints**: No CDN dependencies (self-host HTMX/Alpine.js). No
JavaScript-heavy rendering. LGPD-compliant PII handling. Portuguese
(pt-BR) UI. No copy about software/pipeline/carga cognitiva on public
pages. Pencil designs before HTML/CSS.

**Scale/Scope**: 6 paginas publicas redesenhadas (Home, Servicos index,
Servico detalhe, Quem Somos, Nossos Clientes, Fale Conosco), 3 templates
de auth corrigidos (login, 2fa setup, 2fa verify), 1 migration, 1
partial novo (metrics), 1 template novo (servico-detalhe.html), ~8
Pencil frames. Conteudo estatico em Go (services map, testimonials
slice, metrics slice).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Spec-Driven | PASS | SPEC-06 follows Spec Kit lifecycle. Currently at plan stage. |
| II. Security-First (LGPD) | PASS | Contact form already LGPD-compliant (SPEC-04). New `company` field is optional PII, same handling. No secrets in repo. |
| III. Single-Binary & Tooling | PASS | Single Go binary. `make check` covers all gates. CI mirrors Makefile. Tailwind build-time only. |
| IV. Test-First & Continuous Quality | PASS | 70% app / 85% auth coverage gate. Tests updated for new copy. `-race -p 1` maintained. ast-grep rules maintained. |
| V. Observability & Structured Logging | PASS | `slog` logging maintained. No `fmt.Println` in non-main code. |
| VI. Forward-Only Migrations | PASS | Migration `000003` is append-only, forward-only. Down migration exists for dev only. |
| VII. Simplicity for Single-User | PASS | Public site has no user complexity. Static content in Go (no CRUD for services/testimonials). YAGNI. |

No violations. No complexity tracking needed.

## Project Structure

### Documentation (this feature)

```text
specs/006-public-site-redesign/
├── plan.md              # This file
├── research.md          # Phase 0: technical research (13 research items)
├── data-model.md        # Phase 1: schema changes (1 migration, static entities)
├── quickstart.md        # Phase 1: 16 validation scenarios
├── contracts/
│   └── endpoints.md     # Phase 1: route + template contracts
├── checklists/
│   └── requirements.md  # 45-item requirements checklist (specify stage)
└── tasks.md             # Phase 2: task breakdown (NOT created yet -- /speckit-tasks)
```

### Source Code (repository root)

```text
# Migration (new)
migrations/
├── 000003_add_company_to_contact_submissions.up.sql   # NEW
└── 000003_add_company_to_contact_submissions.down.sql # NEW

# SQL queries (modified)
internal/db/queries/
└── contacts.sql                 # MODIFIED: add company param to CreateContactSubmission

# Generated sqlc (auto-regenerated)
internal/db/
└── contacts.sql.go              # REGENERATED via make sqlc

# Handlers (modified)
internal/handler/
├── institutional.go             # MODIFIED: add ServicoDetalhe, services map, testimonials, metrics
├── contact.go                   # MODIFIED: add company field to contactForm
└── institutional_test.go        # MODIFIED: update assertions, add ServicoDetalhe tests

# Templates (rewritten)
internal/template/
├── home.html                    # REWRITTEN: hero full-bleed, metrics, services, testimonials
├── servicos.html                # REWRITTEN: index with 5 services
├── servico-detalhe.html         # NEW: detail page for /servicos/{slug}
├── quem-somos.html              # REWRITTEN: founder story, mission/vision/values, CRECI
├── nossos-clientes.html         # REWRITTEN: testimonials + metrics (no empty state)
├── fale-conosco.html            # REWRITTEN: 2-col layout, contact info in page
├── login.html                   # MODIFIED: add app.css, design system classes
├── totp_setup.html              # MODIFIED: add app.css, design system classes
├── totp_verify.html             # MODIFIED: add app.css, design system classes
└── partials/
    ├── nav.html                 # REWRITTEN: market-facing copy
    ├── footer.html              # REWRITTEN: address, phones, WhatsApp, Instagram
    └── metrics.html             # NEW: reusable metrics strip partial

# Design system (extended)
input.css                        # MODIFIED: add .alert, .alert-error classes

# Static assets (new)
static/img/
├── hero-comercial.jpg           # NEW: hero background (placeholder)
└── .gitkeep                     # ensure directory exists

# Design files (new)
designs/
└── prospeccao.pen               # NEW: Pencil designs (8 frames)

# Router (modified)
cmd/prospeccao/
└── main.go                      # MODIFIED: add GET /servicos/{slug} to public router

# CSS (auto-built)
static/css/
└── app.css                      # REBUILT via make build-css
```

**Structure Decision**: Mantem a estrutura existente do SPEC-04
(handlers em `internal/handler/`, templates em `internal/template/`,
partials em `internal/template/partials/`). Adiciona `servico-detalhe.html`
e `partials/metrics.html`. Nao cria novos packages. Servicos, depoimentos
e metricas sao estaticos em `institutional.go` (maps/slices Go), nao no
banco. Pencil designs em `designs/prospeccao.pen` (ja referenciado no
AGENTS.md).

## Complexity Tracking

> No constitution violations. No complexity tracking needed.

## Implementation Phases (preview for tasks stage)

1. **Phase 1: Setup** -- migration, sqlc, Pencil designs, static img dir
2. **Phase 2: Foundation** -- design system classes (.alert), metrics partial, services map, testimonials, metrics data
3. **Phase 3: US1 Home** -- hero, metrics, services, testimonials, CTA
4. **Phase 4: US2 Servicos** -- index + detail pages, router
5. **Phase 5: US3 Quem Somos** -- founder story, mission/vision/values, CRECI
6. **Phase 6: US4 Nossos Clientes** -- testimonials + metrics (no empty state)
7. **Phase 7: US5 Fale Conosco** -- visual polish, contact info, company field
8. **Phase 8: US6 Auth CSS** -- login, 2fa setup, 2fa verify templates
9. **Phase 9: Polish** -- make check, CI, commit, push

## Key Decisions

1. **Servicos estaticos em Go** (map), nao no banco. YAGNI para 5-10
   servicos fixos de uma consultoria.
2. **Depoimentos estaticos em Go** (slice), nao no banco. 3 depoimentos
   do legacy site. Se crescer, vira outra spec.
3. **Metricas estaticas em Go** (slice). Valores placeholder ate o
   usuario fornecer os reais.
4. **Pencil designs antes de HTML/CSS**. FR-027 exige. AGENTS.md exige
   Pencil como visual source of truth.
5. **Manter estrutura de templates standalone** (cada template e HTML
   completo com nav/footer partials). Nao introduzir `base.html`.
6. **Manter comportamento do formulario** (HTMX, validacao, persistencia,
   fallback no-JS). Apenas visual e campo `company` mudam.
7. **Manter funcionalidade de auth** (login, 2FA, cookies, redirects).
   Apenas CSS muda.
8. **Hero com `background-image` em `<div>` overlay**, nao `<img>`.
   Evita layout shift e permite `bg-cover` sem JS.
9. **Nova classe `.alert`** no design system para mensagens de erro
   nos templates de auth.
10. **Migration `000003`** adiciona coluna `company` a
    `contact_submissions`. Forward-only, append-only.
