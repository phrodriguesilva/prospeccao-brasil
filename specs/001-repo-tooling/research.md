# Research: SPEC-01 -- Repo Tooling & Dev Environment

**Date**: 2026-07-31
**Spec**: [spec.md](./spec.md)
**Plan**: [plan.md](./plan.md)

## Method

The reference project (`/Users/relterborges/Documents/Dev/pragmaos`) has a
proven, working tooling stack. This research is a **transplant analysis**: for
each tooling component, what to copy verbatim, what to adapt, what to write
new, and why. The goal is to reuse the engineering investment of the reference
project while tailoring the domain-specific artifacts (AGENTS.md,
constitution, Tailwind tokens) to Prospecção Brasil.

Each item is documented with: (1) what exists in the reference, (2) the
decision (copy / adapt / new), (3) the rationale, and (4) the specific
substitutions or changes.

---

## Item 1: Makefile

**Reference**: `pragmaos/Makefile` -- targets: setup, dev, check, lint, test,
build-css, build, migrate, migrate-down, sqlc, fmt, ast-grep, run. Loads
`.env.local`. Test target uses `-p 1` (sequential) and excludes generated
`internal/db/` and `cmd/` from coverage gate.

**Decision**: ADAPT (copy with substitutions).

**Rationale**: The Makefile structure is identical for any Go + HTMX +
Postgres monolith. Only the binary name and module path change.

**Substitutions**:
- `pragmaos` -> `prospeccao` (binary name, `cmd/pragmaos` -> `cmd/prospeccao`,
  `bin/pragmaos` -> `bin/prospeccao`).
- `make dev` runs `go run ./cmd/prospeccao`.
- `make build` runs `go build -o bin/prospeccao ./cmd/prospeccao`.
- Test coverage exclusions: `internal/db/` (sqlc-generated), `cmd/prospeccao/`
  (entry point), `cmd/test-login` (drop -- does not exist here).
- DB name in bootstrap: `pragmaos` -> `prospeccaobrasil`.

---

## Item 2: scripts/bootstrap.sh

**Reference**: `pragmaos/scripts/bootstrap.sh` -- 7-step setup: check
prerequisites, create .env.local, create dev DB, run migrations, run sqlc,
install pre-commit, run make check. Checks Go 1.26+, Postgres 16+.

**Decision**: ADAPT (copy with substitutions).

**Rationale**: Identical workflow for any Go + Postgres project. Only the
project name and DB name change.

**Substitutions**:
- `PragmaOS` -> `Prospecção Brasil` (echo banners).
- `DB_NAME="pragmaos"` -> `DB_NAME="prospeccaobrasil"`.
- Go version check stays at 1.26+ (same requirement).
- Postgres version check stays at 16+.

---

## Item 3: .pre-commit-config.yaml

**Reference**: `pragmaos/.pre-commit-config.yaml` -- gitleaks v8.21.2,
pre-commit-hooks v5.0.0 (trailing-whitespace, end-of-file-fixer, check-yaml,
check-json, check-merge-conflict, check-added-large-files, detect-private-key),
local hooks (gofmt, go-imports, golangci-lint, ast-grep).

**Decision**: COPY verbatim.

**Rationale**: Pre-commit hooks are project-agnostic. The gitleaks version,
pre-commit-hooks version, and local Go hooks are identical. The
`check-yaml` exclude pattern for `.specify/extensions/` templates is also
identical (the extensions.yml structure is the same).

**No substitutions needed.**

---

## Item 4: .ast-grep/rules/ (7 rules)

**Reference**: `pragmaos/.ast-grep/rules/` -- 7 rules: go-bare-error,
go-handler-missing-auth, go-hardcoded-secret, go-missing-context,
go-missing-tenant-filter, go-slog-fmt, tmpl-bare-button.

**Decision**: COPY verbatim.

**Rationale**: The ast-grep rules encode constitution constraints that are
identical between the two projects (no bare errors, no hardcoded secrets,
handler auth, tenant filter, slog not fmt, no bare buttons). The
`go-missing-tenant-filter` rule will fire once SPEC-03 adds `tenant_id` to
queries; shipping it now is correct because the constitution principle II
(tenant isolation) is part of this project too, even though the MVP has 1
tenant.

**No substitutions needed.** The rules are Go/HTML structural patterns, not
domain-specific.

---

## Item 5: sgconfig.yml

**Reference**: `pragmaos/sgconfig.yml` -- points to `.ast-grep/rules`, file
types go + html (tmpl, templ).

**Decision**: COPY verbatim.

**No substitutions needed.**

---

## Item 6: sqlc.yaml

