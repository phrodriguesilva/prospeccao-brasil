# Feature Specification (Slim): SPEC-01 -- Repo Tooling & Dev Environment

**Feature Branch**: `001-repo-tooling`

**Created**: 2026-07-31

**Status**: Draft

**Template**: slim (for infrastructure/tooling specs). See AGENTS.md
"Spec template selection" for when to use this vs the full template.

**Input**: User description: "SPEC-01: Repo Tooling & Dev Environment -- Establish the Go + HTMX + Postgres 16 monolith scaffold for Prospeccao Brasil. Transplant the tooling stack from the PragmaOS reference project: Makefile orchestrating all quality gates, pre-commit hooks (gitleaks, gofmt, goimports, golangci-lint, ast-grep), CI parity, ast-grep structural rules, sqlc config, Tailwind build-time config, self-hosted HTMX/Alpine JS, one-command dev setup (scripts/bootstrap.sh), AGENTS.md, constitution, .devin configs (config.json, mcp_config.json), and the slim spec template. This spec validates the tooling foundation that every subsequent spec (SPEC-02 schema, SPEC-03 auth, SPEC-04 design system, SPEC-05 site, SPEC-06 sistema interno) depends on."

## Overview

SPEC-01 establishes the engineering foundation that every subsequent spec
depends on. Without a reproducible dev environment, deterministic quality
gates, green CI, and agent-ready documentation, no other feature can be built
safely. This is the "zeroth" spec: it does not deliver product value to Luiz
Claudio (the company owner) or site visitors directly, but it is the
prerequisite that makes all product specs (SPEC-02 onward) buildable,
testable, and shippable.

The spec transplants the proven tooling stack from the PragmaOS reference
project (a Go + HTMX + Postgres monolith for law-firm management) and adapts
it to the Prospeccao Brasil domain (commercial real-estate prospection &
expansion). The stack is identical in tooling (Go, chi, pgx, sqlc,
golang-migrate, Tailwind, HTMX, Alpine, ast-grep, pre-commit, gitleaks,
govulncheck, speckit, Pencil, Graphify) but the design system, constitution,
and AGENTS.md are new -- tailored to the premium "Real Estate Intelligence"
brand and the single-user (Luiz Claudio) MVP scope, with the auth/tenant
encanamento ready for future commercialization.

## Context

This spec is the entry point of the Spec-Driven Development roadmap. It is
the first spec and depends on nothing (foundation spec).

**Canonical sources:**

- Reference project (tooling stack to transplant):
  `/Users/relterborges/Documents/Dev/pragmaos` -- AGENTS.md, Makefile,
  scripts/bootstrap.sh, .pre-commit-config.yaml, sgconfig.yml, sqlc.yaml,
  tailwind.config.js, input.css, package.json, .env.example, .ast-grep/rules/,
  .github/workflows/ci.yml, .devin/config.json, .devin/mcp_config.json,
  .specify/extensions.yml, .specify/templates/spec-template-slim.md.
- Current site (content reference): https://prospeccaobrasil.com.br/
- Reference sites (design inspiration): plda.com.br,
  rrnegociosimobiliariosadm.com.br, ocupantes.com.br, amplitudere.com.br,
  gruposinop.com.br.
- Design system spec (provided by user): Deep Navy primary, Sóbrio Gold
  secondary, Montserrat/Inter typography, soft radius, ambient shadows,
  80px section gaps, premium minimalist corporate aesthetic.
- Constitution (7 principles, to be created in this spec):
  [.specify/memory/constitution.md](../../.specify/memory/constitution.md).

**Dependencies**: None (foundation spec).
**Gate to run**: Clean clone of the repo on a machine with Go 1.26+ and
Postgres 16+ available.

## Goals

1. **Reproducible dev environment:** A new contributor (or the lead developer
   on a fresh machine) clones the repo and runs `make setup` to get a working
   dev environment (Go toolchain checked, Postgres dev database created, env
   file copied, quality gates runnable).
2. **Deterministic quality gates:** `make check` runs the full quality gate
   suite (golangci-lint, go test, go build, ast-grep scan, build-css) and
   exits non-zero on any failure. The same gates run in CI and in pre-commit.
3. **Green CI:** A trivial commit passes all CI checks.
4. **Constitution enforcement at the tooling layer:** ast-grep rules and
   pre-commit hooks encode the constitution's constraints (no secrets, no
   bare errors, no `fmt.Println` in non-main code, handler auth presence, no
   hardcoded secrets, no bare buttons in templates) so that violations are
   caught deterministically. The multi-tenant filter rule is included but
   scoped to apply once SPEC-03 adds tenant_id (the rule file ships now; it
   fires on queries once the schema has tenant_id).
