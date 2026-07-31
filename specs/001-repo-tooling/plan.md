# Implementation Plan: SPEC-01 -- Repo Tooling & Dev Environment

**Branch**: `001-repo-tooling` | **Date**: 2026-07-31 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-repo-tooling/spec.md`

## Summary

SPEC-01 transplants the proven tooling stack from the PragmaOS reference
project (`/Users/relterborges/Documents/Dev/pragmaos`) and adapts it to the
Prospeccao Brasil domain (commercial real-estate prospection & expansion).
The deliverable is a reproducible dev environment (`make setup`), deterministic
quality gates (`make check`), green CI, agent-ready documentation (AGENTS.md +
constitution), and the design-system tooling foundation (Tailwind tokens for
the premium "Real Estate Intelligence" brand). The approach is copy-and-adapt:
the tooling files (Makefile, bootstrap, pre-commit, ast-grep rules, sqlc
config, CI workflow, .devin configs, extensions.yml, slim spec template) are
copied near-verbatim with project-name and domain substitutions; the
AGENTS.md, constitution, and Tailwind config are written new because the
domain (real estate, single-user MVP, premium aesthetic) differs from the
reference (law firms, multi-tenant, sober/dense UI). SPEC-01 adds a minimal
Go entry point (`cmd/prospeccao/main.go` with `/healthz` + `slog`) so `make
build` and `make dev` work, plus self-hosted HTMX/Alpine JS. No business logic,
no schema, no auth, no UI components -- those are SPEC-02 onward.

## Technical Context

**Language/Version**: Go 1.26+ (go.mod declares `go 1.26.0`). The bootstrap
script enforces Go 1.26+ (required for stdlib CVE fixes per govulncheck).

**Primary Dependencies**: None at the application level yet. SPEC-01 adds no
third-party Go dependencies -- the minimal `cmd/prospeccao/main.go` uses only
the Go standard library (`net/http`, `log/slog`). Tooling dependencies
(golangci-lint, sqlc, migrate, ast-grep, pre-commit, gitleaks,
govulncheck) are external CLI tools, not Go module dependencies. Tailwind CSS
is a build-time-only npm devDependency (pinned `3.4.17`), not a runtime
dependency. HTMX (`1.9.12`) and Alpine.js (`3.14.1`) are self-hosted static
files, not npm packages.

**Storage**: PostgreSQL 16+ (dev database `prospeccaobrasil`). SPEC-01 creates
the `migrations/` directory with a `.gitkeep` but writes NO schema migrations
(deferred to SPEC-02). The dev database is created by
`scripts/bootstrap.sh` but no tables exist at the end of SPEC-01.

**Testing**: `go test` with `-race -coverprofile -p 1`. The 85% coverage gate
(Constitution principle IV) is documented and enforced from SPEC-02 onward
once there is business code. SPEC-01 adds a minimal test for the minimal
`cmd/prospeccao/main.go` to satisfy the gate on the minimal code. Shell
scripts are tested by execution in `make setup` and CI, not by unit tests.

**Target Platform**: macOS (dev) and Linux (CI, ubuntu-latest). The dev
environment runs natively; no Docker/K8s required for the MVP.

**Project Type**: Server-rendered web monolith (Go + HTMX + Postgres). SPEC-01
delivers only the entry point and tooling; no HTTP routes beyond `/healthz`,
no templates, no business logic. The same binary will later serve both the
public site (SPEC-05) and the protected sistema interno (SPEC-06).

**Performance Goals**: N/A for SPEC-01 (no business logic). The only
performance-related target is that `make check` completes in under 30 seconds
on a clean repo.

**Constraints**:
- Single Go binary + Postgres. No Turborepo, no npm workspaces, no SPA, no
  React, no Docker/K8s for dev.
- No secrets in repo (gitleaks enforced).
- No emojis anywhere -- code, UI, comments, docs, commits. Strict.
- Structured logging via `slog` (no `fmt.Println` in non-main code).
- Forward-only migrations (golang-migrate); SPEC-01 creates the directory
  only, no schema.
- Self-hosted JS (HTMX, Alpine) -- no CDN (SPOF avoidance).
- CI parity: Makefile and `.github/workflows/ci.yml` must stay in sync
  (`-p 1` flag, `ENCRYPTION_KEY` env var where applicable).
- 85% test coverage on new Go code (enforced from SPEC-02; SPEC-01's minimal
  stub is fully tested).

**Scale/Scope**: ~1 Go file (`cmd/prospeccao/main.go`) + 1 test file + 7
ast-grep rules + 1 sqlc config + 1 Tailwind config + 1 Makefile + 1 bootstrap
script + 1 pre-commit config + 1 CI workflow + 1 AGENTS.md + 1 constitution +
1 extensions.yml + 1 slim spec template + 2 .devin configs + 1 .env.example +
1 .gitignore + 2 self-hosted JS files + 1 migrations/.gitkeep. Total new Go
code: ~30-50 lines. Total config/docs: ~15 files.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

NOTE: The constitution is itself a deliverable of SPEC-01 (FR-010). The check
below is against the constitution as designed in this spec (the 7 principles
listed in spec.md Goals/Constraints). The constitution file is written during
implementation; this check validates the design is self-consistent.

| # | Principle | Status | Notes |
|---|-----------|--------|-------|
| I | Spec-Driven Development | PASS | This spec follows the lifecycle; no implementation before plan + tasks. The slim template is shipped for future infra specs. |
| II | Security-First Design (LGPD) | PASS | No client/property data in SPEC-01. gitleaks pre-commit enforces no secrets. `.env.local` gitignored. No PII. The tenant_id rule ships now but fires only once SPEC-03 adds the column. |
| III | Single-Binary & Tooling Consistency | PASS | Single Go binary. `make check` orchestrates all gates. No `/api/v1/` endpoints in SPEC-01 (only `/healthz`). |
| IV | Test-First & Continuous Quality | PASS | 85% coverage gate documented and wired in CI. SPEC-01 adds a minimal test to satisfy the gate on the minimal code. govulncheck runs in CI. |
| V | Observability & Structured Logging | PASS | `cmd/prospeccao/main.go` uses `slog` (JSON handler in prod, text in dev). No `fmt.Println` in non-main code (ast-grep rule `go-slog-fmt.yml` enforces). `/healthz` endpoint included. |
| VI | Forward-Only Migrations | PASS | SPEC-01 creates `migrations/` directory with `.gitkeep` only. No schema migrations (SPEC-02). `migrate up` on empty dir is a no-op pass. |
| VII | Simplicity for Single-User | PASS | SPEC-01 ships the auth/tenant encanamento rule (ast-grep) but does NOT build multi-user UI. The MVP scope (1 tenant, 1 admin user) is preserved. No premature complexity. |

**Gate result: PASS.** No violations. No complexity tracking entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/001-repo-tooling/
├── spec.md                # Created by /speckit-specify
├── checklists/
│   └── requirements.md    # Spec quality checklist
├── plan.md                # This file
├── research.md            # Phase 0 output -- transplant decisions
├── data-model.md          # Phase 1 output -- no data entities (tooling spec)
├── quickstart.md          # Phase 1 output -- validation guide
└── tasks.md               # Phase 2 output (/speckit-tasks -- NOT created yet)
```

