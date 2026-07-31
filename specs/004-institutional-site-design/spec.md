# Feature Specification: Institutional Site & Design System

**Feature Branch**: `004-institutional-site-design`

**Created**: 2026-07-31

**Status**: Draft

**Input**: User description: "SPEC-04: Institutional Site & Design System. Design system component classes (buttons, badges, cards, forms, nav, footer) built on top of the Tailwind tokens from SPEC-01, AND the institutional pages (Home, Quem somos, Servicos, Nossos clientes, Fale Conosco, Newsletter)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Visitante acessa a Home (Priority: P1)

Um visitante chega ao site institucional da Prospecção Brasil pela
primeira vez. Ele vê um hero section com a proposta de valor da empresa,
uma seção de serviços em destaque, uma seção de clientes (logos ou
depoimentos), e uma chamada para ação (CTA) para "Fale Conosco" ou
"Solicite uma prospecção". A página carrega rapidamente (server-rendered
HTML, sem JavaScript pesado), tem navegação clara no header, e um footer
com informações de contato e links para redes sociais.

**Why this priority**: A Home é a porta de entrada. Sem ela, nenhum outro
conteúdo institucional faz sentido. É a primeira impressão de credibilidade
para um prospect que pesquisa a empresa antes de fechar negócio.

**Independent Test**: Pode ser testada abrindo `http://localhost:8080/` e
verificando que o hero, serviços, clientes e CTA estão visíveis e que a
navegação funciona.

**Acceptance Scenarios**:

1. **Given** que sou um visitante anônimo, **When** acesso a Home, **Then** vejo um hero com a proposta de valor, pelo menos 3 serviços em destaque, e um CTA para "Fale Conosco"
2. **Given** que sou um visitante em um dispositivo móvel, **When** acesso a Home, **Then** o layout se adapta (responsive) e o menu de navegação colapsa em um hamburger menu
3. **Given** que sou um visitante, **When** clico em "Solicite uma prospecção" no hero, **Then** sou redirecionado para a página "Fale Conosco" com o formulário de contato

---

### User Story 2 - Visitante navega para "Quem somos" e "Servicos" (Priority: P2)

Um visitante interessado quer entender melhor a empresa. Ele clica em
"Quem somos" no menu e vê a história da empresa, missão, visão, valores,
e a equipe (pelo menos o fundador Luiz Claudio). Depois clica em
"Servicos" e vê uma lista detalhada dos serviços oferecidos: prospecção
de imóveis comerciais, análise de viabilidade, relatórios PDF, gestão de
pipeline de clientes, etc. Cada serviço tem um ícone, título, descrição,
e um CTA para "Saiba mais" ou "Fale Conosco".

**Why this priority**: "Quem somos" e "Servicos" constroem a credibilidade
que um prospect precisa antes de iniciar contato. Sem elas, o site é uma
landing page oca.

**Independent Test**: Pode ser testada navegando para `/quem-somos` e
`/servicos` e verificando que o conteúdo está presente, bem estruturado,
e usa os componentes do design system (cards, badges, ícones).

**Acceptance Scenarios**:

1. **Given** que sou um visitante, **When** acesso `/quem-somos`, **Then** vejo a história, missão, visão, valores, e pelo menos 1 membro da equipe
2. **Given** que sou um visitante, **When** acesso `/servicos`, **Then** vejo pelo menos 4 serviços com ícone, título, descrição, e CTA
3. **Given** que sou um visitante em qualquer página institucional, **When** clico em um item do menu de navegação, **Then** sou redirecionado para a página correspondente e o item ativo é destacado no menu

---

### User Story 3 - Visitante envia mensagem via "Fale Conosco" (Priority: P2)

Um visitante decide entrar em contato. Ele acessa "Fale Conosco", preenche
um formulário com nome, email, telefone, assunto, e mensagem. O sistema
valida os campos (email válido, telefone opcional, mensagem mínima de 10
caracteres). Ao submeter, o sistema exibe uma mensagem de sucesso e
registra a mensagem no banco de dados (tabela `contact_submissions` para
auditoria e follow-up). O admin pode ver as submissões no sistema interno
(SPEC-05).

**Why this priority**: O formulário de contato é a principal conversão do
site institucional. Sem ele, o site gera tráfego mas não captura leads.

**Independent Test**: Pode ser testada preenchendo e submetendo o
formulário em `/fale-conosco` e verificando que a mensagem de sucesso
aparece e que o registro aparece no banco de dados.

**Acceptance Scenarios**:

