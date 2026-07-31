# Tasks: Institutional Site & Design System

**Input**: Design documents from `/specs/004-institutional-site-design/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are included (85% coverage gate is a constitution requirement).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Migration, sqlc queries, self-hosted JS, and design system CSS classes.

- [ ] T001 Create migration `migrations/2_contact_newsletter.sql` with `contact_submissions` and `newsletter_subscribers` tables (id UUID PK, name, email, phone, subject, message, status CHECK, created_at, updated_at; newsletter: id, email UNIQUE, subscribed_at, active). Add indexes. (FR-014, FR-018, FR-019, data-model.md)
- [ ] T002 Run `make migrate` and `make sqlc` to generate typed Go code for the new tables. Verify `internal/db/models.go` has `ContactSubmission` and `NewsletterSubscriber` structs. (FR-014, FR-018)
- [ ] T003 [P] Download self-hosted HTMX 1.9.12 to `static/js/htmx.min.js` and Alpine.js 3.14.1 to `static/js/alpine.min.js`. Verify both files are non-empty (~48KB and ~44KB). (FR-024, research.md R8)
- [ ] T004 [P] Add design system component classes to `input.css` using `@layer components`: `.btn` (primary, secondary, outline, ghost; sizes sm, md, lg), `.badge` (success, warning, error, info), `.card` (elevation, padding), `.form-input` (text, email, tel, textarea, select with focus/error/disabled states), `.nav-link` (active state). Use Tailwind tokens from `tailwind.config.js`. Run `make build-css` to compile. (FR-001 to FR-007, research.md R1)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Base template, static file serving, and router wiring. MUST be complete before any user story.

- [ ] T005 Create `internal/template/base.html` -- shared layout with `<head>` (meta viewport, title block, link to app.css, script tags for htmx.min.js and alpine.min.js with `defer`), `<nav>` (logo, menu items with active state via template var, mobile hamburger menu via Alpine.js `x-data`/`x-show`), `{{block "content" .}}{{end}}`, `<footer>` (company info, contact links, newsletter form with `hx-post="/newsletter"`, social media links, copyright). (FR-005, FR-006, FR-017, FR-021, FR-024, FR-025, research.md R3, R6, R8)
- [ ] T006 Create `internal/template/partials/nav.html` and `internal/template/partials/footer.html` -- extracted partials included by `base.html` via `{{template "nav" .}}` and `{{template "footer" .}}`. Nav has menu items: Home, Quem somos, Servicos, Nossos clientes, Fale Conosco. Active page highlighted via `.nav-link.active` class based on `{{.ActivePage}}` template var. (FR-005, FR-025)
- [ ] T007 Create `internal/handler/institutional.go` -- `InstitutionalHandler` struct with `queries *db.Queries`, `tmpl *template.Template`. Methods: `Home`, `QuemSomos`, `Servicos`, `NossosClientes`, `FaleConosco`, `NotFound`. Each renders the corresponding template with `base.html` layout and `ActivePage` var. `NotFound` returns 404 status code. (FR-008 to FR-012, FR-022, contracts/endpoints.md)
- [ ] T008 Update `cmd/prospeccao/main.go` -- add institutional routes to chi router (public group): `GET /`, `GET /quem-somos`, `GET /servicos`, `GET /nossos-clientes`, `GET /fale-conosco`, `POST /fale-conosco`, `POST /newsletter`. Add `r.NotFound(authHandler.NotFound)` for 404. Add static file server: `r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))`. Update template loading to parse all templates including partials and fragments. (FR-008 to FR-012, FR-022, FR-024, contracts/endpoints.md)

**Checkpoint**: Foundation ready -- base template, static files, router wired. User story implementation can begin.

---

## Phase 3: User Story 1 - Home page (Priority: P1)

**Goal**: Visitor sees hero, services preview, clients preview, and CTA on the Home page.

**Independent Test**: `curl http://localhost:8080/` returns 200 with hero, services, and CTA content.

### Implementation for User Story 1

- [ ] T009 [US1] Create `internal/template/home.html` -- defines `{{block "content" .}}` with: hero section (headline, subheadline, CTA button "Solicite uma prospecção" linking to `/fale-conosco`), services preview (3 service cards with `.card` class, icon, title, description, "Saiba mais" link to `/servicos`), clients preview (link to `/nossos-clientes` or empty state), and a final CTA section. Uses `.btn`, `.card` design system classes. (FR-008, FR-001, FR-003)
- [ ] T010 [US1] Create `internal/handler/institutional_test.go` -- integration test: `TestHomeGET` verifies 200 status, hero text present, services preview present, CTA link to `/fale-conosco` present. Uses httptest + chi router. (SC-001, SC-007, quickstart.md Scenario 1)

**Checkpoint**: Home page is functional and tested independently.

---

## Phase 4: User Story 2 - Quem somos & Servicos (Priority: P2)

**Goal**: Visitor can navigate to "Quem somos" and "Servicos" pages with rich content.

**Independent Test**: `curl http://localhost:8080/quem-somos` and `/servicos` return 200 with structured content.

