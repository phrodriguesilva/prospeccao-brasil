# Tasks: SPEC-01 -- Repo Tooling & Dev Environment

**Input**: Design documents from `/specs/001-repo-tooling/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

**Tests**: Tests ARE included -- the constitution mandates 85% coverage (principle IV), and the spec's FR-002 requires `make check` to pass (which includes the coverage gate). The minimal `cmd/prospeccao/main.go` must have a test to satisfy the coverage gate.

**Organization**: Tasks are grouped by user story (derived from the spec's Goals, since this is a slim/infra spec). Each story is independently testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single Go binary**: `cmd/prospeccao/` (entry point), `internal/` (future packages), `migrations/` (SQL), `sqlc.yaml` (config), `scripts/` (bootstrap), `.github/workflows/` (CI), `.ast-grep/rules/` (structural rules), `static/` (self-hosted assets), `.devin/` (agent configs), `.specify/` (spec kit)
- Paths are relative to repo root: `/Users/relterborges/Documents/Dev/prospeccao-brasil/`
- Reference project for COPY/ADAPT decisions: `/Users/relterborges/Documents/Dev/pragmaos/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create all foundational files that have no dependencies on each other. These are the "COPY verbatim" and "ADAPT with substitutions" items from research.md. They can all be created in parallel since they are independent files.