5. **Agent readiness:** AGENTS.md and the constitution are present and
   referenced, so any AI coding agent (Devin, Claude, Cursor) working in this
   repo follows the project's conventions from the first commit. The slim
   spec template is available for infrastructure specs.
6. **Design system tooling ready:** Tailwind config with the Prospeccao Brasil
   token set (Deep Navy, Sóbrio Gold, Montserrat/Inter, soft radius, ambient
   shadows) is in place so SPEC-04 (design system) and SPEC-05 (site) can
   build UI on top of it. The `build-css` target compiles Tailwind to
   `static/css/app.css`.
7. **Self-hosted JS:** HTMX and Alpine.js are self-hosted in `static/js/`
   (no CDN) per the reference project's SPOF-avoidance policy.

## Non-Goals

The following are explicitly deferred to later specs and MUST NOT be
implemented in SPEC-01:

- **Database schema and migration content** -> SPEC-02 (Database Schema &
  Migrations). SPEC-01 only creates the `migrations/` directory and ensures
  the `migrate` tool is installed; it does NOT write the initial schema.
- **Auth, sessions, 2FA TOTP, RBAC, tenant middleware** -> SPEC-03. SPEC-01
  ships the ast-grep rule for tenant filters but does not implement auth.
- **Design system components and Pencil design files** -> SPEC-04. SPEC-01
  ships the Tailwind token config only; component classes and .pen files are
  SPEC-04.
- **Site institucional pages** (Home, Quem somos, Servicos, Nossos clientes,
  Fale Conosco, Newsletter) -> SPEC-05.
- **Sistema interno** (imoveis, clientes, prospecoes CRUD, PDF generation via
  chromedp) -> SPEC-06.
- **Production deployment** (TLS, backups, hosting) -> future spec.
- **Docker/K8s** -- not required for the MVP dev environment; the dev
  environment runs natively on macOS/Linux with Go + Postgres installed
  locally.

## Requirements

These requirements are the verifiable acceptance criteria. Each is
non-negotiable.

- **FR-001**: `go.mod` exists with module `prospeccaobrasil` and `go 1.26.0`,
  and `go build ./...` succeeds (or fails only because no .go files exist
  yet, in which case a placeholder `cmd/prospeccao/main.go` with a `package
  main` declaration exists so the build target is valid).
- **FR-002**: `Makefile` exists with targets: `setup`, `dev`, `check`, `lint`,
  `test`, `build-css`, `build`, `migrate`, `migrate-down`, `sqlc`, `fmt`,
  `ast-grep`, `run`. `make check` runs lint + test + build-css + build +
  ast-grep and exits 0 on a clean repo.
- **FR-003**: `scripts/bootstrap.sh` exists, is executable, checks
  prerequisites (go, psql, golangci-lint, sqlc, migrate, ast-grep,
  pre-commit, gitleaks, gh, node, npm), creates `.env.local` from
  `.env.example`, creates the dev database `prospeccaobrasil`, runs any
  existing migrations, installs pre-commit hooks, and runs `make check`.
  Verified by `make setup` exiting 0 on a machine with all tools installed.
- **FR-004**: `.pre-commit-config.yaml` exists with gitleaks, pre-commit-hooks
  (trailing-whitespace, end-of-file-fixer, check-yaml, check-json,
  check-merge-conflict, check-added-large-files, detect-private-key), and
  local hooks (gofmt, go-imports, golangci-lint, ast-grep). Verified by
  `pre-commit run --all-files` exiting 0 (or only flagging files that
  genuinely violate, on a clean repo it exits 0).
- **FR-005**: `.ast-grep/rules/` contains the adapted rules:
  `go-bare-error.yml`, `go-handler-missing-auth.yml`, `go-hardcoded-secret.yml`,
  `go-missing-context.yml`, `go-missing-tenant-filter.yml`, `go-slog-fmt.yml`,
  `tmpl-bare-button.yml`. `sgconfig.yml` points to `.ast-grep/rules`.
  Verified by `ast-grep scan` exiting 0 on a clean repo.
- **FR-006**: `sqlc.yaml` exists configured for postgresql, pgx/v5, queries in
  `internal/db/queries/`, schema in `migrations/`, emitting to `internal/db/`
  with interface and pointers for null types. Verified by `sqlc generate`
  succeeding (or no-oping cleanly with no queries yet).
