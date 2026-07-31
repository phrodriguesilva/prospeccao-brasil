# Endpoint Contracts: Institutional Site & Design System

**Date**: 2026-07-31
**Spec**: [spec.md](spec.md)

All institutional endpoints are public (no auth, no session cookie required).
All responses are HTML (server-rendered). Form submissions accept
`application/x-www-form-urlencoded` and return HTML fragments (HTMX) or
redirects (no-JS fallback).

## Public Pages (GET, returns full HTML page)

| Method | Path | Handler | Template | Description |
|--------|------|---------|----------|-------------|
| GET | `/` | `InstitutionalHandler.Home` | `home.html` | Home page with hero, services preview, CTA |
| GET | `/quem-somos` | `InstitutionalHandler.QuemSomos` | `quem-somos.html` | About page with history, mission, team |
| GET | `/servicos` | `InstitutionalHandler.Servicos` | `servicos.html` | Services page with 4+ service cards |
| GET | `/nossos-clientes` | `InstitutionalHandler.NossosClientes` | `nossos-clientes.html` | Clients page with testimonials or empty state |
| GET | `/fale-conosco` | `InstitutionalHandler.FaleConosco` | `fale-conosco.html` | Contact page with form |

## 404 (GET, any unmatched path)

| Method | Path | Handler | Template | Description |
|--------|------|---------|----------|-------------|
| GET | `*` (unmatched) | `InstitutionalHandler.NotFound` | `404.html` | Custom 404 with institutional layout |

## Contact Form (POST, returns HTML fragment or redirect)

| Method | Path | Handler | Request | Response (HTMX) | Response (no-JS) |
|--------|------|---------|---------|-----------------|------------------|
| POST | `/fale-conosco` | `ContactHandler.Submit` | `application/x-www-form-urlencoded`: name, email, phone, subject, message | 200 with `fragments/contact_success.html` or `fragments/contact_error.html` | 302 redirect to `/fale-conosco` with query param `?error=...` or `?success=1` |

**Validation errors** (returned as `contact_error.html` fragment or
`?error=validation` redirect):
- Missing or too-short name (min 2 chars)
- Invalid email format
- Missing or too-short message (min 10 chars)
- Rate limit exceeded (429 status code)

**Success** (returned as `contact_success.html` fragment or
`?success=1` redirect):
- Submission persisted to `contact_submissions` table
- Form cleared (HTMX replaces form with success message)
- slog Info: "contact_submission_created" with submission ID

## Newsletter (POST, returns HTML fragment)

| Method | Path | Handler | Request | Response (HTMX) | Response (no-JS) |
|--------|------|---------|---------|-----------------|------------------|
| POST | `/newsletter` | `NewsletterHandler.Subscribe` | `application/x-www-form-urlencoded`: email | 200 with `fragments/newsletter_success.html` or `fragments/newsletter_error.html` | 302 redirect to referrer with `?newsletter=success` or `?newsletter=error` |

**Responses**:
- New subscription: `newsletter_success.html` with "Inscrição confirmada!"
- Already subscribed: `newsletter_success.html` with "Você já está inscrito!"
- Invalid email: `newsletter_error.html` with "Email inválido"
- Rate limit exceeded: 429 status code

## Static Files

| Path | Description |
|------|-------------|
| `/static/css/app.css` | Compiled Tailwind CSS (build-time) |
| `/static/js/htmx.min.js` | Self-hosted HTMX 1.9.12 |
| `/static/js/alpine.min.js` | Self-hosted Alpine.js 3.14.1 |

## Rate Limiting

All form POST endpoints (`/fale-conosco`, `/newsletter`) are rate-limited
using the `RateLimiter` from SPEC-03 (5 requests per 15 seconds per IP).
Rate-limited requests return 429 Too Many Requests.

## Request/Response Examples

### Contact Form (HTMX, success)

```http
POST /fale-conosco HTTP/1.1
Content-Type: application/x-www-form-urlencoded
HX-Request: true

name=João+Silva&email=joao%40example.com&phone=%2B55+11+99999-9999&subject=Prospecção+comercial&message=Gostaria+de+saber+mais+sobre+imóveis+comerciais+em+São+Paulo
```

```http
HTTP/1.1 200 OK
Content-Type: text/html

<div class="bg-green-50 border border-green-200 rounded-lg p-4">
  <p class="text-green-800">Mensagem enviada com sucesso! Entraremos em contato em breve.</p>
</div>
```

### Newsletter (HTMX, already subscribed)

```http
POST /newsletter HTTP/1.1
Content-Type: application/x-www-form-urlencoded
HX-Request: true

email=joao%40example.com
```

```http
HTTP/1.1 200 OK
Content-Type: text/html

<div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
  <p class="text-blue-800">Você já está inscrito!</p>
</div>
```