**Reference**: `pragmaos/sqlc.yaml` -- postgresql, pgx/v5, queries in
`internal/db/queries/`, schema in `migrations/`, emit json_tags, interface,
pointers for null types.

**Decision**: COPY verbatim.

**Rationale**: The sqlc config is identical for any Go + Postgres + pgx
project using the same code generation conventions.

**No substitutions needed.**

---

## Item 7: tailwind.config.js

**Reference**: `pragmaos/tailwind.config.js` -- sober theme: gray scale, navy
scale, border tokens, status colors, sharp corners (radius 0px), no shadows,
Inter font, dense table font sizes.

**Decision**: NEW (write from scratch using the user-provided design system
spec).

**Rationale**: The Prospecção Brasil design system is the OPPOSITE aesthetic
of PragmaOS. PragmaOS is sober/dense/sharp-corners/no-shadows (law-firm
utility). Prospecção Brasil is premium/minimalist/soft-radius/ambient-shadows
(Real Estate Intelligence for B2B executives). Copying the pragmaos Tailwind
config would produce the wrong visual identity.

**New token set** (from the user-provided design system):
- Colors: `primary` (Deep Navy `#031636` / `#1a2b4c`), `secondary` (Sóbrio
  Gold `#765a1a` / `#b89650`), `surface` (`#fcf9f8`), `surface-container`
  variants, `slate-gray` (`#334155`), `whatsapp-green` (`#25D366`), `error`
  (`#ba1a1a`), outline/outline-variant, inverse colors, fixed/variant tokens.
- Typography: `display-lg` (Montserrat 48px/700), `headline-lg` (32px/600),
  `headline-md` (24px/600), `body-lg` (Inter 18px/400), `body-md` (16px/400),
  `label-sm` (12px/600, uppercase, tracking 0.05em).
- Radius: `sm` 0.125rem, `DEFAULT` 0.25rem, `md` 0.375rem, `lg` 0.5rem,
  `xl` 0.75rem, `full` 9999px. (NOT 0px like pragmaos.)
- Shadows: ambient (blur 20px, opacity 4%, color `#1A2B4C`) for floated cards.
  (NOT `none` like pragmaos.)
- Spacing: `section-gap` 80px, `margin-mobile` 20px, `gutter-mobile` 16px,
  `stack-sm/md/lg` 8/16/32px.
- Fonts: Montserrat (sans-display), Inter (sans-body).
- Content paths: `./internal/ui/templates/**/*.html`,
  `./internal/handler/templates/**/*.html`, `./**/*.go`.

---

## Item 8: input.css

**Reference**: `pragmaos/input.css` -- Tailwind directives + badge component
classes (badge-green, badge-red, etc. using status colors).

**Decision**: ADAPT (copy the Tailwind directives; defer component classes to
SPEC-04).

**Rationale**: SPEC-01 ships only the Tailwind tooling (build-css target
works). The component classes (badges, buttons, cards) are SPEC-04 (design
system). Shipping them now would be premature -- the token set is new and
component classes should be designed against Pencil frames in SPEC-04.

**Content**: Just the three Tailwind directives (`@tailwind base/components/
utilities`). SPEC-04 will add `@layer components` blocks.

---

## Item 9: package.json

**Reference**: `pragmaos/package.json` -- name "pragmaos", private, scripts
build-css/watch-css, devDependency tailwindcss 3.4.17.

**Decision**: ADAPT (copy with name substitution).

**Substitutions**:
- `"name": "pragmaos"` -> `"name": "prospeccaobrasil"`.
- Description updated to reference Prospecção Brasil.
- tailwindcss version stays pinned at `3.4.17` (NOT `@latest` -- per the
  constitution's "no floating ranges" constraint and the global rule about
  newly-published versions).

---

## Item 10: .env.example

**Reference**: `pragmaos/.env.example` -- DATABASE_URL, SESSION_SECRET,
TOTP_ISSUER, ENCRYPTION_KEY, rate limiters, WHATSAPP_API_TOKEN,
WHATSAPP_PHONE_ID, CNJ_API_BASE, LLM_PROVIDER/API_KEY/MODEL, SMTP,
APP_BASE_URL.

**Decision**: ADAPT (copy, remove inapplicable vars, change issuer).

