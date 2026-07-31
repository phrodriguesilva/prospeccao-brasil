# Research: Institutional Site & Design System

**Date**: 2026-07-31
**Spec**: [spec.md](spec.md)

## R1: Tailwind @layer components for design system

**Decision**: Use `@layer components` in `input.css` to define reusable
CSS component classes (`.btn`, `.badge`, `.card`, `.form-input`, `.nav`,
`.footer`).

**Rationale**: Tailwind's `@layer components` allows defining component
classes that use Tailwind utilities and tokens. These classes are
compiled by `npx tailwindcss` at build time (same as SPEC-01's
`make build-css`). No runtime CSS framework needed. The classes are
pure CSS -- no JavaScript, no Go template functions. This is the
idiomatic Tailwind approach for design systems.

**Alternatives considered**:
- Go template functions (e.g., `{{button "primary" "Click me"}}`):
  rejected -- mixes logic with presentation, harder to maintain, not
  idiomatic for Tailwind.
- CSS-in-JS (e.g., styled-components): rejected -- not applicable to
  server-rendered Go, adds runtime overhead.
- Separate CSS files per component: rejected -- Tailwind's `@layer`
  is cleaner and ensures proper cascade order.

## R2: HTMX for async form submission

**Decision**: Use HTMX for contact form and newsletter form submission.
Forms use `hx-post` to submit async and `hx-target` to replace the form
with a success/error fragment. If JavaScript is disabled, the form's
native `action` and `method` attributes provide a full-page-reload
fallback.

**Rationale**: HTMX provides progressive enhancement -- the forms work
without JavaScript (standard POST) and enhance with JavaScript (async
partial update). This is the AGENTS.md convention (HTMX for
interactivity, no SPA). Self-hosted HTMX 1.9.12 in `static/js/`.

**Alternatives considered**:
- Full page reload for all forms: rejected -- poor UX, causes layout
  shift on success/error.
- fetch() + custom JavaScript: rejected -- reinvents HTMX, more code,
  harder to maintain.
- Alpine.js for form submission: rejected -- Alpine.js is for
  micro-state (dropdowns, modals), not HTTP requests. HTMX is the
  right tool for async HTTP.

## R3: Alpine.js for mobile hamburger menu

**Decision**: Use Alpine.js for the mobile hamburger menu toggle in the
navigation bar. The menu state (open/closed) is managed by Alpine.js
`x-data` and `x-show` directives.

**Rationale**: Alpine.js is the AGENTS.md convention for micro-state
(UI toggles, dropdowns). The hamburger menu is a simple show/hide
toggle -- perfect for Alpine.js. Self-hosted Alpine.js 3.14.1 in
`static/js/`.

**Alternatives considered**:
- CSS-only (checkbox hack): rejected -- accessibility issues, can't
  handle focus trapping.
- HTMX: rejected -- HTMX is for HTTP requests, not client-side state.
- Vanilla JS: rejected -- more code, harder to maintain, Alpine.js is
  already a dependency.

## R4: Contact form validation strategy

**Decision**: Server-side validation in Go is the source of truth.
Client-side validation (HTML5 `required`, `pattern`, `minlength`) is a
progressive enhancement. The handler validates all fields, returns
error fragments (HTMX) or redirects with error messages (no-JS fallback).

**Validation rules**:
- nome: required, min 2 chars, max 255 chars
- email: required, valid email format (Go `net/mail.ParseAddress`)
- telefone: optional, Brazilian phone format (regex: `^\+?55?\s?\d{2}\s?\d{4,5}-?\d{4}$` or similar, lenient)
- assunto: required, min 2 chars, max 255 chars
- mensagem: required, min 10 chars, max 5000 chars

**Rationale**: Server-side validation is non-negotiable (security, data
integrity). Client-side validation improves UX but is bypassable. HTML
auto-escaping prevents XSS. Rate limiting (5/15s per IP) prevents abuse.

**Alternatives considered**:
- Client-side only: rejected -- insecure, bypassable.
- Separate validation library: rejected -- Go stdlib + regex is
  sufficient for this scope.

## R5: Newsletter idempotency via UNIQUE constraint

**Decision**: Add a `UNIQUE` constraint on `newsletter_subscribers.email`.
The handler catches the unique violation (pgx `ErrCode 23505`) and
returns "Você já está inscrito!" instead of an error.

**Rationale**: Database-level uniqueness is the most reliable way to
prevent duplicates. Application-level checks have race conditions
(check-then-insert). The UNIQUE constraint is atomic.

**Alternatives considered**:
- Application-level check (SELECT then INSERT): rejected -- race
  condition between check and insert.
- UPSERT (ON CONFLICT DO NOTHING): considered -- works but the handler
  still needs to know if it was a new insert or an existing subscriber
  to show the right message. Catching the error is simpler.

## R6: Template structure (base.html + page templates)

**Decision**: Use Go `html/template` with a `base.html` layout that
defines blocks (`{{block "content" .}}{{end}}`, `{{block "title" .}}{{end}}`).
Each page template (`home.html`, `quem-somos.html`, etc.) defines the
content block. The base template includes the nav, footer, and newsletter
form. Templates are parsed with `template.ParseGlob` and executed with
`template.ExecuteTemplate`.

**Rationale**: Go's `html/template` supports template inheritance via
`{{block}}`. This is the idiomatic Go approach for server-rendered HTML.
No external template engine needed. Auto-escaping prevents XSS.

**Alternatives considered**:
- templ (type-safe Go templates): considered -- good but adds a new
  dependency and code generation step. html/template is already used
  in SPEC-03.
- Handlebars/EJS: rejected -- not Go, requires JavaScript runtime.

## R7: 404 page with institutional layout

**Decision**: Use chi's `NotFound` handler to serve a custom 404 page
that uses the institutional `base.html` template. The 404 page shows a
friendly message and a link back to Home.

**Rationale**: chi's `r.NotFound(handler)` sets a custom 404 handler.
Using the same base template ensures consistent navigation/footer.

## R8: Self-hosting HTMX and Alpine.js

**Decision**: Download HTMX 1.9.12 and Alpine.js 3.14.1 minified files
to `static/js/`. Reference them in `base.html` with `<script src="/static/js/htmx.min.js" defer></script>` and `<script src="/static/js/alpine.min.js" defer></script>`. Serve `static/` via chi's `http.FileServer`.

**Rationale**: AGENTS.md mandates self-hosting (no CDN SPOF). The files
are ~92KB total. Pinned versions prevent supply chain attacks from CDN.

**Alternatives considered**:
- CDN (unpkg, jsdelivr): rejected -- AGENTS.md prohibits, SPOF risk.
- npm package: rejected -- Tailwind is already the only npm dep, adding
  HTMX/Alpine.js to npm is unnecessary (they're static files, not
  build-time deps).