- [X] T001 [P] Create `go.mod` -- module `prospeccaobrasil`, `go 1.26.0`, no require blocks yet (no third-party deps). Run `go mod init prospeccaobrasil` then edit the go directive to `1.26.0`. (FR-001, research.md Item 1)
- [X] T002 [P] Create `cmd/prospeccao/main.go` -- minimal Go entry point with `/healthz` handler using `net/http` + `log/slog` (JSON handler). No third-party deps. Reads `PORT` env (default `:8080`). Registers `GET /healthz` returning `200 {"status":"ok"}`. Extract handler into a testable function `healthHandler(w http.ResponseWriter, r *http.Request)`. Uses `slog.Info` for startup log. (FR-001, research.md Item 18)
- [X] T003 [P] Create `migrations/.gitkeep` -- empty file so git tracks the `migrations/` directory. No SQL content. (FR-006, research.md Item 6)
- [X] T004 [P] Create `sqlc.yaml` -- COPY verbatim from `pragmaos/sqlc.yaml`: `version: "2"`, engine `postgresql`, queries `internal/db/queries/`, schema `migrations/`, gen.go package `db` out `internal/db/`, sql_package `pgx/v5`, emit_json_tags, emit_interface, emit_pointers_for_null_types. ALSO create `internal/db/queries/.gitkeep` (empty file) so the queries directory referenced by sqlc.yaml exists in git and `sqlc generate` does not error on a missing dir (same pattern as `migrations/.gitkeep` in T003). (FR-006, research.md Item 6)
- [X] T005 [P] Create `sgconfig.yml` -- COPY verbatim from `pragmaos/sgconfig.yml`: ruleDirs `.ast-grep/rules`, fileTypes go `["go"]` and html `["html", "tmpl", "templ"]`. (FR-005, research.md Item 5)
- [X] T006 [P] Create `.ast-grep/rules/` directory with 7 rule files -- COPY verbatim from `pragmaos/.ast-grep/rules/`: `go-bare-error.yml`, `go-handler-missing-auth.yml`, `go-hardcoded-secret.yml`, `go-missing-context.yml`, `go-missing-tenant-filter.yml`, `go-slog-fmt.yml`, `tmpl-bare-button.yml`. No substitutions (rules are Go/HTML structural patterns, not domain-specific). (FR-005, research.md Item 4)
- [X] T007 [P] Create `.gitignore` -- ignore: `.env.local`, `bin/`, `node_modules/`, `coverage.out`, `coverage-nogen.out`, `prospeccaobrasil` (built binary), `.DS_Store`, `*.pen` (if binary design files). (FR-016, research.md)
- [X] T008 [P] Create `.env.example` -- ADAPT from `pragmaos/.env.example`: `DATABASE_URL=postgres://postgres:postgres@localhost:5432/prospeccaobrasil?sslmode=disable`, `SESSION_SECRET=change-me-to-a-long-random-string`, `TOTP_ISSUER=Prospeccao Brasil`, `ENCRYPTION_KEY=` (with `openssl rand -base64 32` comment), `RATE_LIMIT_PER_IP=10`, `RATE_LIMIT_PER_EMAIL=5`, `RATE_LIMIT_WINDOW=1m`, `SMTP_HOST/PORT/USER/PASS`, `APP_BASE_URL=http://localhost:8080`. REMOVE: `WHATSAPP_API_TOKEN`, `WHATSAPP_PHONE_ID`, `CNJ_API_BASE`, `LLM_PROVIDER`, `LLM_API_KEY`, `LLM_MODEL`, `LLM_FALLBACK_MODEL` (no AI/WhatsApp/CNJ in this project). No real secrets. (FR-015, research.md Item 10)
- [X] T009 [P] Create `tailwind.config.js` -- NEW (write from scratch using the user-provided design system spec). Content paths: `./internal/ui/templates/**/*.html`, `./internal/handler/templates/**/*.html`, `./**/*.go`. Colors: `primary` (Deep Navy `#031636`/`#1a2b4c`), `secondary` (Sobrio Gold `#765a1a`/`#b89650`), `surface` (`#fcf9f8`), surface-container variants, `slate-gray` (`#334155`), `whatsapp-green` (`#25D366`), `error` (`#ba1a1a`), outline/outline-variant, inverse colors. Typography: display-lg (Montserrat 48px/700), headline-lg (32px/600), headline-md (24px/600), body-lg (Inter 18px/400), body-md (16px/400), label-sm (12px/600 uppercase tracking 0.05em). Radius: sm 0.125rem, DEFAULT 0.25rem, md 0.375rem, lg 0.5rem, xl 0.75rem, full 9999px. Shadows: ambient (blur 20px, opacity 4%, color `#1A2B4C`). Spacing: section-gap 80px, margin-mobile 20px, gutter-mobile 16px, stack-sm/md/lg 8/16/32px. Fonts: Montserrat (sans-display), Inter (sans-body). (FR-007, research.md Item 7)
- [X] T010 [P] Create `input.css` -- ADAPT from `pragmaos/input.css`: just the three Tailwind directives (`@tailwind base`, `@tailwind components`, `@tailwind utilities`). DEFER component classes (badges, buttons) to SPEC-04. (FR-007, research.md Item 8)
- [X] T011 [P] Create `package.json` -- ADAPT from `pragmaos/package.json`: `"name": "prospeccaobrasil"`, `"version": "0.1.0"`, `"private": true`, scripts `build-css` and `watch-css`, devDependency `tailwindcss` pinned at `3.4.17` (NOT `@latest`). (FR-007, research.md Item 9)
- [X] T012 [P] Create `static/js/htmx.min.js` -- COPY verbatim from `pragmaos/static/js/htmx.min.js` (v1.9.12, 48KB). Verify the version comment is preserved. (FR-008, research.md Item 17)
- [X] T013 [P] Create `static/js/alpine.min.js` -- COPY verbatim from `pragmaos/static/js/alpine.min.js` (v3.14.1, 44KB). Verify the version comment is preserved. (FR-008, research.md Item 17)
- [X] T014 [P] Create `static/js/modal-trap.js` -- COPY verbatim from `pragmaos/static/js/modal-trap.js` (manual focus trap, no Alpine trap plugin). (FR-008, research.md Item 17)
- [X] T015 [P] Create `.specify/extensions.yml` -- COPY verbatim from `pragmaos/.specify/extensions.yml`: installed (review, spectest, tekimax-security), settings (auto_execute_hooks: true), hooks (after_implement: review + spectest gaps + security audit + ast-grep scan; after_specify: data-contract optional; after_plan: threat-model optional; before_implement: gate-check mandatory; before_analyze: red-team optional). (FR-011, research.md Item 13)
- [X] T016 [P] Create `.specify/templates/spec-template-slim.md` -- COPY verbatim from `pragmaos/.specify/templates/spec-template-slim.md`. (FR-012, research.md Item 14)
- [X] T017 [P] Create `.devin/config.json` -- ADAPT from `pragmaos/.devin/config.json`: ast-grep MCP server via `uvx`, with `AST_GREP_CONFIG` env var pointing to `/Users/relterborges/Documents/Dev/prospeccao-brasil/sgconfig.yml` (substitute the path). Keep PATH env var (homebrew/go bin paths are user-global). (FR-013, research.md Item 12)
- [X] T018 [P] Create `.devin/mcp_config.json` -- ADAPT from `pragmaos/.devin/mcp_config.json`: same as T017 (ast-grep MCP server, path substitution). (FR-013, research.md Item 12)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Create the tooling files that orchestrate the quality gates and dev setup. These depend on Phase 1 files existing (Makefile references `cmd/prospeccao`, bootstrap references `.env.example` and `migrations/`, CI references Makefile targets).

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T019 Create `Makefile` -- ADAPT from `pragmaos/Makefile`: targets `setup`, `dev` (runs `go run ./cmd/prospeccao`), `check` (lint + test + build-css + build + ast-grep), `lint` (golangci-lint run), `test` (go test -race -p 1 -timeout 20m -coverprofile=coverage.out -covermode=atomic ./... + coverage exclusions for internal/db, cmd/prospeccao), `build-css` (npx tailwindcss -i input.css -o static/css/app.css --minify), `build` (go build -o bin/prospeccao ./cmd/prospeccao), `migrate`, `migrate-down`, `sqlc`, `fmt` (gofmt + goimports), `ast-grep` (ast-grep scan), `run` (build then run). Load `.env.local`. (FR-002, research.md Item 1)
- [X] T020 Create `scripts/bootstrap.sh` -- ADAPT from `pragmaos/scripts/bootstrap.sh`: 7-step setup. Substitute `PragmaOS` -> `Prospecção Brasil` in echo banners, `DB_NAME="pragmaos"` -> `DB_NAME="prospeccaobrasil"`. Keep Go 1.26+ and Postgres 16+ version checks. Keep prerequisite checks (go, psql, golangci-lint, sqlc, migrate, ast-grep, pre-commit, gitleaks, gh, node, npm). Keep graceful handling (no overwrite .env.local, no drop existing DB). Make executable (`chmod +x`). (FR-003, research.md Item 2)
- [X] T021 Create `.pre-commit-config.yaml` -- COPY verbatim from `pragmaos/.pre-commit-config.yaml`: gitleaks v8.21.2, pre-commit-hooks v5.0.0 (trailing-whitespace, end-of-file-fixer, check-yaml with exclude for .specify/extensions templates, check-json, check-merge-conflict, check-added-large-files --maxkb=500, detect-private-key), local hooks (gofmt, go-imports, golangci-lint, ast-grep). (FR-004, research.md Item 3)
- [X] T022 Create `.github/workflows/ci.yml` -- ADAPT from `pragmaos/.github/workflows/ci.yml`: jobs for golangci-lint, go test with `-p 1` and `ENCRYPTION_KEY` env var, go build (`go build -o bin/prospeccao ./cmd/prospeccao`), ast-grep scan, govulncheck. Go version pinning 1.26+. CI parity with Makefile (same flags, same env vars). (FR-014, research.md Item 11)