### Implementation for User Story 2

- [ ] T011 [P] [US2] Create `internal/template/quem-somos.html` -- defines content block with: company history section, mission/vision/values cards (`.card` class), team section (at least 1 member: Luiz Claudio with photo placeholder, name, role). Uses `.badge` for values, `.card` for team members. (FR-009)
- [ ] T012 [P] [US2] Create `internal/template/servicos.html` -- defines content block with at least 4 service cards: Prospecção de imóveis comerciais, Análise de viabilidade, Relatórios PDF, Gestão de pipeline. Each card has icon, title (`.headline-md`), description (`.body-md`), and CTA button (`.btn .btn-primary`) linking to `/fale-conosco`. (FR-010)
- [ ] T013 [US2] Add tests to `internal/handler/institutional_test.go`: `TestQuemSomosGET` (200, mission/vision/values present, team member present), `TestServicosGET` (200, at least 4 service cards present), `TestNavActiveState` (each page has correct `active` class on nav item). (SC-007, quickstart.md Scenario 2, 12)

**Checkpoint**: Quem somos and Servicos pages are functional and tested.

---

## Phase 5: User Story 3 - Fale Conosco form (Priority: P2)

**Goal**: Visitor can submit a contact form that persists to the database.

**Independent Test**: POST to `/fale-conosco` with valid data returns success and creates a DB record.

### Implementation for User Story 3

- [ ] T014 [US3] Create `internal/handler/contact.go` -- `ContactHandler` struct with `queries`, `tmpl`, `log`, `limiter`. `Submit` method: parses form, validates fields (name min 2, email valid via `net/mail.ParseAddress`, message min 10, phone optional), rate-limits via `limiter.AllowBoth(ip, email)`, persists via `queries.CreateContactSubmission`, logs via slog, returns HTMX fragment (`contact_success.html`) or error fragment (`contact_error.html`). If not HTMX request, redirects with query params. (FR-013 to FR-016, contracts/endpoints.md, research.md R4)
- [ ] T015 [P] [US3] Create `internal/template/fale-conosco.html` -- content block with contact form: name (required, min 2), email (required, type=email), phone (optional, type=tel, placeholder "+55 11 99999-9999"), subject (required), message (required, textarea, min 10). Form uses `hx-post="/fale-conosco"`, `hx-target="#contact-form-container"`, `hx-swap="outerHTML"`. Fallback `action="/fale-conosco" method="POST"`. Uses `.form-input`, `.btn` classes. (FR-012, FR-013, research.md R2)
- [ ] T016 [P] [US3] Create `internal/template/fragments/contact_success.html` and `internal/template/fragments/contact_error.html` -- HTMX response fragments. Success: green alert box "Mensagem enviada com sucesso!". Error: red alert box with validation error messages. Both use Tailwind alert classes. (FR-015, research.md R2)
- [ ] T017 [US3] Create `internal/handler/contact_test.go` -- integration tests: `TestContactSubmitValid` (200, success fragment, DB record created), `TestContactSubmitInvalidEmail` (error fragment), `TestContactSubmitShortMessage` (error fragment), `TestContactSubmitMissingName` (error fragment), `TestContactSubmitRateLimited` (429 after 5 attempts). Uses httptest + real Postgres. (SC-003, SC-007, quickstart.md Scenario 4, 5, 14)

**Checkpoint**: Contact form is functional, validated, persisted, and tested.

---

## Phase 6: User Story 4 - Newsletter (Priority: P3)

**Goal**: Visitor can subscribe to the newsletter from the footer on any page.

**Independent Test**: POST to `/newsletter` with a new email returns success and creates a DB record. Same email twice returns "already subscribed".

### Implementation for User Story 4

- [ ] T018 [US4] Create `internal/handler/newsletter.go` -- `NewsletterHandler` with `queries`, `tmpl`, `log`, `limiter`. `Subscribe` method: parses email, validates via `net/mail.ParseAddress`, rate-limits, tries `queries.CreateNewsletterSubscriber`, catches unique violation (pgx error code 23505) and returns "already subscribed" fragment, logs via slog, returns HTMX fragment. (FR-017 to FR-020, contracts/endpoints.md, research.md R5)
- [ ] T019 [P] [US4] Create `internal/template/fragments/newsletter_success.html` and `internal/template/fragments/newsletter_error.html` -- HTMX fragments. Success: "Inscrição confirmada!" or "Você já está inscrito!". Error: "Email inválido". (FR-019, research.md R5)
- [ ] T020 [US4] Create `internal/handler/newsletter_test.go` -- integration tests: `TestNewsletterSubscribeNew` (success, DB record created), `TestNewsletterSubscribeDuplicate` ("already subscribed", no duplicate), `TestNewsletterSubscribeInvalidEmail` (error fragment), `TestNewsletterSubscribeRateLimited` (429). Uses httptest + real Postgres. (SC-004, SC-007, quickstart.md Scenario 6, 7, 8)

**Checkpoint**: Newsletter form is functional, idempotent, and tested.