- **FR-007**: `tailwind.config.js` exists with the Prospeccao Brasil token
  set: Deep Navy primary (`#031636`/`#1a2b4c`), Sóbrio Gold secondary
  (`#765a1a`/`#b89650`), surface (`#fcf9f8`), slate-gray (`#334155`),
  whatsapp-green (`#25D366`), Montserrat + Inter font families, soft radius
  (sm 0.125rem, DEFAULT 0.25rem, md 0.375rem, lg 0.5rem), ambient shadow
  tokens. `input.css` has the Tailwind directives. `package.json` has
  tailwindcss devDependency (pinned, not `@latest`). Verified by
  `make build-css` producing `static/css/app.css`.
- **FR-008**: `static/js/htmx.min.js` and `static/js/alpine.min.js` exist
  (self-hosted, no CDN references). Verified by `ls static/js/` showing both
  files and `grep -r "cdn\|unpkg\|jsdelivr" static/ internal/` returning no
  matches.
- **FR-009**: `AGENTS.md` exists at repo root, adapted from the PragmaOS
  AGENTS.md to the Prospeccao Brasil domain: removes law-firm/AI/CNJ/WhatsApp/
  LLM sections; keeps speckit lifecycle, ast-grep, sqlc, tailwind, pencil,
  self-host JS, quality gates, CI parity, spec template selection (slim vs
  full), constitution reference, key commands, MCP servers table. Verified by
  the file existing and containing sections: "Project Overview",
  "Spec-Driven Development", "Code Conventions", "Quality Gates",
  "Frontend Conventions", "Key Commands".
- **FR-010**: `.specify/memory/constitution.md` exists with 7 principles
  adapted to Prospeccao Brasil: I. Spec-Driven Development, II. Security-First
  Design (LGPD for client/property data, 2FA, tenant_id), III. Single-Binary
  & Tooling Consistency, IV. Test-First & Continuous Quality (85% coverage),
  V. Observability & Structured Logging (slog), VI. Forward-Only Migrations,
  VII. Simplicity for Single-User (UI must not be complex for Luiz Claudio;
  future-proof encanamento without premature multi-user UI). Includes a
  Quality Gates table. Verified by the file existing with all 7 principles
  and the Quality Gates table.
- **FR-011**: `.specify/extensions.yml` exists with the same hook structure as
  the reference project: review, spectest, tekimax-security extensions
  installed; after_implement hooks (review, spectest gaps, security audit,
  ast-grep scan); after_specify (data-contract, optional); after_plan
  (threat-model, optional); before_implement (gate-check, mandatory);
  before_analyze (red-team, optional). Verified by the file existing and
  being valid YAML (`pre-commit check-yaml` or `python -c "import yaml;
  yaml.safe_load(open('.specify/extensions.yml'))"` succeeds).
- **FR-012**: `.specify/templates/spec-template-slim.md` exists (copied from
  the reference project, adapted placeholders). Verified by the file
  existing.
- **FR-013**: `.devin/config.json` and `.devin/mcp_config.json` exist with the
  ast-grep MCP server configured, pointing to this repo's `sgconfig.yml`.
  Verified by both files existing and being valid JSON.
- **FR-014**: `.github/workflows/ci.yml` exists with CI parity to the Makefile
  (golangci-lint, go test with `-p 1` and `ENCRYPTION_KEY` env var, go build,
  ast-grep scan, govulncheck). Verified by the file existing and a trivial
  commit passing CI (or, if GitHub repo not yet pushed, by `yamllint` /
  manual review confirming parity with Makefile targets).
- **FR-015**: `.env.example` exists with: `DATABASE_URL`,
  `SESSION_SECRET`, `TOTP_ISSUER=Prospeccao Brasil`, `ENCRYPTION_KEY`,
  `RATE_LIMIT_PER_IP`, `RATE_LIMIT_PER_EMAIL`, `RATE_LIMIT_WINDOW`,
  `SMTP_HOST/PORT/USER/PASS`, `APP_BASE_URL`. No real secrets. Verified by
  the file existing and `gitleaks detect --source . --no-git` exiting 0.
- **FR-016**: `.gitignore` exists ignoring: `.env.local`, `bin/`, `node_modules/`,
  `coverage.out`, `coverage-nogen.out`, `*.pen` (if binary), the built binary
  `prospeccaobrasil`, `.DS_Store`. Verified by the file existing.