**Checkpoint**: All tooling files created. `make check` can now be run (may fail until npm install + go mod tidy are done -- those are in Phase 3).

---

## Phase 3: User Story 1 - Reproducible Dev Environment (Priority: P1) -- MVP

**Goal**: A new contributor can clone the repo, run `make setup`, and reach a working dev environment with all gates passing.

**Independent Test**: Run `make setup`, verify `.env.local` exists, `prospeccaobrasil` database exists, pre-commit hooks installed, `make check` passes, `make dev` starts server with `/healthz` returning 200.

### Tests for User Story 1

> The constitution mandates 85% coverage. The test for `cmd/prospeccao/main.go` is written here and MUST pass.

- [X] T023 [P] [US1] Create `cmd/prospeccao/main_test.go` -- test `healthHandler` via `httptest.NewRecorder`: send `GET /healthz`, assert status 200, assert body `{"status":"ok"}`, assert `Content-Type` contains `application/json`. Achieve >= 85% coverage on `main.go`. (FR-002, Constitution principle IV, research.md Item 18)

### Implementation for User Story 1

- [X] T024 [US1] Run `npm install` to install Tailwind CSS devDependency from `package.json`. Verify `node_modules/` is created and gitignored. (FR-007)
- [X] T025 [US1] Run `go mod tidy` to ensure `go.mod` and `go.sum` are consistent (should be minimal -- no third-party deps yet). (FR-001)
- [X] T026 [US1] Run `make build-css` and verify `static/css/app.css` is produced. Verify the brand tokens are present: `grep -q "031636\|1a2b4c" static/css/app.css` (Deep Navy), `grep -q "765a1a\|b89650" static/css/app.css` (Sobrio Gold), `grep -q "Montserrat" static/css/app.css`, `grep -q "Inter" static/css/app.css`. (FR-007, quickstart.md Scenario 9)
- [X] T027 [US1] Run `make check` and verify it passes: golangci-lint, go test (with coverage >= 85%), build-css, go build, ast-grep scan. If any gate fails, fix the underlying issue (do NOT disable the gate). (FR-002, quickstart.md Scenario 2)
- [X] T028 [US1] Run `make setup` end-to-end and verify all 7 steps pass: prerequisites OK, `.env.local` created, dev DB created (or exists), migrations no-op, sqlc generate no-op, pre-commit installed, `make check` passed. If any step fails, fix the underlying issue. (FR-003, quickstart.md Scenario 1)
- [X] T029 [US1] Verify `make dev` starts the server and `curl http://localhost:8080/healthz` returns `200 {"status":"ok"}`. If it fails, fix the underlying issue. (FR-001, quickstart.md Scenario 3)

