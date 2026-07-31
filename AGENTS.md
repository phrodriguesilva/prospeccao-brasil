# AGENTS.md -- AI Agent Rules for Prospecção Brasil

This file is read by any AI coding agent (Devin, Claude, Cursor) working in this
repo. Follow these conventions for every change.

## Project Overview

Prospecção Brasil is a commercial real-estate prospecting platform. It is a
Go + HTMX + Postgres server-rendered monolith with a premium institutional
front-end and a simple internal system for property and client management.
The goal is to reduce the cognitive load on the prospector (Luiz Claudio):
the software manages properties, clients, and prospecting; the prospector
focuses on relationships and deals.

**MVP scope**: single-tenant, single-admin user. The backend is future-proofed
for multi-tenant (tenant_id, RBAC middleware, session + 2FA) but the UI is
minimal -- one admin manages everything. No premature multi-user complexity.

## Spec-Driven Development

Every feature follows the Spec Kit lifecycle:

1. `/speckit-specify SPEC-XX: Title` -> `spec.md`
2. `/speckit-plan` -> `plan.md`
3. `/speckit-tasks` -> `tasks.md`
4. `/speckit-analyze` -> cross-artifact consistency
5. `/speckit-implement` -> code + tests per task

One spec at a time. Finish four stages, verify acceptance, trigger next. Spec
directory naming: `specs/001-<slug>`, `002-<slug>`, ... (sequential).
`.specify/feature.json` points to the active feature directory via the
`feature_directory` key (Spec Kit 0.12.9 schema; do not use `feature_dir`,
`feature_number`, or `feature_title` -- they are not read by the runtime).

## Spec template selection

Two spec templates exist in `.specify/templates/`:

- **`spec-template.md`** (full): user stories with Gherkin acceptance scenarios,
  edge cases, key entities, success criteria, assumptions. Use for product
  specs that deliver user value.
- **`spec-template-slim.md`** (slim): overview, context, goals, non-goals,
  requirements, constraints, Definition of Done. Use for infrastructure/tooling
  specs that deliver engineering value, not user stories.

Use the **slim** template for: SPEC-01 (repo tooling), SPEC-02 (database
schema), SPEC-03 (auth middleware).

Use the **full** template for everything else (SPEC-04 institutional site &
design system, SPEC-05 internal system). When in doubt, default to full -- a
slim spec that grows user-facing behavior mid-implementation is a sign it
should have been full.

Note: SPEC-04 (Design System) and SPEC-05 (Institutional Site) were merged
into a single spec "SPEC-04: Institutional Site & Design System" to deliver
user-facing value sooner and avoid a design system with no consumer. The
original SPEC-06 (Sistema interno) became SPEC-05 in the new numbering.

## Constitution

`.specify/memory/constitution.md` defines the 7 principles. Key constraints:

- 85% test coverage minimum.
- No secrets in repo (gitleaks enforces).
- Forward-only SQL migrations (golang-migrate).
- Structured logs via `slog` (no `fmt.Println` in non-main code).
- Multi-tenant isolation: `tenant_id` filter on every query; cross-tenant
  access is a critical bug (encanamento ships now, fires when SPEC-03 adds
  the column).
- No emojis anywhere -- code, UI, comments, docs, commits. Strict.

## Code Conventions

### Go
- `gofmt` + `golangci-lint` (replaces ruff/black/mypy/eslint/prettier).
- `sqlc` for SQL -> typed Go. No ORM. SQL is the source of truth.
- Error wrapping with `fmt.Errorf("...: %w", err)`. No bare `return err`.
- Context propagation: `ctx context.Context` as first param on DB/HTTP work.

### Templates
- Go `html/template` or `templ`; premium institutional design with soft corners,
  subtle shadows, Montserrat headlines, Inter body. Tailwind with the
  Prospecção Brasil brand tokens (see `tailwind.config.js`).
- No emojis anywhere (strict).

### API
- URL path versioning `/api/v1/` only where JSON endpoints exist (most UI is
  HTML over HTMX).
- Consistent error envelope for JSON endpoints.
- Cursor pagination for lists.

### Auth
- Session cookie HttpOnly + SameSite=Strict + Secure.
- 2FA TOTP required for the admin user.
- RBAC middleware ships now (encanamento) but MVP is single-admin.
- PDF generation via `chromedp` (HTML to PDF) for property presentation
  documents (SPEC-05).

### Commits
- No conventional-commit prefixes (rejected by CI).
- No emojis anywhere -- code, UI, comments, docs, commits.

## Quality Gates

`make check` runs: golangci-lint + go test + build-css + go build + ast-grep scan.

### CI parity

The CI workflow (`.github/workflows/ci.yml`) and the Makefile MUST stay in
sync. If you change a test flag in one, change it in the other. The
`-p 1` flag (sequential package execution) and `ENCRYPTION_KEY` env var
are required in both places -- omitting either causes CI failures that
do not reproduce locally.