- **FR-017**: `git init` has been run, the repo is a git repository, and an
  initial commit exists with the scaffold. Verified by `git rev-parse
  --is-inside-work-tree` exiting 0 and `git log --oneline` showing at least
  one commit.

## Constraints

1. **No secrets in repo:** `.env.example` has placeholder values only;
   `.env.local` is gitignored. Gitleaks pre-commit hook enforces.
2. **No emojis anywhere:** code, UI, comments, docs, commits. Strict (carried
   from the reference project's constitution).
3. **Single binary:** Prospeccao Brasil is one Go binary + Postgres. No
   Turborepo, no npm workspaces, no SPA, no React. Tailwind is build-time
   only (not a runtime dependency).
4. **Self-hosted JS:** HTMX and Alpine.js must be in `static/js/`, never
   loaded from a CDN. Pinned versions, documented in a comment.
5. **CI parity:** The CI workflow and the Makefile MUST stay in sync. The
   `-p 1` flag and `ENCRYPTION_KEY` env var (if applicable to test runs) are
   required in both places.
6. **Forward-only migrations:** `migrations/` directory created but empty (no
   schema yet -- that is SPEC-02). The `migrate` tool is installed by
   bootstrap.
7. **85% coverage gate:** Enforced from SPEC-02 onward once there is code to
   test. SPEC-01 itself has no business code, so the coverage gate is
   documented but not yet enforced on zero-code.
8. **Go 1.26+:** Required for stdlib CVE fixes (govulncheck). Bootstrap checks
   the version.
9. **Postgres 16+:** Required. Bootstrap checks the version.

## Definition of Done

SPEC-01 (Repo Tooling & Dev Environment) is done when ALL of the following
are verifiable:

| # | FR | Acceptance Criterion | Verification Command | Status |
|---|----|----------------------|----------------------|--------|
| 1 | FR-001 | Go module builds | `go build ./...` exits 0 | [ ] |
| 2 | FR-002 | All Makefile targets exist | `make -n check lint test build-css build migrate migrate-down sqlc fmt ast-grep run dev setup` exits 0 | [ ] |
| 3 | FR-002 | `make check` passes on clean repo | `make check` exits 0 | [ ] |
| 4 | FR-003 | Bootstrap sets up dev env | `make setup` exits 0 (on a machine with all tools) | [ ] |
| 5 | FR-004 | Pre-commit hooks installed and pass | `pre-commit run --all-files` exits 0 | [ ] |
| 6 | FR-005 | ast-grep rules scan clean | `ast-grep scan` exits 0 | [ ] |
| 7 | FR-006 | sqlc config valid | `sqlc generate` exits 0 (or clean no-op) | [ ] |
| 8 | FR-007 | Tailwind builds | `make build-css` produces `static/css/app.css` | [ ] |
| 9 | FR-008 | Self-hosted JS present, no CDN | `ls static/js/htmx.min.js static/js/alpine.min.js` and `grep -r "cdn\|unpkg\|jsdelivr" static/ internal/` returns nothing | [ ] |
| 10 | FR-009 | AGENTS.md present with required sections | `grep -c "Project Overview\|Spec-Driven Development\|Code Conventions\|Quality Gates\|Frontend Conventions\|Key Commands" AGENTS.md` >= 6 | [ ] |
| 11 | FR-010 | Constitution present with 7 principles | `grep -c "^### " .specify/memory/constitution.md` >= 7 | [ ] |
| 12 | FR-011 | extensions.yml valid | `python3 -c "import yaml; yaml.safe_load(open('.specify/extensions.yml'))"` exits 0 | [ ] |
| 13 | FR-012 | Slim spec template present | `test -f .specify/templates/spec-template-slim.md` exits 0 | [ ] |
| 14 | FR-013 | .devin configs present and valid JSON | `python3 -c "import json; json.load(open('.devin/config.json')); json.load(open('.devin/mcp_config.json'))"` exits 0 | [ ] |
| 15 | FR-014 | CI workflow present | `test -f .github/workflows/ci.yml` exits 0 | [ ] |
| 16 | FR-015 | .env.example present, no secrets | `gitleaks detect --source . --no-git` exits 0 | [ ] |
| 17 | FR-016 | .gitignore present | `test -f .gitignore` exits 0 | [ ] |
| 18 | FR-017 | Repo is a git repo with initial commit | `git rev-parse --is-inside-work-tree && git log --oneline -1` exits 0 | [ ] |

**Spec is ready for `/speckit-implement` when all rows are checked.**