**Substitutions**:
- `TOTP_ISSUER=PragmaOS` -> `TOTP_ISSUER=Prospeccao Brasil`.
- Remove: `WHATSAPP_API_TOKEN`, `WHATSAPP_PHONE_ID`, `CNJ_API_BASE`,
  `LLM_PROVIDER`, `LLM_API_KEY`, `LLM_MODEL`, `LLM_FALLBACK_MODEL` (no
  AI/WhatsApp/CNJ in this project's MVP).
- Keep: `DATABASE_URL`, `SESSION_SECRET`, `ENCRYPTION_KEY`, rate limiters,
  `SMTP_*`, `APP_BASE_URL`.
- `DATABASE_URL` default: `postgres://postgres:postgres@localhost:5432/prospeccaobrasil?sslmode=disable`.

---

## Item 11: .github/workflows/ci.yml

**Reference**: `pragmaos/.github/workflows/ci.yml` -- golangci-lint, go test
with `-p 1` and `ENCRYPTION_KEY` env var, go build, ast-grep scan,
govulncheck.

**Decision**: ADAPT (copy with path substitutions).

**Substitutions**:
- `./cmd/pragmaos` -> `./cmd/prospeccao` in build step.
- `bin/pragmaos` -> `bin/prospeccao`.
- Go version pinning stays at 1.26+.
- The `-p 1` flag and `ENCRYPTION_KEY` env var are kept (CI parity rule).
- govulncheck stays.

---

## Item 12: .devin/config.json + mcp_config.json

**Reference**: `pragmaos/.devin/config.json` and `mcp_config.json` -- both
configure the ast-grep MCP server via `uvx`, with `AST_GREP_CONFIG` env var
pointing to the repo's `sgconfig.yml`.

**Decision**: ADAPT (copy with path substitution).

**Substitutions**:
- `AST_GREP_CONFIG` path: `/Users/relterborges/Documents/Dev/pragmaos/sgconfig.yml`
  -> `/Users/relterborges/Documents/Dev/prospeccao-brasil/sgconfig.yml`.
- The `PATH` env var stays the same (homebrew, go bin paths are user-global).

---

## Item 13: .specify/extensions.yml

**Reference**: `pragmaos/.specify/extensions.yml` -- installed: review,
spectest, tekimax-security. Hooks: after_implement (review, spectest gaps,
security audit, ast-grep scan), after_specify (data-contract, optional),
after_plan (threat-model, optional), before_implement (gate-check,
mandatory), before_analyze (red-team, optional).

**Decision**: COPY verbatim.

**Rationale**: The hook structure is project-agnostic. The security hooks
(data-contract, threat-model, red-team, gate-check, audit) apply equally to
Prospecção Brasil because it handles LGPD-sensitive client/property data
(SPEC-02, SPEC-03, SPEC-06). The review and spectest hooks are universal.

**No substitutions needed.** The `optional` flags and priorities are
identical.

---

## Item 14: .specify/templates/spec-template-slim.md

**Reference**: `pragmaos/.specify/templates/spec-template-slim.md` -- slim
template for infrastructure/tooling specs.

**Decision**: COPY verbatim.

**Rationale**: The slim template is project-agnostic. It has placeholders
like `[FEATURE NAME]` and references to AGENTS.md/constitution that work for
any project.

**No substitutions needed.**

---

## Item 15: AGENTS.md

**Reference**: `pragmaos/AGENTS.md` -- comprehensive agent rules: project
overview (law-firm SaaS), spec-driven development, spec template selection,
hooks & implementation reports, code conventions (Go, templates, API, auth,
commits), quality gates, CI parity, frontend conventions (self-host JS,
focus trap), structural code analysis, visual design (Pencil), documentation,
key commands, MCP servers, CLI tools.

**Decision**: NEW (write from scratch, modeled on the reference structure but
adapted to the Prospecção Brasil domain).

**Rationale**: AGENTS.md is domain-specific. The PragmaOS version has
extensive sections on AI/CNJ/WhatsApp/LLM/legal-judgment/multi-tenant-RBAC
that do not apply here. The Prospecção Brasil version keeps the structural
sections (spec-driven, code conventions, quality gates, frontend, key
commands, MCP servers) but rewrites the project overview, removes AI/legal
sections, and adds the single-user MVP scope note.

**Sections to keep (adapted)**: Project Overview, Spec-Driven Development,
Spec template selection, Code Conventions (Go, Templates, API, Auth, Commits),
Quality Gates, CI parity, Frontend Conventions (self-host JS, focus trap),
Structural Code Analysis, Visual Design (Pencil), Documentation, Key
Commands, MCP Servers & CLI Tools.

**Sections to remove**: Spec Kit Hooks & Implementation Reports (the
security-hook skip policy is law-firm-spec-specific; keep a simplified
version), AI/CNJ/WhatsApp/LLM references, multi-tenant RBAC role list
(socio/advogado/etc), integration test isolation tech debt note (no
integration tests yet).

**Sections to add**: Single-User MVP scope note (Luiz Claudio is the only
user; encanamento ready for future commercialization but no multi-user UI
prematurely), PDF generation note (chromedp, SPEC-06).

---

## Item 16: .specify/memory/constitution.md

**Reference**: `pragmaos/.specify/memory/constitution.md` -- 7 principles:
I. Spec-Driven, II. Security-First (LGPD), III. Single-Binary & Tooling, IV.
Test-First, V. Observability, VI. Agent Safety & Guardrails (AI human-in-
the-loop), VII. Forward-Only Migrations.

**Decision**: NEW (write from scratch, 7 principles adapted to Prospecção
Brasil).

**Rationale**: The constitution is domain-specific. Principle VI (Agent
Safety & Guardrails for AI legal suggestions) does not apply -- there is no
AI in this project's MVP. It is replaced by "Simplicity for Single-User"
(the UI must not be complex for Luiz Claudio; future-proof encanamento
without premature multi-user features).

**New 7 principles**:
1. **Spec-Driven Development** -- same as reference.
2. **Security-First Design** -- LGPD for client/property data, 2FA, tenant_id
   encanamento, audit logging (deferred to SPEC-03+ but principle stands).
3. **Single-Binary & Tooling Consistency** -- same as reference.
4. **Test-First & Continuous Quality** -- 85% coverage, govulncheck.
5. **Observability & Structured Logging** -- slog, /healthz, /readyz.
6. **Forward-Only Migrations** -- same as reference.
7. **Simplicity for Single-User** -- UI must not be complex for the owner
   (Luiz Claudio). Encanamento (auth, tenant_id, RBAC middleware) is
   future-proof for commercialization, but no premature multi-user UI. YAGNI
   for the MVP.

**Quality Gates table**: same structure as reference (Format, Lint, Vet,
Test+coverage, Build, Structural scan, Secret scan, Dependency vuln, Spec Kit
review, Security audit, Security gate).

---

## Item 17: Self-hosted JS (HTMX, Alpine, modal-trap)

**Reference**: `pragmaos/static/js/htmx.min.js` (v1.9.12, 48KB),
`alpine.min.js` (v3.14.1, 44KB), `modal-trap.js` (manual focus trap).

**Decision**: COPY verbatim.

**Rationale**: These are versioned, self-hosted JS libraries. The versions
are pinned and documented. The modal-trap.js implements the focus-trap
pattern without the Alpine trap plugin (avoids npm dependency chain). All
three are project-agnostic.

**No substitutions needed.** Versions documented in comments inside each
file (carried from the reference).

---

## Item 18: cmd/prospeccao/main.go + main_test.go

**Reference**: `pragmaos/cmd/pragmaos/main.go` -- but the reference version
is already complex (full router, DB pool, auth, handlers). SPEC-01 here needs
only the MINIMAL stub (the reference's SPEC-01 was gap-filling a missing
entry point; the full router came in later specs).

**Decision**: NEW (write a minimal stub, modeled on what pragmaos's SPEC-01
would have produced, not what pragmaos has now).

**Rationale**: The minimal stub makes `make build` and `make dev` work
without introducing business logic. It serves `/healthz` with `slog` logging.
The full chi router, DB pool, auth, and handlers come in SPEC-03 onward.

**Content**:
- `main.go`: `package main`, imports `net/http`, `log/slog`, `os`. `main()`
  reads `PORT` env (default `:8080`), sets up `slog` with JSON handler, registers
  `/healthz` handler on `http.NewServeMux()`, starts `http.ListenAndServe`.
  `healthHandler` returns `200 {"status":"ok"}`.
- `main_test.go`: tests `healthHandler` via `httptest.NewRecorder`, asserts
  status 200 and body `{"status":"ok"}`. Achieves ~100% coverage on the
  minimal code.

---

## Summary table

| Item | Decision | Substitutions |
|------|----------|---------------|
| Makefile | ADAPT | binary/module name |
| bootstrap.sh | ADAPT | project name, DB name |
| .pre-commit-config.yaml | COPY | none |
| .ast-grep/rules/ (7) | COPY | none |
| sgconfig.yml | COPY | none |
| sqlc.yaml | COPY | none |
| tailwind.config.js | NEW | full token set from design spec |
| input.css | ADAPT | directives only, defer components to SPEC-04 |
| package.json | ADAPT | name |
| .env.example | ADAPT | remove AI/WhatsApp/CNJ, change issuer + DB name |
| ci.yml | ADAPT | binary path |
| .devin configs | ADAPT | AST_GREP_CONFIG path |
| extensions.yml | COPY | none |
| spec-template-slim.md | COPY | none |
| AGENTS.md | NEW | remove AI/legal, add single-user scope |
| constitution.md | NEW | replace principle VI with Simplicity for Single-User |
| static/js/ (3 files) | COPY | none |
| cmd/prospeccao/main.go | NEW | minimal stub (healthz + slog) |
