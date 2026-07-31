#!/usr/bin/env bash
# Prospecção Brasil one-command dev setup.
# Checks prerequisites, creates .env.local, creates the dev database, runs
# migrations, generates sqlc, and runs `make check` to verify.
set -euo pipefail

echo "=== Prospecção Brasil bootstrap ==="

# --- 1. Check prerequisites ---
echo "[1/7] Checking prerequisites..."

check_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "  MISSING: $1 -- install it before continuing."
    return 1
  fi
  echo "  OK: $1"
}

MISSING=0
check_cmd go || MISSING=1
check_cmd psql || MISSING=1
check_cmd golangci-lint || MISSING=1
check_cmd sqlc || MISSING=1
check_cmd migrate || MISSING=1
check_cmd ast-grep || MISSING=1
check_cmd pre-commit || MISSING=1
check_cmd gitleaks || MISSING=1
check_cmd gh || MISSING=1
check_cmd node || MISSING=1
check_cmd npm || MISSING=1

if [ "$MISSING" -ne 0 ]; then
  echo ""
  echo "Some prerequisites are missing. Install them and re-run:"
  echo "  go:              https://go.dev/dl/ (need 1.26+)"
  echo "  postgres:        brew install postgresql@16"
  echo "  golangci-lint:   brew install golangci-lint"
  echo "  sqlc:            brew install sqlc"
  echo "  migrate:         go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
  echo "  ast-grep:        brew install ast-grep"
  echo "  pre-commit:      brew install pre-commit  (or: pip install pre-commit)"
  echo "  gitleaks:        brew install gitleaks"
  echo "  gh:              brew install gh  (then: gh auth login)"
  echo "  node:            brew install node  (Node 20+ for Tailwind CSS build)"
  echo "  npm:             included with node"
  exit 1
fi

# Go version check (need 1.26+ for stdlib CVE fixes required by govulncheck)
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
GO_MINOR=$(echo "$GO_VERSION" | cut -d. -f2)
if [ "$GO_MAJOR" -lt 1 ] || { [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 26 ]; }; then
  echo "  Go version $GO_VERSION is below 1.26. Upgrade Go (needed for stdlib CVE fixes)."
  exit 1
fi
echo "  Go version $GO_VERSION OK"

# Postgres version check (need 16+)
PG_VERSION=$(psql --version 2>/dev/null | awk '{print $3}' | cut -d. -f1)
if [ -z "$PG_VERSION" ]; then
  echo "  WARNING: could not determine Postgres version (psql --version failed)"
elif [ "$PG_VERSION" -lt 16 ]; then
  echo "  Postgres version $PG_VERSION is below 16. Upgrade Postgres."
  exit 1
else
  echo "  Postgres version $PG_VERSION OK"
fi

# --- 2. Create .env.local from .env.example ---
echo "[2/7] Creating .env.local..."
if [ ! -f .env.local ]; then
  if [ -f .env.example ]; then
    cp .env.example .env.local
    echo "  Created .env.local from .env.example (edit with real values)"
  else
    echo "  WARNING: .env.example not found, skipping .env.local"
  fi
else
  echo "  .env.local already exists, skipping"
fi

# --- 3. Create the dev database ---
echo "[3/7] Creating dev database..."
DB_NAME="prospeccaobrasil"
if command -v psql >/dev/null 2>&1; then
  if psql -lqt 2>/dev/null | cut -d'|' -f1 | grep -qw "$DB_NAME"; then
    echo "  Database '$DB_NAME' already exists, skipping"
  else
    createdb "$DB_NAME" 2>/dev/null && echo "  Created database '$DB_NAME'" || \
      echo "  WARNING: could not create database '$DB_NAME' (create it manually)"
  fi
else
  echo "  WARNING: psql not available, skipping database creation"
fi

# --- 4. Run migrations ---
echo "[4/7] Running migrations..."
if [ -d migrations ] && [ -n "$(ls -A migrations 2>/dev/null)" ]; then
  DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/$DB_NAME?sslmode=disable}"
  if migrate -path migrations -database "$DATABASE_URL" up; then
    echo "  Migrations applied"
  else
    echo "  WARNING: migrate up failed (is Postgres running?)"
  fi
else
  echo "  No migrations directory or empty, skipping"
fi

# --- 5. Run sqlc generate ---
echo "[5/7] Generating sqlc..."
if command -v sqlc >/dev/null 2>&1; then
  if [ -n "$(ls -A internal/db/queries/*.sql 2>/dev/null)" ]; then
    if sqlc generate; then
      echo "  sqlc generate done"
    else
      echo "  WARNING: sqlc generate failed"
    fi
  else
    echo "  No sqlc queries yet, skipping"
  fi
else
  echo "  sqlc not installed, skipping"
fi

# --- 6. Install pre-commit hooks ---
echo "[6/7] Installing pre-commit hooks..."
if pre-commit install >/dev/null 2>&1; then
  echo "  pre-commit installed at .git/hooks/pre-commit"
else
  echo "  WARNING: pre-commit install failed (is this a git repo?)"
fi

# --- 7. Run make check ---
echo "[7/7] Running make check..."
if make check 2>/dev/null; then
  echo "  make check passed"
else
  echo "  WARNING: make check failed (expected on a fresh repo with no code yet)"
fi

echo ""
echo "=== Prospecção Brasil bootstrap complete ==="
echo "Next: edit .env.local with real values, then 'make dev'"