**Checkpoint**: `make setup` works end-to-end. `make check` passes. User Story 1 is fully functional and testable independently.

---

## Phase 4: User Story 2 - Quality Gates Catch Violations (Priority: P1)

**Goal**: Pre-commit hooks (gitleaks, gofmt, goimports, golangci-lint, ast-grep) block commits that violate constitution constraints, with clear messages.

**Independent Test**: Introduce each violation type, attempt to commit, verify the commit is blocked. Revert and verify a clean commit succeeds.

### Implementation for User Story 2

- [X] T030 [US2] Run `pre-commit install` to install git hooks. Then run `pre-commit run --all-files` and verify it passes on the clean repo (all hooks: gitleaks, trailing-whitespace, end-of-file-fixer, check-yaml, check-json, check-merge-conflict, check-added-large-files, detect-private-key, gofmt, go-imports, golangci-lint, ast-grep). If any hook fails on clean code, fix the underlying issue. (FR-004, quickstart.md Scenario 5)
- [X] T031 [US2] Verify ast-grep rules scan without errors: run `ast-grep scan` (or `make ast-grep`) and confirm exit 0 with no matches. Verify all 7 rules exist in `.ast-grep/rules/`. If any rule errors (crash, not "no matches"), fix the rule YAML. (FR-005, quickstart.md Scenario 6)
- [X] T032 [US2] Violation round-trip test: for each of the 7 ast-grep rules, temporarily introduce a matching violation in a scratch file, run `ast-grep scan`, verify the rule matches and reports the file/line, then revert the scratch file. Do NOT commit the violations. (FR-005, quickstart.md Scenario 5)
- [X] T033 [US2] Verify `sqlc generate` succeeds (no-op on empty queries). Verify `sgconfig.yml` points to `.ast-grep/rules`. (FR-005, FR-006, quickstart.md Scenario 8)

**Checkpoint**: All quality gates are operational and catch violations. User Story 2 is fully functional.

---

## Phase 5: User Story 3 - CI is Green on a Trivial Commit (Priority: P1)

**Goal**: A trivial commit passes all CI checks (build, test with coverage, lint, ast-grep, govulncheck) and the PR is mergeable.

**Independent Test**: Push a trivial commit to a branch, open a PR via `gh`, verify all CI checks pass.

### Implementation for User Story 3

