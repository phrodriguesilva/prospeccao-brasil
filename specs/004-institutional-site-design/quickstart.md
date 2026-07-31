# Quickstart: Institutional Site & Design System

**Date**: 2026-07-31
**Spec**: [spec.md](spec.md)
**Plan**: [plan.md](plan.md)
**Contracts**: [endpoints.md](contracts/endpoints.md)

## Prerequisites

- Go 1.26+
- PostgreSQL 16+ (running on localhost:5432)
- Node.js 20+ (for Tailwind CSS build)
- `make check` passes (SPEC-01/02/03 complete)
- `DATABASE_URL` set in `.env.local` (pointing to test DB)
- `ENCRYPTION_KEY` set in `.env.local`

## Setup

```bash
# 1. Run migrations (creates contact_submissions + newsletter_subscribers)
make migrate

# 2. Build CSS (compiles Tailwind with design system component classes)
make build-css

# 3. Start the server
make dev
```

## Validation Scenarios

### Scenario 1: Home page renders

```bash
curl -s http://localhost:8080/ | grep -o "Prospecção Brasil"
# Expected: "Prospecção Brasil" appears in the hero section
```

### Scenario 2: Navigation works

```bash
# All 5 institutional pages return 200
for path in / /quem-somos /servicos /nossos-clientes /fale-conosco; do
  echo -n "$path: "
  curl -s -o /dev/null -w "%{http_code}" http://localhost:8080$path
  echo ""
done
# Expected: all return 200
```

### Scenario 3: 404 page renders with institutional layout

```bash
curl -s http://localhost:8080/nonexistent | grep -o "Página não encontrada\|404"
# Expected: 404 page with nav and footer
```

### Scenario 4: Contact form submission (valid)

```bash
curl -s -X POST http://localhost:8080/fale-conosco \
  -d "name=João Silva&email=joao@example.com&phone=&subject=Teste&message=Gostaria de saber mais sobre imóveis comerciais" \
  | grep -o "sucesso\|enviada"
# Expected: success message appears
```

### Scenario 5: Contact form validation (invalid email)

```bash
curl -s -X POST http://localhost:8080/fale-conosco \
  -d "name=João&email=invalid&subject=Teste&message=Mensagem muito curta" \
  | grep -o "inválido\|erro"
# Expected: validation error message
```

### Scenario 6: Newsletter subscription (new)

```bash
curl -s -X POST http://localhost:8080/newsletter \
  -d "email=new@example.com" \
  | grep -o "confirmada\|inscrito"
# Expected: success message
```

### Scenario 7: Newsletter idempotency (same email twice)

```bash
# First subscription
curl -s -X POST http://localhost:8080/newsletter -d "email=duplicate@example.com" | grep -o "confirmada"
# Second subscription (same email)
curl -s -X POST http://localhost:8080/newsletter -d "email=duplicate@example.com" | grep -o "já está inscrito"
# Expected: first says "confirmada", second says "já está inscrito"
```

### Scenario 8: Newsletter validation (invalid email)

```bash
curl -s -X POST http://localhost:8080/newsletter -d "email=not-an-email" | grep -o "inválido"
# Expected: validation error
```

### Scenario 9: Design system component classes exist in compiled CSS

```bash
grep -o "\.btn\|\.badge\|\.card\|\.form-input" static/css/app.css | sort -u
# Expected: .btn, .badge, .card, .form-input classes are present
```

### Scenario 10: Self-hosted JS files exist

```bash
ls -la static/js/htmx.min.js static/js/alpine.min.js
# Expected: both files exist, non-empty
```

### Scenario 11: Static files served correctly

```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/static/js/htmx.min.js
# Expected: 200
```

### Scenario 12: Nav highlights active page

```bash
curl -s http://localhost:8080/servicos | grep -o "active.*Serviços\|Serviços.*active"
# Expected: the "Serviços" nav item has an "active" class
```

### Scenario 13: Mobile responsive (viewport meta tag)

```bash
curl -s http://localhost:8080/ | grep -o 'viewport'
# Expected: <meta name="viewport" ...> tag present
```

### Scenario 14: Rate limiting on contact form

```bash
# 5 rapid submissions (should succeed), 6th should be 429
for i in $(seq 1 6); do
  echo -n "Attempt $i: "
  curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/fale-conosco \
    -d "name=Test&email=test@example.com&subject=Test&message=This is a test message for rate limiting" \
    -H "X-Forwarded-For: 10.0.0.1"
  echo ""
done
# Expected: first 5 return 200, 6th returns 429
```

### Scenario 15: make check passes

```bash
make check
# Expected: golangci-lint 0 issues, all tests pass, coverage >= 85%, build succeeds, ast-grep 0 errors
```

### Scenario 16: Database tables exist

```bash
psql -U $USER -d prospeccaobrasil_test -c "\dt contact_submissions"
psql -U $USER -d prospeccaobrasil_test -c "\dt newsletter_subscribers"
# Expected: both tables exist with correct columns
```