---

## Phase 7: User Story 5 - Nossos clientes (Priority: P3)

**Goal**: Visitor sees client testimonials or an elegant empty state.

**Independent Test**: `curl http://localhost:8080/nossos-clientes` returns 200 with testimonials or empty state.

### Implementation for User Story 5

- [ ] T021 [US5] Create `internal/template/nossos-clientes.html` -- content block with: if `{{.Testimonials}}` is non-empty, render testimonial cards (`.card` class with client name, quote, metric badge). If empty, render elegant empty state: centered message "Em breve nossos clientes e cases de sucesso" with a subtle icon. Uses `.badge` for metrics, `.card` for testimonials. (FR-011)
- [ ] T022 [US5] Add test to `internal/handler/institutional_test.go`: `TestNossosClientesGET` (200, empty state or testimonials present). (SC-007, quickstart.md Scenario 2)

**Checkpoint**: Nossos clientes page is functional and tested.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: 404 page, final CSS build, full test run, CI verification.

- [ ] T023 [P] Create `internal/template/404.html` -- content block with friendly 404 message "Página não encontrada", a link back to Home (`.btn .btn-primary`), and institutional header/footer from base template. (FR-022)
- [ ] T024 [P] Add test to `internal/handler/institutional_test.go`: `TestNotFound` (404 status, "não encontrada" text, nav and footer present). (FR-022, quickstart.md Scenario 3)
- [ ] T025 Run `make build-css` to compile final CSS with all design system component classes. Verify `.btn`, `.badge`, `.card`, `.form-input` classes exist in `static/css/app.css`. (FR-001 to FR-007, quickstart.md Scenario 9)
- [ ] T026 Run `make check` and verify: golangci-lint 0 issues, all tests pass, coverage >= 85% (excluding internal/db, cmd/prospeccao, scripts), build succeeds, ast-grep 0 errors. Fix any failures. (SC-007, quickstart.md Scenario 15)
- [ ] T027 Run all 16 quickstart.md validation scenarios and verify each passes. Document any failures and fix. (All FRs, quickstart.md)
- [ ] T028 Update `tailwind.config.js` content paths to include `internal/template/**/*.html` (currently points to `internal/ui/templates/` and `internal/handler/templates/` which don't exist). Ensure all template files are scanned by Tailwind for class purging. (FR-001 to FR-007)
- [ ] T029 Commit all changes. Push to main. Verify CI passes via `gh run watch`. Check that the Test step runs integration tests against Postgres service container and the coverage gate passes. (SC-007)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies. T001 first (migration), T002 depends on T001. T003, T004 parallel.
- **Foundational (Phase 2)**: Depends on Phase 1. T005 first (base template), T006 depends on T005. T007 depends on T005. T008 depends on T007.
- **User Stories (Phase 3-7)**: All depend on Phase 2 completion.
  - US1 (Phase 3): T009 depends on T008. T010 depends on T009.
  - US2 (Phase 4): T011, T012 parallel (depend on T008). T013 depends on T011, T012.
  - US3 (Phase 5): T014 depends on T008, T002. T015, T016 parallel (depend on T005). T017 depends on T014, T015, T016.
  - US4 (Phase 6): T018 depends on T008, T002. T019 depends on T005. T020 depends on T018, T019.
  - US5 (Phase 7): T021 depends on T008. T022 depends on T021.
- **Polish (Phase 8)**: Depends on all user stories. T023, T024 parallel. T025 depends on T004. T026 depends on all. T027 depends on T026. T028 can run anytime. T029 depends on T026.

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational only. No dependencies on other stories.
- **US2 (P2)**: Depends on Foundational only. Independent of US1.
- **US3 (P2)**: Depends on Foundational + migration (T001/T002). Independent of US1, US2.
- **US4 (P3)**: Depends on Foundational + migration (T001/T002). Independent of US1, US2, US3.
- **US5 (P3)**: Depends on Foundational only. Independent of all other stories.

### Parallel Opportunities

- T003, T004 can run in parallel (Setup)
- T005, T006 can run in parallel after T005 (Foundational -- actually T006 depends on T005)
- T011, T012 can run in parallel (US2 templates)
- T015, T016 can run in parallel (US3 templates + fragments)
- T019 can run in parallel with T018 (US4)
- T023, T024 can run in parallel (Polish)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (migration, sqlc, JS, CSS)
2. Complete Phase 2: Foundational (base template, router, static files)
3. Complete Phase 3: User Story 1 (Home page)
4. **STOP and VALIDATE**: `curl http://localhost:8080/` shows Home with hero, services, CTA
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational -> Foundation ready
2. Add US1 (Home) -> Test independently -> Deploy
3. Add US2 (Quem somos + Servicos) -> Test independently -> Deploy
4. Add US3 (Fale Conosco) -> Test independently -> Deploy
5. Add US4 (Newsletter) -> Test independently -> Deploy
6. Add US5 (Nossos clientes) -> Test independently -> Deploy
7. Polish (404, CSS, make check, CI) -> Final verification