### Source Code (repository root)

```text
prospeccao-brasil/
├── cmd/
│   └── prospeccao/
│       ├── main.go            # NEW: minimal entry point (healthz + slog)
│       └── main_test.go       # NEW: test for the entry point (85% coverage)
├── internal/                  # empty for now; SPEC-02+ adds packages
│   └── db/
│       └── queries/
│           └── .gitkeep       # NEW: empty dir so sqlc + git track it
├── migrations/
│   └── .gitkeep               # NEW: empty dir so migrate + git track it
├── static/
│   ├── css/
│   │   └── app.css            # NEW: Tailwind build output (build-css target)
│   └── js/
│       ├── htmx.min.js        # NEW: self-hosted HTMX v1.9.12
│       ├── alpine.min.js      # NEW: self-hosted Alpine.js v3.14.1
│       └── modal-trap.js      # NEW: manual focus trap (no Alpine trap plugin)
├── scripts/
│   └── bootstrap.sh           # NEW: one-command dev setup (adapted from pragmaos)
├── .ast-grep/rules/           # NEW: 7 rules (adapted from pragmaos)
│   ├── go-bare-error.yml
│   ├── go-handler-missing-auth.yml
│   ├── go-hardcoded-secret.yml
│   ├── go-missing-context.yml
│   ├── go-missing-tenant-filter.yml
│   ├── go-slog-fmt.yml
│   └── tmpl-bare-button.yml
├── .github/workflows/
│   └── ci.yml                 # NEW: CI parity with Makefile
├── .devin/
│   ├── config.json            # NEW: ast-grep MCP server config
│   └── mcp_config.json        # NEW: same (Devin MCP registry)
├── .specify/
│   ├── memory/
│   │   └── constitution.md    # NEW: 7 principles (adapted)
│   ├── extensions.yml         # NEW: review/spectest/tekimax-security hooks
│   └── templates/
│       └── spec-template-slim.md  # NEW: slim template (copied from pragmaos)
├── Makefile                   # NEW: setup/dev/check/lint/test/build-css/build/migrate/sqlc/fmt/ast-grep/run
├── sqlc.yaml                  # NEW: postgresql + pgx/v5 config
├── sgconfig.yml               # NEW: ast-grep config
├── tailwind.config.js         # NEW: Prospeccao Brasil token set (Deep Navy, Gold)
├── input.css                  # NEW: Tailwind directives
├── package.json               # NEW: tailwindcss devDependency (build-time only)
├── .pre-commit-config.yaml    # NEW: gitleaks + hooks + local Go hooks
├── .env.example               # NEW: placeholder env vars
├── .gitignore                 # NEW: ignore .env.local, bin/, node_modules/, etc.
├── go.mod                     # NEW: module prospeccaobrasil, go 1.26.0
├── go.sum                     # NEW: (empty or minimal, no third-party deps yet)
├── AGENTS.md                  # NEW: agent rules (adapted from pragmaos)
└── README.md                  # EXISTS: from git init
```

**Structure Decision**: Single Go binary layout (`cmd/prospeccao/` entry point
+ `internal/` for future packages + `migrations/` for SQL + `sqlc.yaml` for
code generation config + `static/` for self-hosted assets). This mirrors the
PragmaOS reference project's structure (Constitution principle III: single
binary) and supports the decision to serve both the public site and the
protected sistema interno from the same binary. No `src/`, no `apps/` split,
no separate frontend/backend -- Prospecção Brasil is one binary.

## Complexity Tracking

> No constitution violations. Table left empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none)    |            |                                      |