1. **Given** que sou um visitante, **When** acesso `/fale-conosco`, **Then** vejo um formulário com campos: nome, email, telefone (opcional), assunto, mensagem
2. **Given** que preenchi o formulário com dados válidos, **When** clico em "Enviar", **Then** vejo uma mensagem de sucesso e o registro é persistido no banco de dados
3. **Given** que preenchi o formulário com email inválido, **When** clico em "Enviar", **Then** vejo uma mensagem de erro de validação e o formulário não é submetido
4. **Given** que não preenchi o campo mensagem, **When** clico em "Enviar", **Then** vejo uma mensagem de erro indicando que a mensagem é obrigatória

---

### User Story 4 - Visitante se inscreve na Newsletter (Priority: P3)

Um visitante quer receber atualizações sobre novos imóveis comerciais
disponíveis. Ele preenche apenas o email no formulário de newsletter
(presente no footer de todas as páginas). O sistema valida o email,
verifica se já está inscrito (idempotente), e registra a inscrição no
banco de dados (tabela `newsletter_subscribers`). Se já estiver inscrito,
exibe "Você já está inscrito!" em vez de uma mensagem de erro.

**Why this priority**: Newsletter é um canal de aquisição secundário. É
importante mas não bloqueia a credibilidade do site.

**Independent Test**: Pode ser testada preenchendo o email no formulário
de newsletter no footer e verificando que a inscrição é registrada e que
submeter o mesmo email duas vezes não cria duplicatas.

**Acceptance Scenarios**:

1. **Given** que sou um visitante em qualquer página, **When** vejo o footer, **Then** há um formulário de newsletter com campo de email e botão "Inscrever"
2. **Given** que preenchi um email válido e não inscrito, **When** clico em "Inscrever", **Then** vejo "Inscrição confirmada!" e o email é persistido na tabela `newsletter_subscribers`
3. **Given** que preenchi um email já inscrito, **When** clico em "Inscrever", **Then** vejo "Você já está inscrito!" e nenhum registro duplicado é criado
4. **Given** que preenchi um email inválido, **When** clico em "Inscrever", **Then** vejo uma mensagem de erro de validação

---

### User Story 5 - Visitante acessa "Nossos clientes" (Priority: P3)

