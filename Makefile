.PHONY: setup dev check lint test build-css build migrate migrate-down sqlc fmt ast-grep run seed

# Load .env.local if it exists (local dev). CI sets env vars directly.
-include .env.local
export

setup:
	bash scripts/bootstrap.sh

dev:
	go run ./cmd/prospeccao

check: lint test build-css build ast-grep

lint:
	golangci-lint run

test:
	go test -race -p 1 -timeout 20m -coverprofile=coverage.out -covermode=atomic ./...
	@# internal/db is sqlc-generated (DO NOT EDIT) -- exclude from coverage gate.
	@# cmd/prospeccao is the server entry point (main) -- not unit-testable.
	@# scripts/ is a standalone seed CLI -- not unit-testable.
	@grep -v "internal/db/" coverage.out | grep -v "cmd/prospeccao/" | grep -v "scripts/" > coverage-nogen.out || true
	@echo "mode: atomic" > coverage-nogen.out.tmp
	@grep -v "^mode:" coverage-nogen.out >> coverage-nogen.out.tmp 2>/dev/null || true
	@mv coverage-nogen.out.tmp coverage-nogen.out
	@echo "Coverage (excluding internal/db, cmd/prospeccao, scripts):"
	@go tool cover -func=coverage-nogen.out | grep total || echo "No non-excluded packages yet (expected in SPEC-01)"
	@COV=$$(go tool cover -func=coverage-nogen.out | grep total | awk '{print $$3}' | tr -d '%') && \
	awk "BEGIN { exit !($$COV >= 70) }" || (echo "Coverage below 70%" && exit 1)
	@echo "Coverage gate: 70% minimum -- PASS"

build-css:
	npx tailwindcss -i input.css -o static/css/app.css --minify

build:
	go build -o bin/prospeccao ./cmd/prospeccao

migrate:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

sqlc:
	sqlc generate

fmt:
	gofmt -l -w .
	goimports -l -w .

ast-grep:
	ast-grep scan

run: build
	./bin/prospeccao

seed:
	go run ./scripts/seed.go
