# Quickstart: SPEC-01 -- Repo Tooling & Dev Environment

**Date**: 2026-07-31
**Spec**: [spec.md](./spec.md)

## Prerequisites

Install these on your machine before running `make setup`:

| Tool | macOS install | Linux install |
|------|---------------|---------------|
| Go 1.26+ | `brew install go` or [go.dev/dl](https://go.dev/dl/) | [go.dev/dl](https://go.dev/dl/) |
| Postgres 16+ | `brew install postgresql@16 && brew services start postgresql@16` | `sudo apt install postgresql-16 && sudo systemctl start postgresql` |
| golangci-lint | `brew install golangci-lint` | [install script](https://golangci-lint.run/usage/install/) |
| sqlc | `brew install sqlc` | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| migrate | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` | same |
| ast-grep | `brew install ast-grep` | [ast-grep install](https://ast-grep.github.io/guide/installation.html) |
| pre-commit | `brew install pre-commit` | `pip install pre-commit` |
| gitleaks | `brew install gitleaks` | [gitleaks install](https://github.com/gitleaks/gitleaks/releases) |
| gh | `brew install gh` then `gh auth login` | [gh install](https://github.com/cli/cli#installation) |
| Node/npm 20+ | `brew install node` | [nodejs.org](https://nodejs.org/) |
| uv | `brew install uv` | [uv install](https://docs.astral.sh/uv/getting-started/installation/) |
| goimports | `go install golang.org/x/tools/cmd/goimports@latest` | same |
| govulncheck | `go install golang.org/x/vuln/cmd/govulncheck@latest` | same |

## Validation Scenarios

### Scenario 1: One-command dev setup (FR-003)

```bash
cd /Users/relterborges/Documents/Dev/prospeccao-brasil
make setup
```

**Expected output** (abridged):
```
=== Prospecção Brasil bootstrap ===
[1/7] Checking prerequisites...
  OK: go
  OK: psql
  OK: golangci-lint
  OK: sqlc
  OK: migrate
  OK: ast-grep
  OK: pre-commit
  OK: gitleaks
  OK: gh
  OK: node
  OK: npm
  Go version 1.26 OK
  Postgres version 16 OK
[2/7] Creating .env.local...
  Created .env.local from .env.example (edit with real values)
[3/7] Creating dev database...
  Created database 'prospeccaobrasil'   # or "already exists, skipping"
[4/7] Running migrations...
  No migrations directory or empty, skipping   # or "Migrations applied"
[5/7] Generating sqlc...
  sqlc generate done   # no-op on empty queries
[6/7] Installing pre-commit hooks...
  pre-commit installed at .git/hooks/pre-commit
[7/7] Running make check...
  make check passed
=== Prospecção Brasil bootstrap complete ===
```

**Pass criteria**: script exits 0; `.env.local` exists; `prospeccaobrasil`
database exists; `make check` passed.

### Scenario 2: Quality gates pass (FR-002)

```bash
make check
```

**Expected**: runs golangci-lint, go test (with coverage), build-css, go
build, and ast-grep scan. All pass. Exit 0.

**Verify coverage**:
```bash
go tool cover -func=coverage.out | grep total
# Expected: total: (statements)    100.0%   # or >= 85%
```

### Scenario 3: Build and run the dev server (FR-001)

```bash
make build
./bin/prospeccao
# or: make dev
```

**Expected**: server starts on `:8080` (or `$PORT`). In another terminal:
```bash
curl http://localhost:8080/healthz
# Expected: {"status":"ok"}
```

**Pass criteria**: `200 OK` with `{"status":"ok"}` body.

### Scenario 4: CI is green on a trivial commit (FR-014)

```bash
git checkout -b test-ci
echo "<!-- trivial edit -->" >> README.md
git add README.md
git commit -m "trivial edit to verify CI"
git push -u origin test-ci
gh pr create --title "test CI" --body "trivial edit"
gh pr checks --watch   # verify all CI checks pass
gh pr merge --delete-branch  # or close and delete locally
git checkout main
git branch -D test-ci   # cleanup after verification
```

**Pass criteria**: all CI jobs (build, test, lint, ast-grep, govulncheck)
pass on the PR.

### Scenario 5: Pre-commit blocks violations (FR-004)

```bash
# Test 1: hardcoded secret
echo 'password := "hunter2"' >> cmd/prospeccao/main.go
git add cmd/prospeccao/main.go
git commit -m "test secret detection"
# Expected: gitleaks blocks the commit

# Revert the change
git checkout cmd/prospeccao/main.go

# Test 2: bare error
cat >> cmd/prospeccao/main.go <<'EOF'
func badFunc() error {
    err := doSomething()
    return err  // bare return err -- should be wrapped
}
EOF
git add cmd/prospeccao/main.go
git commit -m "test bare error detection"
# Expected: ast-grep go-bare-error.yml or golangci-lint blocks the commit

git checkout cmd/prospeccao/main.go
```

**Pass criteria**: each violation is blocked with a clear message
identifying the file and line.

### Scenario 6: ast-grep rules scan cleanly (FR-005)

```bash
make ast-grep
# or: ast-grep scan
```

**Expected**: no matches (no violations), exit 0.

**Verify all 7 rules exist**:
```bash
ls .ast-grep/rules/
# Expected:
# go-bare-error.yml
# go-handler-missing-auth.yml
# go-hardcoded-secret.yml
# go-missing-context.yml
# go-missing-tenant-filter.yml
# go-slog-fmt.yml
# tmpl-bare-button.yml
```

### Scenario 7: AGENTS.md and constitution present (FR-009, FR-010)

```bash
test -f AGENTS.md && echo "AGENTS.md OK"
test -f .specify/memory/constitution.md && echo "constitution OK"
grep -q "constitution" AGENTS.md && echo "AGENTS.md references constitution"
grep -c "^### " .specify/memory/constitution.md
# Expected: 7 (the 7 principles)
```

**Expected**: all checks pass; constitution has 7 principles.

### Scenario 8: sgconfig.yml and sqlc.yaml present (FR-005, FR-006)

```bash
cat sgconfig.yml
# Expected:
# ruleDirs:
#   - .ast-grep/rules
# fileTypes:
#   - go: ["go"]
#   - html: ["html", "tmpl", "templ"]

cat sqlc.yaml
# Expected: postgresql, pgx/v5, internal/db/queries, migrations/

sqlc generate
# Expected: exits 0 (no-op on empty queries)
```

### Scenario 9: Tailwind builds with brand tokens (FR-007)

```bash
make build-css
test -f static/css/app.css && echo "app.css built OK"
grep -q "031636\|1a2b4c" static/css/app.css && echo "Deep Navy tokens present"
grep -q "765a1a\|b89650" static/css/app.css && echo "Sobrio Gold tokens present"
grep -q "Montserrat" static/css/app.css && echo "Montserrat font present"
grep -q "Inter" static/css/app.css && echo "Inter font present"
```

**Pass criteria**: `app.css` exists and contains the brand color tokens and
font families.

### Scenario 10: Self-hosted JS, no CDN (FR-008)

```bash
ls static/js/htmx.min.js static/js/alpine.min.js static/js/modal-trap.js
# Expected: all three files exist

grep -r "cdn\|unpkg\|jsdelivr" static/ internal/ 2>/dev/null
# Expected: no matches (no CDN references anywhere)
```

**Pass criteria**: all three JS files present; no CDN URLs in static/ or
internal/.

### Scenario 11: extensions.yml and slim template present (FR-011, FR-012)

```bash
python3 -c "import yaml; yaml.safe_load(open('.specify/extensions.yml'))" && echo "extensions.yml valid"
test -f .specify/templates/spec-template-slim.md && echo "slim template OK"
```

### Scenario 12: .devin configs present and valid (FR-013)

```bash
python3 -c "import json; json.load(open('.devin/config.json')); json.load(open('.devin/mcp_config.json'))" && echo "devin configs valid"
grep -q "prospeccao-brasil/sgconfig.yml" .devin/config.json && echo "AST_GREP_CONFIG path correct"
```

### Scenario 13: .env.example present, no secrets (FR-015)

```bash
test -f .env.example && echo ".env.example OK"
gitleaks detect --source . --no-git
# Expected: exits 0 (no secrets detected)
grep -q "TOTP_ISSUER=Prospeccao Brasil" .env.example && echo "issuer correct"
grep -q "prospeccaobrasil" .env.example && echo "DB name correct"
! grep -q "WHATSAPP\|CNJ_API\|LLM_" .env.example && echo "AI/WhatsApp/CNJ vars removed"
```

### Scenario 14: .gitignore present (FR-016)

```bash
test -f .gitignore && echo ".gitignore OK"
grep -q ".env.local" .gitignore && echo "env.local ignored"
grep -q "node_modules" .gitignore && echo "node_modules ignored"
grep -q "coverage.out" .gitignore && echo "coverage ignored"
grep -q "bin/" .gitignore && echo "bin ignored"
```

### Scenario 15: Repo is a git repo with commits (FR-017)

```bash
git rev-parse --is-inside-work-tree && echo "git repo OK"
git log --oneline -1 && echo "at least one commit OK"
git remote -v | grep -q "prospeccao-brasil" && echo "remote configured"
```

## Definition of Done Verification

Run all 15 scenarios above. SPEC-01 is done when all pass. See the
[Definition of Done table in spec.md](./spec.md#definition-of-done) for
the per-criterion verification commands (18 rows mapping each FR to a
command).