Um visitante quer ver prova social. Ele clica em "Nossos clientes" e vê
uma página com logos ou nomes de clientes atendidos, depoimentos em
formato de cards, e métricas de resultado (ex: "50+ imóveis prospectados",
"R$ 100M+ em negócios intermediados"). Se ainda não houver clientes
cadastrados, a página mostra um estado vazio elegante ("Em breve nossos
clientes e cases de sucesso").

**Why this priority**: Prova social acelera a conversão, mas o MVP pode
funcionar sem ela (o formulário de contato é mais importante).

**Independent Test**: Pode ser testada acessando `/nossos-clientes` e
verificando que a página renderiza com o estado vazio ou com depoimentos.

**Acceptance Scenarios**:

1. **Given** que sou um visitante, **When** acesso `/nossos-clientes`, **Then** vejo depoimentos em cards ou um estado vazio elegante
2. **Given** que há depoimentos cadastrados, **When** acesso a página, **Then** cada depoimento tem nome do cliente, texto, e métrica de resultado

---

### Edge Cases

- What happens when the contact form submission fails (DB error)? The system shows a generic error message and logs the error via slog. The form data is NOT lost (the user can retry).
- What happens when a bot submits the newsletter form with a disposable email? The system accepts it (no domain blacklist in MVP) but logs it for future review.
- What happens when a visitor accesses a non-existent institutional page (e.g., `/blog`)? The system returns a 404 page with the institutional header/footer and a link back to Home.
- What happens when the newsletter form is submitted via HTMX (no full page reload)? The system returns an HTML fragment with the success/error message, replacing the form.
- What happens when the contact form is submitted with HTML in the message field? The system escapes HTML (Go html/template auto-escaping) to prevent XSS.
- What happens when the admin has not seeded any client testimonials? The "Nossos clientes" page shows an elegant empty state, not a blank page or error.

## Requirements *(mandatory)*

### Functional Requirements

**Design System Components:**

- **FR-001**: System MUST provide reusable CSS component classes for buttons (primary, secondary, outline, ghost; sizes sm, md, lg) using the Tailwind tokens from SPEC-01
- **FR-002**: System MUST provide reusable CSS component classes for badges (success, warning, error, info) with consistent padding, border-radius, and color
- **FR-003**: System MUST provide reusable CSS component classes for cards (elevation, padding, border) using the shadow and surface tokens
- **FR-004**: System MUST provide reusable CSS component classes for form inputs (text, email, tel, textarea, select) with focus states, error states, and disabled states
- **FR-005**: System MUST provide a navigation bar component (sticky header, logo, menu items, active state highlighting, mobile hamburger menu via Alpine.js)
- **FR-006**: System MUST provide a footer component (company info, contact links, newsletter form, social media links, copyright)
- **FR-007**: All component classes MUST be defined in `input.css` using `@layer components` and compiled by Tailwind CSS build-time

**Institutional Pages:**

- **FR-008**: System MUST serve a Home page at `/` with hero section, services preview, clients preview, and CTA to "Fale Conosco"
- **FR-009**: System MUST serve a "Quem somos" page at `/quem-somos` with company history, mission, vision, values, and team section
- **FR-010**: System MUST serve a "Servicos" page at `/servicos` with at least 4 service cards (icon, title, description, CTA)
- **FR-011**: System MUST serve a "Nossos clientes" page at `/nossos-clientes` with testimonials in cards or an elegant empty state
- **FR-012**: System MUST serve a "Fale Conosco" page at `/fale-conosco` with a contact form (nome, email, telefone opcional, assunto, mensagem)

**Contact Form:**

- **FR-013**: System MUST validate contact form fields: nome (min 2 chars), email (valid format), mensagem (min 10 chars), telefone (optional, Brazilian format)
- **FR-014**: System MUST persist contact form submissions in a `contact_submissions` table with fields: id, name, email, phone, subject, message, created_at, status (new, read, archived)
- **FR-015**: System MUST display a success message after form submission and clear the form
- **FR-016**: System MUST log contact form submissions via slog for audit trail

**Newsletter:**

- **FR-017**: System MUST provide a newsletter signup form in the footer of all institutional pages (email field + "Inscrever" button)
- **FR-018**: System MUST persist newsletter subscriptions in a `newsletter_subscribers` table with fields: id, email, subscribed_at, active
- **FR-019**: System MUST be idempotent for newsletter subscriptions: submitting the same email twice does NOT create a duplicate record
- **FR-020**: System MUST validate the newsletter email field (valid email format) before persisting

**Infrastructure:**

- **FR-021**: All institutional pages MUST use a shared base template (`base.html`) with the navigation bar, footer, and newsletter form
- **FR-022**: System MUST return a 404 page with institutional header/footer for non-existent routes
- **FR-023**: All institutional pages MUST be public (no auth required) and accessible without a session cookie
- **FR-024**: System MUST self-host HTMX and Alpine.js in `static/js/` (no CDN dependencies) per AGENTS.md self-hosting rule
- **FR-025**: Navigation bar MUST highlight the active page using server-side template logic (not client-side JS)

### Key Entities *(include if feature involves data)*

- **ContactSubmission**: Represents a message sent via the "Fale Conosco" form. Attributes: id (UUID), name, email, phone (optional), subject, message, created_at, status (new/read/archived). No tenant_id (public form, not tenant-scoped).
- **NewsletterSubscriber**: Represents an email subscribed to the newsletter. Attributes: id (UUID), email (unique), subscribed_at, active (boolean). No tenant_id (public form, not tenant-scoped).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 6 institutional pages (Home, Quem somos, Servicos, Nossos clientes, Fale Conosco, 404) render in under 200ms server-side (measured via `curl -w "%{time_total}"`)
- **SC-002**: The design system has at least 6 reusable component classes (button, badge, card, form-input, nav, footer) that are used across all institutional pages
- **SC-003**: A visitor can complete the contact form submission flow (fill + submit + see success) in under 30 seconds
- **SC-004**: A visitor can subscribe to the newsletter in under 10 seconds (email + click)
- **SC-005**: All institutional pages pass the Lighthouse accessibility audit with a score of 90+ (semantic HTML, alt text, color contrast)
- **SC-006**: The site is fully responsive: all pages pass the Lighthouse mobile audit with a score of 90+
- **SC-007**: Test coverage for the institutional site handlers and form validation is >= 85%

## Assumptions

- The Tailwind design tokens from SPEC-01 (tailwind.config.js) are the canonical source of truth for colors, typography, spacing, shadows, and border-radius. No new tokens are added in this spec.
- The institutional site content (copy, service descriptions, team info) is provided by Luiz Claudio or is placeholder text that can be edited later. The spec defines the structure and components, not the final copy.
- HTMX and Alpine.js are self-hosted in `static/js/` (downloaded and committed, not loaded from CDN). The versions are pinned: HTMX 1.9.12, Alpine.js 3.14.1.
- The contact form and newsletter form use HTMX for async submission (no full page reload). If JavaScript is disabled, the forms fall back to standard POST with full page reload.
- The `contact_submissions` and `newsletter_subscribers` tables are NOT tenant-scoped (they are public forms on the institutional site, not internal system features). This is intentional -- the institutional site is public and has no tenant context.
- The 404 page uses the institutional base template (header + footer) so visitors never see a raw "Not Found" text.
- The admin user (from SPEC-03 seed) can view contact submissions in the internal system (SPEC-05), but that viewing UI is out of scope for this spec.
- No email sending is implemented in this spec. Contact form submissions are persisted to the DB for the admin to review. Email notifications are a future feature.
- The site is in Portuguese (pt-BR). All UI text, labels, and messages are in Portuguese.
- The design system component classes are defined in `input.css` using `@layer components` and compiled by Tailwind. They are NOT Go template functions -- they are CSS classes applied in HTML.

## Data Contract

Generated by `speckit-tekimax-security-data-contract` hook (optional but
executed for SPEC-04: introduces PII data entities -- contact form collects
name, email, phone under LGPD). Adapted for Go + Postgres + sqlc.

### Sources

| Name | Origin | Trust | Schema Location | PII? |
|------|--------|-------|-----------------|------|
| contact_submissions | DB table (new migration) | unvetted (public form) | `migrations/` + sqlc | name, email, phone, message |
| newsletter_subscribers | DB table (new migration) | unvetted (public form) | `migrations/` + sqlc | email |

### Schema Definition

Two new tables added via forward-only SQL migration:

**contact_submissions** (NOT tenant-scoped -- public institutional form):
- `id` UUID PK (gen_random_uuid())
- `name` VARCHAR(255) NOT NULL
- `email` VARCHAR(255) NOT NULL
- `phone` VARCHAR(20) NULL (optional, Brazilian format)
- `subject` VARCHAR(255) NOT NULL
- `message` TEXT NOT NULL
- `status` VARCHAR(20) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'read', 'archived'))
- `created_at` TIMESTAMPTZ NOT NULL DEFAULT now()
- `updated_at` TIMESTAMPTZ NOT NULL DEFAULT now()