- [X] T034 [US3] Verify `.github/workflows/ci.yml` runs all required jobs: Go build, test with coverage (85% gate), lint (golangci-lint), ast-grep scan, govulncheck. Confirm CI parity with Makefile (same `-p 1` flag, same `ENCRYPTION_KEY` env var). If any job is missing or misconfigured, fix it. (FR-014, quickstart.md Scenario 4)
- [X] T035 [US3] Verify CI handles the empty-migrations case: the migrate step runs on a dir with only `.gitkeep`. If golang-migrate errors on a dir with no `.sql` files, add a guard. (FR-014, research.md)
- [X] T036 [US3] Verify CI handles the empty-sqlc-queries case: `sqlc generate` with `sqlc.yaml` pointing to `internal/db/queries/` (which contains only `.gitkeep` from T004). Confirm `sqlc generate` exits 0 (no-op on empty queries). (FR-014, research.md)
- [X] T037 [US3] Commit all SPEC-01 files, push to a branch, open a PR via `gh pr create`, and verify all CI checks pass on GitHub using `gh pr checks --watch`. If any check fails, fix the underlying issue and re-push. (FR-014, quickstart.md Scenario 4)

**Checkpoint**: CI is green. User Story 3 is fully functional.

---

## Phase 6: User Story 4 - Agent Readiness & Constitution (Priority: P2)

**Goal**: AGENTS.md, constitution, extensions.yml, slim template, and .devin configs are present so any AI coding agent follows project conventions from the first commit.

**Independent Test**: Read AGENTS.md, verify it references the constitution, verify the constitution has 7 principles, verify extensions.yml is valid YAML, verify .devin configs are valid JSON.

### Implementation for User Story 4

- [X] T038 [P] [US4] Create `AGENTS.md` -- NEW (write from scratch, modeled on `pragmaos/AGENTS.md` structure but adapted to Prospecção Brasil). Sections to keep (adapted): Project Overview (commercial real-estate prospection, single-user MVP for Luiz Claudio, future-proof encanamento), Spec-Driven Development, Spec template selection (slim vs full), Code Conventions (Go, Templates, API, Auth, Commits), Quality Gates, CI parity, Frontend Conventions (self-host JS, focus trap), Structural Code Analysis, Visual Design (Pencil), Documentation, Key Commands, MCP Servers & CLI Tools. Sections to REMOVE: AI/CNJ/WhatsApp/LLM references, multi-tenant RBAC role list (socio/advogado/etc), integration test isolation tech debt note, Spec Kit Hooks skip policy (law-firm-specific). Sections to ADD: Single-User MVP scope note, PDF generation note (chromedp, SPEC-06). Must reference `.specify/memory/constitution.md`. (FR-009, research.md Item 15)
- [X] T039 [P] [US4] Create `.specify/memory/constitution.md` -- NEW (write from scratch, 7 principles adapted to Prospecção Brasil). Principles: I. Spec-Driven Development, II. Security-First Design (LGPD for client/property data, 2FA, tenant_id encanamento), III. Single-Binary & Tooling Consistency, IV. Test-First & Continuous Quality (85% coverage), V. Observability & Structured Logging (slog), VI. Forward-Only Migrations, VII. Simplicity for Single-User (UI must not be complex for Luiz Claudio; future-proof encanamento without premature multi-user UI; YAGNI for MVP). Include Quality Gates table (Format, Lint, Vet, Test+coverage, Build, Structural scan, Secret scan, Dependency vuln, Spec Kit review, Security audit, Security gate). Version 1.0.0, ratified 2026-07-31. (FR-010, research.md Item 16)
- [X] T040 [US4] Verify `.specify/extensions.yml` (created in T015) is valid YAML: `python3 -c "import yaml; yaml.safe_load(open('.specify/extensions.yml'))"`. Verify it has the review, spectest, tekimax-security extensions and all hooks. (FR-011, quickstart.md Scenario 11)
- [X] T041 [US4] Verify `.specify/templates/spec-template-slim.md` (created in T016) exists and has the slim template structure. (FR-012, quickstart.md Scenario 11)
- [X] T042 [US4] Verify `.devin/config.json` and `.devin/mcp_config.json` (created in T017, T018) are valid JSON and the `AST_GREP_CONFIG` path points to `/Users/relterborges/Documents/Dev/prospeccao-brasil/sgconfig.yml`. (FR-013, quickstart.md Scenario 12)
- [X] T043 [US4] Verify `.specify/feature.json` points to `specs/001-repo-tooling` as the active feature directory. (Spec Kit lifecycle)

