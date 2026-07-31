# Prospecção Brasil -- Project Constitution

**Version**: 1.0.0
**Ratified**: 2026-07-31
**Status**: Active

This constitution defines the non-negotiable principles governing all
development on the Prospecção Brasil platform. Every spec, plan, and
implementation MUST comply with these principles. Conflicts are resolved by
adjusting the work, not by diluting the principle.

---

### I. Spec-Driven Development

All features follow the Spec Kit lifecycle: specify -> plan -> tasks ->
analyze -> implement. No implementation begins before the spec and plan are
complete. Infrastructure specs use the slim template; product specs use the
full template. One spec at a time.

### II. Security-First Design (LGPD)

Client and property data is PII under LGPD. Every data entity that stores PII
must have a threat model (optional hook for pure tooling specs, mandatory for
data/auth/integration specs). No secrets in the repo (gitleaks enforces).
`.env.local` is gitignored. Session cookies are HttpOnly + SameSite=Strict +
Secure. 2FA TOTP is required for the admin user. The `tenant_id` isolation
rule ships from SPEC-01 (ast-grep) and fires once SPEC-03 adds the column.

### III. Single-Binary & Tooling Consistency

Prospecção Brasil is a single Go binary. `make check` orchestrates all quality
gates (lint, test, build, ast-grep). CI mirrors the Makefile exactly -- same
flags, same env vars, same exclusions. No divergence between local and CI is
acceptable. Tailwind CSS is a build-time tool (not a runtime dependency).

### IV. Test-First & Continuous Quality

85% test coverage minimum (enforced by CI gate, excluding sqlc-generated
`internal/db` and the `cmd/prospeccao` entry point). Tests run with `-race`
and `-p 1` (sequential) to catch data races and integration test isolation
issues. `govulncheck` runs in CI. ast-grep structural rules catch violations
that linters miss (bare errors, missing tenant filters, missing auth, missing
context, hardcoded secrets, bare buttons, fmt.Println in non-main code).

### V. Observability & Structured Logging

All logging via `log/slog` (JSON handler in production, text in development).
No `fmt.Println` in non-main code (ast-grep rule `go-slog-fmt.yml` enforces).
The `/healthz` endpoint is public (no auth) for liveness probes. Error
messages are wrapped with `fmt.Errorf("...: %w", err)` for traceability -- no
bare `return err` (ast-grep rule `go-bare-error.yml` enforces).

### VI. Forward-Only Migrations

SQL migrations are forward-only via `golang-migrate`. No `migrate down` in
production. Each migration is append-only and reversible only via a new
"undo" migration. The `migrations/` directory is tracked in git. SPEC-01
creates the directory with `.gitkeep` only; SPEC-02 adds the schema.

### VII. Simplicity for Single-User

The MVP is single-tenant, single-admin (Luiz Claudio). The backend is
future-proofed (tenant_id, RBAC, 2FA, sessions) but the UI must not be complex
for a single user. No premature multi-user UI, no role management screens, no
tenant switcher. YAGNI for MVP: ship the encanamento (rules, middleware, schema)
but defer the multi-user UI until there is a second user. The prospector's
cognitive load is the primary UX metric.

---

## Quality Gates

| Gate | Tool | When | Enforced By |
|------|------|------|-------------|
| Format | `gofmt` | pre-commit | `.pre-commit-config.yaml` |
| Lint | `golangci-lint` | pre-commit + CI | `.pre-commit-config.yaml`, `ci.yml` |
| Vet | `go vet` | CI | `ci.yml` |
| Test + coverage (85%) | `go test -race -p 1` | pre-commit + CI | Makefile, `ci.yml` |
| Build | `go build` | CI | `ci.yml` |
| Structural scan | `ast-grep scan` | pre-commit + CI | `.pre-commit-config.yaml`, `ci.yml` |
| Secret scan | `gitleaks` | pre-commit | `.pre-commit-config.yaml` |
| Dependency vuln | `govulncheck` | CI | `ci.yml` |
| Spec Kit review | `/speckit-analyze` | after tasks | Spec Kit lifecycle |
| Security audit | `tekimax-security` (optional) | after implement | `.specify/extensions.yml` |
| Security gate | `gate-check` (mandatory) | before implement | `.specify/extensions.yml` |
