# Endpoint Contracts: Public Site Redesign

**Date**: 2026-07-31
**Spec**: [spec.md](spec.md)

Todas as paginas publicas sao server-rendered HTML (Go `html/template`).
Nao ha auth, nao ha session cookie. Formularios usam HTMX com fallback
no-JS.

---

## Public Pages (GET, returns full HTML page)

| Method | Path | Handler | Template | Description |
|--------|------|---------|----------|-------------|
| GET | `/` | `InstitutionalHandler.Home` | `home.html` | Home com hero full-bleed, metricas, servicos, depoimentos, CTA |
| GET | `/quem-somos` | `InstitutionalHandler.QuemSomos` | `quem-somos.html` | Historia do fundador, missao/visao/valores, CRECI |
| GET | `/servicos` | `InstitutionalHandler.Servicos` | `servicos.html` | Indice de servicos com cards descritivos |
| GET | `/servicos/{slug}` | `InstitutionalHandler.ServicoDetalhe` | `servico-detalhe.html` | Pagina detalhada de um servico (metodologia, etapas, CTA) |
| GET | `/nossos-clientes` | `InstitutionalHandler.NossosClientes` | `nossos-clientes.html` | Depoimentos, metricas, CTA |
| GET | `/fale-conosco` | `InstitutionalHandler.FaleConosco` | `fale-conosco.html` | Formulario + info de contato em destaque |

### Nova rota

- `GET /servicos/{slug}` -- nova no SPEC-06. `{slug}` e o identificador
  do servico (ex: `expansao-de-redes`). Se o slug nao existir no map
  estatico, retorna 404 com a pagina de erro institucional.

---

## 404 (GET, any unmatched path)

| Method | Path | Handler | Template | Description |
|--------|------|---------|----------|-------------|
| GET | `*` (unmatched) | `InstitutionalHandler.NotFound` | `404.html` | Custom 404 com layout institucional |

Inclui `/servicos/{slug-inexistente}`.

---

## Contact Form (POST, returns HTML fragment or redirect)

| Method | Path | Handler | Request | Response (HTMX) | Response (no-JS) |
|--------|------|---------|---------|-----------------|------------------|
| POST | `/fale-conosco` | `ContactHandler.Submit` | `application/x-www-form-urlencoded`: name, email, phone, **company** (opcional, novo), subject, message | 200 with `fragments/contact_success.html` or `fragments/contact_error.html` | 302 redirect to `/fale-conosco?success=1` or `/fale-conosco?error=...` |

### Campos do formulario

| Campo | name attr | Type | Required | Validation |
|-------|-----------|------|----------|------------|
| Empresa | `company` | text | nao | max 255 chars |
| Nome | `name` | text | sim | min 2, max 255 chars |
| Email | `email` | email | sim | formato email valido |
| Telefone | `phone` | tel | nao | max 20 chars |
| Assunto | `subject` | text | sim | min 2, max 255 chars |
| Mensagem | `message` | textarea | sim | min 10, max 5000 chars |

### Novo campo

- `company` (Empresa): opcional, VARCHAR(255), NULL permitido. Novo no
  SPEC-06. Requer migration `000003_add_company_to_contact_submissions`
  e regeneracao do sqlc.

### Validation errors

Retornados como `contact_error.html` fragment (HTMX) ou
`/fale-conosco?error=validation` redirect (no-JS):

- Missing or too-short name (min 2 chars)
- Invalid email format
- Missing or too-short subject (min 2 chars)
- Missing or too-short message (min 10 chars)
- Rate limit exceeded (429 status code)

### Success

Retornado como `contact_success.html` fragment (HTMX) ou
`/fale-conosco?success=1` redirect (no-JS).

---

## Newsletter (POST, returns HTML fragment or redirect)

| Method | Path | Handler | Request | Response (HTMX) | Response (no-JS) |
|--------|------|---------|---------|-----------------|------------------|
| POST | `/newsletter` | `NewsletterHandler.Subscribe` | `application/x-www-form-urlencoded`: email | 200 with `fragments/newsletter_success.html` or `fragments/newsletter_error.html` | 302 redirect |

**Inalterado do SPEC-04.** Nenhuma mudanca no SPEC-06.

---

## Auth Pages (GET, internal subdomain only)

Estas paginas sao servidas apenas em `sistema.prospeccaobrasil.com`
(buildInternalRouter), nao no dominio publico.

| Method | Path | Handler | Template | Description |
|--------|------|---------|----------|-------------|
| GET | `/login` | `AuthHandler.LoginGET` | `login.html` | Pagina de login com design system CSS |
| GET | `/2fa/setup` | `AuthHandler.TotpSetupGET` | `totp_setup.html` | Configuracao 2FA com QR code, design system CSS |
| GET | `/2fa/verify` | `AuthHandler.TotpVerifyGET` | `totp_verify.html` | Verificacao 2FA TOTP, design system CSS |

### Mudancas nos templates de auth (SPEC-06)

- Adicionar `<link rel="stylesheet" href="/static/css/app.css">` no `<head>`
- Trocar `<style="color:red">` por `<div class="alert alert-error">`
- Trocar `<input>` sem classe por `<input class="form-input">`
- Trocar `<button>` sem classe por `<button class="btn btn-primary">`
- Layout centrado em `<div class="card max-w-md mx-auto mt-20">`
- **Funcionalidade identica**: form actions, method, name attrs, redirects, cookies -- nada muda

---

## Static Assets

| Path | Description |
|------|-------------|
| `/static/css/app.css` | Tailwind CSS compilado (build-time) |
| `/static/js/htmx.min.js` | HTMX 1.9.12 (self-hosted) |
| `/static/js/alpine.min.js` | Alpine.js 3.14.1 (self-hosted) |
| `/static/img/*` | Imagens do site (hero, fundador, etc.) -- novo no SPEC-06 |

---

## Template Data Structures

### pageData (atualizada)

```go
type pageData struct {
    ActivePage   string
    Success      bool
    Form         contactForm
    Errors       contactErrors
    Testimonials []testimonial
    Metrics      []metric       // NEW: faixa de metricas
    Services     []serviceDetail // NEW: lista de servicos para index
    Service      *serviceDetail  // NEW: servico detalhado para /servicos/{slug}
}

type contactForm struct {
    Company string // NEW: campo empresa
    Name    string
    Email   string
    Phone   string
    Subject string
    Message string
}
```