**newsletter_subscribers** (NOT tenant-scoped -- public institutional form):
- `id` UUID PK (gen_random_uuid())
- `email` VARCHAR(255) NOT NULL UNIQUE (idempotency via unique constraint)
- `subscribed_at` TIMESTAMPTZ NOT NULL DEFAULT now()
- `active` BOOLEAN NOT NULL DEFAULT true

### PII Handling

LGPD-sensitive fields (Constitution principle II). These tables collect PII
from public forms (no auth, no tenant context). PII is stored in Postgres
(volume-level encryption at rest). No app-layer encryption for these fields
(they are not used for lookup/hashing like sessions -- they are displayed
to the admin in the internal system).

| Field | Table | Strategy | Implementation |
|-------|-------|----------|----------------|
| name | contact_submissions | plaintext (volume encryption) | Postgres volume encryption at rest |
| email | contact_submissions, newsletter_subscribers | plaintext (volume encryption) | Unique constraint on newsletter_subscribers.email for idempotency |
| phone | contact_submissions | plaintext (volume encryption) | Optional field, Brazilian format (+55 11 99999-9999) |
| message | contact_submissions | plaintext (volume encryption) | HTML-escaped on display (Go html/template auto-escaping) |

### Data Retention

- Contact submissions: retained indefinitely (audit trail). Admin can
  archive (status='archived') but not delete (forward-only, LGPD audit).
- Newsletter subscribers: retained until unsubscribe (active=false). No
  automatic deletion -- the admin can set active=false via the internal
  system (SPEC-05). LGPD deletion requests are handled manually via a
  future "right to be forgotten" migration (out of scope for MVP).

### Input Validation

- All form inputs are validated server-side (Go) before DB insertion.
- Client-side validation (HTML5 required, pattern, minlength) is a
  progressive enhancement -- server-side is the source of truth.
- HTML is escaped on display via Go html/template auto-escaping (XSS
  prevention). No user input is rendered as raw HTML.
- Rate limiting on form submission endpoints (reuse RateLimiter from
  SPEC-03, per-IP, 5 requests per 15 seconds).