**Checkpoint**: Agent readiness verified. User Story 4 is fully functional.

---

## Phase 7: User Story 5 - Design System Tooling (Priority: P2)

**Goal**: Tailwind config with the Prospecção Brasil brand tokens compiles successfully, and self-hosted JS libraries are present with no CDN references.

**Independent Test**: Run `make build-css`, verify `static/css/app.css` contains brand tokens. Verify `static/js/` has htmx, alpine, modal-trap. Grep for CDN URLs returns nothing.

### Implementation for User Story 5

- [X] T044 [US5] Verify `make build-css` produces `static/css/app.css` (already done in T026, but verify again after all changes). Confirm brand tokens present: Deep Navy (`#031636`/`#1a2b4c`), Sobrio Gold (`#765a1a`/`#b89650`), Montserrat, Inter. (FR-007, quickstart.md Scenario 9)
- [X] T045 [US5] Verify self-hosted JS: `ls static/js/htmx.min.js static/js/alpine.min.js static/js/modal-trap.js` all exist. Run `grep -r "cdn\|unpkg\|jsdelivr" static/ internal/ 2>/dev/null` and confirm no matches (no CDN references). (FR-008, quickstart.md Scenario 10)
- [X] T046 [US5] Verify `.env.example` (created in T008) has no secrets: `gitleaks detect --source . --no-git` exits 0. Verify `TOTP_ISSUER=Prospeccao Brasil` and DB name `prospeccaobrasil`. Verify no WHATSAPP/CNJ/LLM vars. (FR-015, quickstart.md Scenario 13)
- [X] T047 [US5] Verify `.gitignore` (created in T007) ignores `.env.local`, `node_modules`, `coverage.out`, `bin/`. (FR-016, quickstart.md Scenario 14)

**Checkpoint**: Design system tooling verified. User Story 5 is fully functional.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final verification that all 18 acceptance criteria pass and the spec is ready for `/speckit-analyze`.

- [X] T048 Run all 15 quickstart.md validation scenarios and verify each passes. Document any failures and fix them. (quickstart.md, all FRs)
- [X] T049 [P] Verify `go build ./...` completes without errors (FR-001). Run `go build ./...` and confirm exit 0.
- [X] T050 [P] Verify `make check` runs all gates (golangci-lint + go test + build-css + go build + ast-grep) and exits 0 (FR-002). Run `make check` and confirm exit 0.
- [X] T051 [P] Verify 85% coverage: run `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | grep total` and confirm the total is >= 85% (FR-002, Constitution principle IV).
- [X] T052 Run the Definition of Done verification table from spec.md -- execute each of the 18 verification commands and confirm all pass. (spec.md Definition of Done)
- [X] T053 Commit all SPEC-01 work and push. Verify CI is green on the final commit. (FR-014, FR-017)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- can start immediately. T001-T018 run in parallel (different files, no dependencies).
- **Foundational (Phase 2)**: Depends on Phase 1 (Makefile references `cmd/prospeccao` from T002; bootstrap references `.env.example` from T008 and `migrations/` from T003; CI references Makefile targets from T019). T019-T022 can run in parallel within Phase 2.
- **User Story 1 (Phase 3)**: Depends on Phase 2 (make check needs Makefile + all tooling). T023 (test) can run in parallel with T024-T026 (npm/go/tailwind setup), but T027 (make check) depends on all.
- **User Story 2 (Phase 4)**: Depends on Phase 2 (gates must be runnable). T030-T033 are verification tasks.
- **User Story 3 (Phase 5)**: Depends on Phase 2 (CI references Makefile). T034-T036 are CI config verification; T037 requires a git push (depends on T034-T036 + all files committed).
- **User Story 4 (Phase 6)**: Depends on Phase 1 (T015-T018 created extensions.yml, slim template, .devin configs). T038, T039 (AGENTS.md, constitution) can run in parallel with Phase 1 since they are independent files -- but are placed here for logical grouping. T040-T043 are verification.
- **User Story 5 (Phase 7)**: Depends on Phase 1 (T009-T014 created tailwind config, input.css, package.json, static/js). T044-T047 are verification.
- **Polish (Phase 8)**: Depends on all user stories complete. T048 is the final end-to-end validation.

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Foundational (Phase 2). No dependencies on other stories. This is the MVP -- deliver first.
- **User Story 2 (P1)**: Depends on Foundational (Phase 2). Independent of US1 (verifies gates, not setup).
- **User Story 3 (P1)**: Depends on Foundational (Phase 2). Independent of US1/US2 (verifies CI, not local gates). Requires all files committed for the PR.
- **User Story 4 (P2)**: Depends on Phase 1 only (T015-T018). Independent of US1-US3. T038/T039 (AGENTS.md, constitution) are independent files.
- **User Story 5 (P2)**: Depends on Phase 1 only (T009-T014). Independent of US1-US4.

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation (TDD where applicable)
- For US1: T023 (test) is written alongside T002 (implementation) -- both are in Phase 1/3
- Verification tasks run after the code they verify exists
- Story complete before moving to next priority

