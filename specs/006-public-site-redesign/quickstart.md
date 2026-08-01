# Quickstart: Public Site Redesign

**Date**: 2026-07-31
**Spec**: [spec.md](spec.md)

---

## Prerequisites

- Go 1.26+
- Postgres 16+ (local, database `prospeccaobrasil_test`)
- Node/npm 20+ (for Tailwind CSS build)
- `make setup` run previously (all tools installed)

---

## Setup

```bash
# 1. Create test database (if not exists)
createdb prospeccaobrasil_test

# 2. Run migrations
export DATABASE_URL="postgres://localhost:5432/prospeccaobrasil_test?sslmode=disable"
migrate -path migrations -database "$DATABASE_URL" up

# 3. Regenerate sqlc (after adding company column to contacts query)
make sqlc

# 4. Build CSS
make build-css

# 5. Build binary
make build
```

---

## Validation Scenarios

### Scenario 1: Home premium (P1)

1. Run `make dev`
2. Open `http://localhost:8080/`
3. Verify:
   - Hero full-bleed with background image (or Deep Navy fallback)
   - Headline: "Encontramos o ponto comercial ideal" (or similar market copy)
   - CTA "Solicite uma apresentacao" visible above the fold
   - Metrics strip: 4 metrics (Pontos Comercializados, Clientes, Cidades, Anos)
   - Services section: 3+ cards linking to /servicos/{slug}
   - Testimonials: 2+ with author name
   - Final CTA "Solicite uma apresentacao" or "Fale Conosco"
4. Resize to 375px (mobile): no horizontal scroll
5. grep for forbidden words: `grep -ri "carga cognitiva\|pipeline\|plataforma\|software" internal/template/home.html` -- must return empty

### Scenario 2: Servicos index (P2)

1. Open `http://localhost:8080/servicos`
2. Verify:
   - Index with 5+ services: Expansao de Redes, Built to Suit, Strip Mall, Lajes Comerciais, Prospecacao de Ponto
   - Each card has title, short description, link to /servicos/{slug}
   - No generic icons for "Relatorios PDF" or "Gestao de Pipeline" (removed)
3. grep: `grep -ri "relatorios pdf\|gestao de pipeline" internal/template/servicos.html` -- must return empty

### Scenario 3: Servico detalhe (P2)

1. Open `http://localhost:8080/servicos/expansao-de-redes`
2. Verify:
   - Page title "Expansao de Redes"
   - Methodology section with steps (Plano Diretor, Macrolocalizacao, Microlocalizacao, Prospecacao)
   - CTA "Fale com um especialista" linking to /fale-conosco
3. Open `http://localhost:8080/servicos/built-to-suit` -- verify similar structure
4. Open `http://localhost:8080/servicos/servico-inexistente` -- verify 404 page

### Scenario 4: Quem somos (P3)

1. Open `http://localhost:8080/quem-somos`
2. Verify:
   - Founder story: "Luiz Claudio", "15 anos Shell Brasil", "redes de franquias e varejo"
   - Mission, Vision, Values blocks (Transparencia, Profissionalismo, Etica, Comprometimento, Agilidade)
   - CRECI mention
3. grep: `grep -ri "carga cognitiva\|plataforma\|software" internal/template/quem-somos.html` -- must return empty

### Scenario 5: Nossos clientes (P4)

1. Open `http://localhost:8080/nossos-clientes`
2. Verify:
   - 2+ testimonials (Larissa Mello, Roberto Andrade, and/or Joao Viana)
   - Metrics strip
   - No "Em breve" or empty state text
   - CTA at bottom

### Scenario 6: Fale Conosco (P5)

1. Open `http://localhost:8080/fale-conosco`
2. Verify:
   - Form with fields: Empresa, Nome, Email, Telefone, Assunto, Mensagem
   - Contact info (address Botafogo RJ, phones, email, WhatsApp) visible on page
   - Submit with valid data: success message (HTMX, no reload)
   - Submit with invalid email: inline error
   - Disable JavaScript, submit: page reloads with success/error

### Scenario 7: Login CSS fix (P6)

1. Open `http://localhost:8080/login` (dev router serves auth)
2. Verify:
   - `<link rel="stylesheet" href="/static/css/app.css">` in page source
   - Form uses `form-input`, `btn`, `btn-primary` classes
   - No `style="color:red"` inline styles
   - Error messages use `alert alert-error` class
3. Open `/2fa/setup` and `/2fa/verify` -- verify same CSS treatment
4. Login with valid credentials: 2FA flow works identically (redirect, cookie)

### Scenario 8: Mobile responsivo (all pages)

1. Open each page at 375px width (DevTools or resize)
2. Verify: no horizontal scroll, hero stacks vertically, nav collapses to hamburger, footer stacks

### Scenario 9: Newsletter (unchanged)

1. Open `http://localhost:8080/`
2. Scroll to footer, enter email in newsletter form
3. Submit: success message appears (HTMX)
4. Submit same email: error message (duplicate)

### Scenario 10: 404 page

1. Open `http://localhost:8080/pagina-inexistente`
2. Verify: institutional 404 page with nav and footer

### Scenario 11: Forbidden copy check (all public pages)

```bash
grep -ri "carga cognitiva\|pipeline\|plataforma\|software\|relatorios pdf\|gestao de pipeline" \
    internal/template/home.html \
    internal/template/servicos.html \
    internal/template/servico-detalhe.html \
    internal/template/quem-somos.html \
    internal/template/nossos-clientes.html \
    internal/template/fale-conosco.html \
    internal/template/partials/nav.html \
    internal/template/partials/footer.html
```
Must return empty (no matches).

### Scenario 12: make check

```bash
export DATABASE_URL="postgres://localhost:5432/prospeccaobrasil_test?sslmode=disable"
export ENCRYPTION_KEY="test-encryption-key-32bytes!!"
make check
```
Verify: golangci-lint 0 issues, all tests pass, coverage >= 70% (app) / 85% (auth), build succeeds, ast-grep clean.

### Scenario 13: CI green

```bash
git push
gh run watch
```
Verify: CI passes green (lint, test, coverage gate, build, ast-grep, govulncheck).

### Scenario 14: Pencil designs exist

```bash
ls designs/prospeccao.pen
```
Verify: file exists with frames for Home (desktop+mobile), Servicos (index+detalhe), Quem Somos, Nossos Clientes, Fale Conosco, Login.

### Scenario 15: Auth functionality unchanged

1. Login with admin credentials
2. Complete 2FA TOTP
3. Access /admin (dashboard)
4. Logout
5. Verify: session cookie cleared, redirect to /login
All behavior identical to SPEC-03/SPEC-05.

### Scenario 16: Contact form persistence

1. Submit Fale Conosco with valid data
2. Query database: `SELECT * FROM contact_submissions ORDER BY created_at DESC LIMIT 1;`
3. Verify: row exists with company field (if filled), name, email, subject, message