### CI verification before declaring a spec done

Before marking a spec as complete, the agent MUST verify that the PR's CI
check passes (via `gh pr checks <PR#>`). A green local `make check` does NOT
guarantee green CI -- env vars, test flags, and runner differences can cause
divergent results. The completion report must include the CI check status,
not just local output.

## Frontend Conventions

- HTMX for interactivity, Alpine.js for micro-state.
- No SPA, no React. Server-rendered HTML.
- URL-synced filter state for list pages (query params).

### Self-host JS libraries (no CDN)

HTMX and Alpine.js MUST be self-hosted in `static/js/`, NOT loaded from a
CDN (unpkg, jsdelivr, etc). CDN dependencies are a SPOF -- if the CDN is down
or slow, the UI degrades. Self-hosting eliminates this risk at the cost of 2
small files (~92KB total).

- `static/js/htmx.min.js` -- HTMX core (v1.9.12, 48KB)
- `static/js/alpine.min.js` -- Alpine.js core (v3.14.1, 44KB)
- `static/js/modal-trap.js` -- manual focus trap (no Alpine trap plugin)
- Pin to a specific version (do not use `@latest`)
- Add `static/js/` to the static file server in `cmd/prospeccao/main.go`
- Update files when a new version is needed (document the version in a comment)

### Focus trap (no Alpine.js trap plugin)

The `@alpinejs/trap` plugin depends on `focus-trap` (another npm package)
and cannot be self-hosted as a standalone file. Instead, implement focus
trap manually in the modal component using Alpine.js directives:
- `x-init` to focus the first focusable element on open
- `@keydown.tab` to cycle focus within the modal
- `@keydown.escape.window` to close on Escape
- `@click.outside` is NOT used (the overlay handles click-to-close)

## Structural Code Analysis

ast-grep is the deterministic layer; rules in `.ast-grep/rules/`. `ast-grep scan`
runs as a mandatory hook before AI review. Rules catch: bare errors, missing
tenant filters, handlers without auth, `fmt.Println` in non-main code, missing
context, hardcoded secrets, bare buttons in templates.

## Visual Design

Pencil (pen.dev) for design files in `designs/`. Design is the visual source of
truth for new screens; code must match. Reference designs in specs: "see
`designs/prospeccao.pen` frame 'Home - Desktop'".

## Key Commands

- `make setup` -- one-command dev setup (runs `scripts/bootstrap.sh`).
- `make dev` -- run the dev server (`go run ./cmd/prospeccao`).
- `make check` -- lint + test + build-css + build + ast-grep scan.
- `make lint` -- golangci-lint.
- `make test` -- go test with coverage.
- `make build` -- build the binary to `bin/prospeccao`.
- `make build-css` -- build Tailwind CSS to `static/css/app.css`.

## MCP Servers & CLI Tools

### MCP servers available

The following MCP servers are available to AI agents working in this repo.

| Server | Purpose |
|--------|---------|
| `ast-grep` | Structural code search via MCP. Configured in `.devin/config.json` via `uvx`. Points to `sgconfig.yml`. |
| `deepwiki` | AI-powered documentation for GitHub repos. Use to research Go libraries. |
| `exa-code` | Web search for code-related queries. |
| `pencil` | Read/write `.pen` design files in `designs/`. Required for SPEC-04 and all UI specs. |

### CLI tools installed

All prerequisites are checked by `scripts/bootstrap.sh`. If any are missing,
`make setup` will report them.

| Tool | Install (macOS) | Purpose |
|------|-----------------|---------|
| Go 1.26+ | `brew install go` | Language toolchain (1.26 required for stdlib CVE fixes) |
| Postgres 16+ | `brew install postgresql@16` | Database |
| golangci-lint v2.12.2 | `brew install golangci-lint` | Linter (v2+ required for Go 1.26) |
| sqlc | `brew install sqlc` | SQL -> Go code generation |
| migrate | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` | Forward-only migrations |
| ast-grep | `brew install ast-grep` | Structural code analysis (7 rules in `.ast-grep/rules/`) |
| pre-commit | `brew install pre-commit` | Git hooks orchestration |
| gitleaks | `brew install gitleaks` | Secret detection (pre-commit hook) |
| gh | `brew install gh` then `gh auth login` | GitHub CLI (PRs, CI status, issues) |
| govulncheck | `go install golang.org/x/vuln/cmd/govulncheck@latest` | Vulnerability scanner (CI job) |
| goimports | `go install golang.org/x/tools/cmd/goimports@latest` | Import formatting (pre-commit hook) |
| uv | `brew install uv` | Python package manager (for ast-grep MCP server) |
| Node/npm 20+ | `brew install node` | Tailwind CSS build-time (SPEC-04+) |