### Parallel Opportunities

- **Phase 1**: T001-T018 run in parallel (18 independent files, no dependencies)
- **Phase 2**: T019-T022 run in parallel (4 tooling files, depend on Phase 1 but not each other)
- **Phase 3-7**: User stories can be verified in parallel if team capacity allows (US1, US2, US3 are independent; US4, US5 are independent)
- **Phase 8**: T049, T050, T051 run in parallel (independent verification commands)

---

## Parallel Example: Phase 1 (Setup)

```bash
# Launch all setup tasks together (18 independent files, no dependencies):
Task: "Create go.mod"                                    # T001
Task: "Create cmd/prospeccao/main.go"                    # T002
Task: "Create migrations/.gitkeep"                       # T003
Task: "Create sqlc.yaml"                                 # T004
Task: "Create sgconfig.yml"                              # T005
Task: "Create .ast-grep/rules/ (7 rules)"                # T006
Task: "Create .gitignore"                                # T007
Task: "Create .env.example"                              # T008
Task: "Create tailwind.config.js"                        # T009
Task: "Create input.css"                                 # T010
Task: "Create package.json"                              # T011
Task: "Create static/js/htmx.min.js"                     # T012
Task: "Create static/js/alpine.min.js"                   # T013
Task: "Create static/js/modal-trap.js"                   # T014
Task: "Create .specify/extensions.yml"                   # T015
Task: "Create .specify/templates/spec-template-slim.md"  # T016
Task: "Create .devin/config.json"                        # T017
Task: "Create .devin/mcp_config.json"                    # T018
```

## Parallel Example: Phase 8 (Polish)

```bash
# Launch all verification tasks together (independent commands):
Task: "Verify go build ./... exits 0"                    # T049
Task: "Verify make check exits 0"                        # T050
Task: "Verify coverage >= 85%"                           # T051
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T018 in parallel)
2. Complete Phase 2: Foundational (T019-T022 in parallel)
3. Complete Phase 3: User Story 1 (T023-T029 -- make `make check` + `make setup` pass)
4. **STOP and VALIDATE**: Run `make setup` on a clean clone; verify all gates pass
5. The dev environment is now reproducible -- other specs can begin

### Incremental Delivery

1. Setup + Foundational -> all tooling files created, `make check` runnable
2. Add User Story 1 -> `make setup` works end-to-end, `make check` passes -> MVP delivered
3. Add User Story 2 -> quality gates verified to catch violations
4. Add User Story 3 -> CI green on trivial commit
5. Add User Story 4 -> agent readiness verified (AGENTS.md, constitution, extensions, .devin)
6. Add User Story 5 -> design system tooling verified (Tailwind tokens, self-hosted JS)
7. Polish -> all 18 acceptance criteria verified, ready for `/speckit-analyze`

### Suggested MVP Scope

**User Story 1 only** (Phases 1-3). After US1, the dev environment is
reproducible and `make check` passes. US2-US5 are verification of tooling
that already exists from Phase 1; they can be done in a single polish pass.
The critical path is T001-T018 (create files) -> T019-T022 (tooling) ->
T027 (make check pass) -> T028 (make setup pass).
