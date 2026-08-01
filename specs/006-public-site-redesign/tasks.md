# Tasks: Public Site Redesign

**Input**: Design documents from `/specs/006-public-site-redesign/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are included (70% coverage gate for app code, 85% for internal/auth, per constitution).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Migration, sqlc, Pencil designs, static images, design system classes. MUST be complete before any user story HTML work.

- [ ] T001 Create migration `migrations/000003_add_company_to_contact_submissions.up.sql` and `.down.sql`. Up: `ALTER TABLE contact_submissions ADD COLUMN company VARCHAR(255) NULL;`. Down: `ALTER TABLE contact_submissions DROP COLUMN IF EXISTS company;`. Run `migrate -path migrations -database "$DATABASE_URL" up` to verify. (data-model.md, R10)
- [ ] T002 Update `internal/db/queries/contacts.sql` -- add `company` parameter to `CreateContactSubmission` query: `INSERT INTO contact_submissions (name, email, phone, company, subject, message) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;`. Run `make sqlc` to regenerate typed Go code. Verify `internal/db/contacts.sql.go` has `Company *string` in `CreateContactSubmissionParams`. (data-model.md, contracts/endpoints.md)
- [ ] T003 Update `internal/handler/contact.go` -- add `Company` field to `contactForm` struct. In `Submit` method, read `r.FormValue("company")` and pass to `CreateContactSubmissionParams.Company`. Update `contactErrors` if needed. (contracts/endpoints.md, R10)
- [ ] T004 [P] Add `.alert` and `.alert-error` classes to `input.css` if not already present (check first -- admin CSS may have added them). Add: `.alert { @apply rounded-md p-4 text-body-md border; }` and `.alert-success { @apply bg-green-50 text-green-800 border-green-200; }` if not present. Run `make build-css`. (R6)
- [ ] T005 [P] Create `static/img/` directory with `.gitkeep`. Add hero placeholder image `static/img/hero-comercial.jpg` (stock photo of commercial real estate, or solid Deep Navy `#031636` 1920x1080 JPG as fallback). (R7, FR-030)
- [ ] T006 Create Pencil designs in `designs/prospeccao.pen` using the Pencil MCP server. Frames required (all using brand tokens: Deep Navy primary, Sóbrio Gold secondary, Montserrat display, Inter body): "Home - Desktop" (1440px, hero full-bleed + metrics + services + testimonials + CTA), "Home - Mobile" (375px, stacked), "Servicos Index - Desktop" (1440px, 5 service cards), "Servicos Detalhe - Desktop" (1440px, methodology + CTA), "Quem Somos - Desktop" (1440px, founder + mission/vision/values + CRECI), "Nossos Clientes - Desktop" (1440px, testimonials + metrics), "Fale Conosco - Desktop" (1440px, 2-col form + contact info), "Login - Desktop" (1440px, already done -- reference only). These are the visual source of truth for all HTML implementation. (FR-027, AGENTS.md, R9)

**Checkpoint**: Migration applied, sqlc regenerated, contact handler accepts company, design system ready, Pencil designs exist. User story HTML can begin.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Static data structures (services map, testimonials, metrics) and reusable partials. MUST be complete before any user story page.

- [ ] T007 Update `internal/handler/institutional.go` -- add `serviceDetail` struct with fields: Slug, Title, Summary, Description, Methodology ([]string), CTA. Add `services` map with 5 entries: `expansao-de-redes` (Plano Diretor, Macrolocalizacao, Microlocalizacao, Prospecacao de Ponto), `built-to-suit` (BTS, aquisicao/venda de terrenos, permuta, desenvolvimento), `strip-mall` (Centros de Conveniencia, mix planning, comercializacao), `lajes-comerciais` (Lajes Corporativas, locacao, sale & leaseback), `prospeccao-de-ponto` (requisitos fisicos/meradologicos, levantamento de valores, abordagem, negociacao). Content adapted from legacy site. (R2, data-model.md, FR-006, FR-007)
- [ ] T008 Update `internal/handler/institutional.go` -- add `metric` struct with Label, Value. Add `defaultMetrics` slice with 4 entries: Pontos Comercializados, Clientes Atendidos, Cidades Atendidas, Anos de Mercado (values can be "0" or placeholders). Add `defaultTestimonials` slice with 3 entries from legacy: Larissa Mello, Roberto Andrade, Joao Viana (quotes from legacy site). (R3, R4, data-model.md, FR-002, FR-004, FR-012)
- [ ] T009 Update `internal/handler/institutional.go` -- extend `pageData` struct with `Metrics []metric`, `Services []serviceDetail`, `Service *serviceDetail` fields. Update all existing handler methods (Home, QuemSomos, Servicos, NossosClientes, FaleConosco) to populate Metrics and Testimonials in pageData. (contracts/endpoints.md, R3, R4)
- [ ] T010 [P] Create `internal/template/partials/metrics.html` -- reusable metrics strip partial. Receives `{{.Metrics}}` (slice of metric structs). Renders 4-column grid (desktop) / 2-column grid (mobile) with large value (font-display, text-primary) and label below (text-on-surface-variant). Use `{{template "metrics" .Metrics}}` to include. (R3, FR-002)

**Checkpoint**: Static data ready, metrics partial exists. Page implementation can begin.

---

## Phase 3: User Story 1 - Home (Priority: P1)

**Goal**: Home premium with hero, metrics, services, testimonials, CTA.

**Independent Test**: Open `/` -- verify hero with photo, headline de mercado, metrics, services, testimonials, CTA. No "carga cognitiva" or "pipeline" copy.

### Implementation for User Story 1

- [ ] T011 [US1] Rewrite `internal/template/home.html` -- per Pencil frame "Home - Desktop". Structure: (1) Hero section full-bleed with `background-image: url('/static/img/hero-comercial.jpg')` on a `<div>` overlay, gradient `from-primary/90 to-primary/50`, headline "Encontramos o ponto comercial ideal para a sua rede" (text-on-primary), subheadline "RETAIL SERVICE PARA GRANDES EMPRESAS" or "Prospecção imobiliária comercial" (text-on-primary/80), CTA button "Solicite uma apresentação" (btn-secondary btn-lg) linking to /fale-conosco. (2) Metrics strip: `{{template "metrics" .Metrics}}`. (3) Services section: 3+ cards from `{{.Services}}` with title, summary, link to `/servicos/{{.Slug}}`. (4) Testimonials section: 2+ from `{{.Testimonials}}` with quote, name, company. (5) Final CTA section with "Solicite uma apresentação" or "Fale Conosco". No "carga cognitiva", "pipeline", "plataforma", "software", "relatorios pdf". (FR-001 to FR-005, R1, R5)
- [ ] T012 [US1] Update `internal/handler/institutional_test.go` -- update `TestHome` assertions: verify 200, verify "ponto comercial" in body (market copy), verify NO "carga cognitiva" in body, verify NO "pipeline" in body, verify metrics labels present, verify at least one testimonial name present. (R12, SC-004)

**Checkpoint**: Home is functional, premium, market-facing. No software copy.

---

## Phase 4: User Story 2 - Servicos (Priority: P2)

**Goal**: Index page + 5 deep service pages with methodology.

**Independent Test**: Open `/servicos` -- verify 5 services. Click each -- verify detail page with methodology. Open invalid slug -- verify 404.

### Implementation for User Story 2

- [ ] T013 [US2] Add `ServicoDetalhe` method to `internal/handler/institutional.go` -- reads `{slug}` from chi URL param (`chi.URLParam(r, "slug")`), looks up in `services` map. If found, renders `servico-detalhe.html` with `pageData{Service: &detail, ActivePage: "servicos"}`. If not found, calls `NotFound` (404). (R2, FR-008, contracts/endpoints.md)
- [ ] T014 [US2] Update `cmd/prospeccao/main.go` -- add `r.Get("/servicos/{slug}", instHandler.ServicoDetalhe)` to `buildPublicRouter` AFTER `r.Get("/servicos", instHandler.Servicos)`. Also add to `buildDevRouter` in the same position. (FR-026, contracts/endpoints.md)
- [ ] T015 [P] [US2] Rewrite `internal/template/servicos.html` -- per Pencil frame "Servicos Index - Desktop". Index page with 5 service cards (from `{{.Services}}`): each card has title, summary, "Saiba mais" link to `/servicos/{{.Slug}}`. No generic icons for "Relatorios PDF" or "Gestao de Pipeline" (removed). No "carga cognitiva" or "pipeline". (FR-006, R5)
- [ ] T016 [P] [US2] Create `internal/template/servico-detalhe.html` -- per Pencil frame "Servicos Detalhe - Desktop". Page for `/servicos/{slug}`: (1) Hero/title section with `{{.Service.Title}}` and `{{.Service.Summary}}`. (2) Description section with `{{.Service.Description}}`. (3) Methodology section: `{{range .Service.Methodology}}` rendering each step as a numbered item or card. (4) CTA section: "Fale com um especialista" (btn-primary btn-lg) linking to /fale-conosco. Uses nav and footer partials. (FR-007, R2)
- [ ] T017 [US2] Update `internal/handler/institutional_test.go` -- add `TestServicoDetalhe` (GET /servicos/expansao-de-redes, 200, verify title and methodology content), `TestServicoDetalheNotFound` (GET /servicos/inexistente, 404), `TestServicoDetalheAll` (loop over all 5 slugs, verify 200 for each). Update `TestServicos` assertions for new index content (no "Relatorios PDF", no "Gestao de Pipeline"). (R12, SC-004)

**Checkpoint**: Servicos index + 5 deep pages functional. No software copy.

---

## Phase 5: User Story 3 - Quem Somos (Priority: P3)

**Goal**: Founder story with market authority, mission/vision/values, CRECI.

**Independent Test**: Open `/quem-somos` -- verify Luiz Claudio, Shell, mission/vision/values, CRECI. No "plataforma" or "carga cognitiva".

### Implementation for User Story 3

- [ ] T018 [US3] Rewrite `internal/template/quem-somos.html` -- per Pencil frame "Quem Somos - Desktop". Structure: (1) Title section "Quem Somos". (2) Company description: "Planejamos, estruturamos, prospectamos e comercializamos" (legacy). (3) Founder section: "Luiz Claudio P. André", "15 anos Shell Brasil", "especialista em redes de franquias e varejo", "lojas de rua, calcadao de bairros, shopping centers" (legacy). Founder photo or initials "LC" in circle. (4) Mission block: "Oferecer a melhor qualidade de servicos de prospecao de pontos comerciais" (legacy). (5) Vision block: "Ser reconhecida como a melhor empresa de prospecao de pontos comerciais do Rio de Janeiro" (legacy). (6) Values block: Transparencia, Profissionalismo, Etica, Comprometimento, Agilidade (legacy). (7) CRECI mention: "licenciados pelo Conselho Federal de Corretores de Imoveis". NO "carga cognitiva", "plataforma", "software", "reduzir a carga". (FR-009, FR-010, FR-011, R5)
- [ ] T019 [US3] Update `internal/handler/institutional_test.go` -- update `TestQuemSomos` assertions: verify "Luiz Claudio" in body, verify "Shell" in body, verify "Missao" in body, verify "CRECI" in body, verify NO "carga cognitiva" in body, verify NO "plataforma" in body. (R12, SC-004)

**Checkpoint**: Quem Somos with founder authority. No software copy.

---

## Phase 6: User Story 4 - Nossos Clientes (Priority: P4)

**Goal**: Testimonials + metrics, no empty state.

**Independent Test**: Open `/nossos-clientes` -- verify 2+ testimonials, metrics, no "Em breve".

### Implementation for User Story 4

- [ ] T020 [US4] Rewrite `internal/template/nossos-clientes.html` -- per Pencil frame "Nossos Clientes - Desktop". Structure: (1) Title "Nossos Clientes". (2) Metrics strip: `{{template "metrics" .Metrics}}`. (3) Testimonials grid: `{{range .Testimonials}}` with quote, name, company. At least 2 visible (Larissa Mello, Roberto Andrade, Joao Viana). (4) NO empty state section -- remove the "Em breve" block entirely. If no testimonials, the section is omitted (but we always have 3 static). (5) CTA at bottom: "Solicite uma apresentacao" or "Fale Conosco". (FR-012, FR-013, FR-014, R4)
- [ ] T021 [US4] Update `internal/handler/institutional_test.go` -- update `TestNossosClientes` assertions: verify 200, verify "Larissa Mello" OR "Roberto Andrade" in body, verify NO "Em breve" in body, verify metrics labels present. (R12)

**Checkpoint**: Nossos Clientes with social proof. No empty state.

---

## Phase 7: User Story 5 - Fale Conosco (Priority: P5)

**Goal**: Visual polish, contact info in page, company field. Keep HTMX behavior.

**Independent Test**: Open `/fale-conosco` -- verify form with Empresa field, contact info visible, submit works (HTMX + no-JS).

### Implementation for User Story 5

- [ ] T022 [P] [US5] Rewrite `internal/template/fale-conosco.html` -- per Pencil frame "Fale Conosco - Desktop". Structure: (1) Title "Fale Conosco". (2) Two-column layout (desktop): left = form, right = contact info. Mobile: stacked. (3) Form fields: Empresa (optional, new), Nome, Email, Telefone, Assunto, Mensagem. Keep `hx-post="/fale-conosco" hx-target="#contact-form-container" hx-swap="outerHTML"` and `action="/fale-conosco" method="POST"` for no-JS fallback. (4) Contact info column: address (Praia de Botafogo, 228 - 16 Andar - Edificio Argentina, Botafogo - RJ), phones (+55 21 99842-3232 / 97034-2617 / 3736-3696), email, WhatsApp link. (5) Success/error display unchanged (HTMX fragments). (FR-015, FR-016, R10)
- [ ] T023 [P] [US5] Update `internal/template/fragments/contact_success.html` and `contact_error.html` -- if they reference form fields, add `Company` to the repopulated form on error. Verify the fragments still work with HTMX swap. (R10)
- [ ] T024 [US5] Update `internal/handler/institutional_test.go` -- update `TestFaleConosco` assertions: verify 200, verify "Empresa" field in body, verify "Botafogo" or phone in body (contact info on page). Update `TestContactSubmitValid` to include `company` form value. Verify success still works. (R12)

**Checkpoint**: Fale Conosco with visual polish and contact info. HTMX behavior preserved.

---

## Phase 8: User Story 6 - Auth Templates CSS (Priority: P6)

**Goal**: SKIP -- already done. Verify only.

**Note**: The user has already deployed CSS fixes to login.html, totp_setup.html, totp_verify.html. These templates already include `<link rel="stylesheet" href="/static/css/app.css">` and use design system classes. Do NOT redo from scratch.

- [ ] T025 [US6] Verify auth templates have CSS: run `grep "app.css" internal/template/login.html internal/template/totp_setup.html internal/template/totp_verify.html` -- all 3 must match. Run `grep 'style="color' internal/template/login.html internal/template/totp_setup.html internal/template/totp_verify.html` -- must return empty (no inline styles). If any template lacks CSS or has inline styles, fix it. Otherwise, mark as done. (FR-017, FR-018, R6)
- [ ] T026 [US6] Add auth CSS tests to `internal/handler/auth_handler_test.go` (if not already present): `TestLoginHasCSS` (GET /login, verify `<link rel="stylesheet" href="/static/css/app.css">` in body), `TestTotpSetupHasCSS` (GET /2fa/setup, same check), `TestTotpVerifyHasCSS` (GET /2fa/verify, same check). These are lightweight -- just verify the CSS link is present. (FR-017, R12)

**Checkpoint**: Auth templates verified with CSS. No inline styles.

---

## Phase 9: Polish (Nav, Footer, Final Verification)

**Purpose**: Update nav/footer, run make check, commit, push, verify CI.

- [ ] T027 [P] Rewrite `internal/template/partials/nav.html` -- keep structure (logo + links + CTA + mobile hamburger). Update copy if needed: "Início" or "Home" (either is fine). Links: /, /quem-somos, /servicos, /nossos-clientes, /fale-conosco (CTA button). Active state via `.ActivePage`. No software copy. (FR-028, R8)
- [ ] T028 [P] Rewrite `internal/template/partials/footer.html` -- keep structure (3-col grid + newsletter). Update content: (1) Brand column: "Prospecção Brasil" + "Prospecção imobiliária comercial" tagline. (2) Contact column: address (Praia de Botafogo, 228 - 16 Andar, Botafogo - RJ), phones (+55 21 99842-3232 / 97034-2617 / 3736-3696), email, WhatsApp link, Instagram link. (3) Newsletter column: keep existing form (HTMX, unchanged). (4) Bottom: copyright + CRECI mention. (FR-028, FR-029, R8)
- [ ] T029 Run `make build-css` to rebuild Tailwind CSS with any new classes from the redesign. Verify `static/css/app.css` is updated. (SC-006)
- [ ] T030 Run forbidden copy check: `grep -ri "carga cognitiva\|pipeline\|plataforma\|software\|relatorios pdf\|gestao de pipeline" internal/template/home.html internal/template/servicos.html internal/template/servico-detalhe.html internal/template/quem-somos.html internal/template/nossos-clientes.html internal/template/fale-conosco.html internal/template/partials/nav.html internal/template/partials/footer.html` -- must return empty. (SC-004, FR-025)
- [ ] T031 Run `make check` with clean test database: `psql -d prospeccaobrasil_test -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; CREATE EXTENSION IF NOT EXISTS pgcrypto;"` then `make check`. Verify: golangci-lint 0 issues, all tests pass, coverage >= 70% (app) / 85% (auth), build succeeds, ast-grep clean. Fix any failures. (SC-006)
- [ ] T032 Verify ast-grep rules pass: `ast-grep scan` -- 0 errors. Rules: `go-handler-missing-auth.yml` (public handlers intentionally no auth -- verify no false positives), `go-missing-tenant-filter.yml` (public queries intentionally no tenant -- verify no false positives). (Constitution II, IV)
- [ ] T033 Commit all SPEC-06 changes. Push to main. Verify CI passes via `gh run watch`. Check that the Test step runs integration tests against Postgres service container and the coverage gate passes (70% app / 85% auth). (SC-006, SC-007)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- can start immediately. T001-T005 can mostly run in parallel. T006 (Pencil) is independent of code.
- **Foundational (Phase 2)**: Depends on Setup (T001-T005). T007-T008 define data structures. T009 depends on T007-T008. T010 depends on T008 (metrics struct).
- **User Stories (Phase 3-8)**: All depend on Foundational phase (T007-T010) completion.
  - US1 (Home) can start after Foundational -- P1, highest priority
  - US2 (Servicos) can start after Foundational -- P2, independent of US1
  - US3 (Quem Somos) can start after Foundational -- P3, independent of US1, US2
  - US4 (Nossos Clientes) can start after Foundational -- P4, independent
  - US5 (Fale Conosco) depends on T001-T003 (company field migration + handler) -- P5
  - US6 (Auth CSS) is SKIP/verify only -- P6, can run anytime
- **Polish (Phase 9)**: Depends on all user stories being complete. T027-T028 can run in parallel with user stories (nav/footer are shared).

### User Story Dependencies

- **US1 (Home)**: Depends on Foundational (services map, testimonials, metrics, metrics partial)
- **US2 (Servicos)**: Depends on Foundational (services map). T013-T014 (handler + router) must complete before T015-T016 (templates).
- **US3 (Quem Somos)**: No dependencies on other stories. Content is static in template.
- **US4 (Nossos Clientes)**: Depends on Foundational (testimonials, metrics, metrics partial)
- **US5 (Fale Conosco)**: Depends on T001-T003 (company field). Form behavior unchanged.
- **US6 (Auth CSS)**: Already done. Verify only.

### Parallel Opportunities

- T004, T005 can run in parallel with T001-T003 (different files)
- T006 (Pencil) can run in parallel with all Setup tasks
- T010 (metrics partial) can run in parallel with T007-T008 (different files, but T010 needs the metric struct from T008 -- so T008 first)
- T015, T016 can run in parallel (different template files)
- T022, T023 can run in parallel (different template files)
- T027, T028 can run in parallel (nav and footer are separate files)

---

## Test Coverage Notes

- **internal/handler**: 70% minimum (app code gate). Existing SPEC-04 tests cover institutional pages rendering. Update assertions for new copy. Add tests for `/servicos/{slug}` (T017). Add auth CSS tests (T026).
- **internal/auth**: 85% minimum (security-critical gate). No changes to auth code in SPEC-06 (templates only). Existing coverage maintained.
- **internal/db**: Excluded (sqlc generated). Migration adds column, sqlc regenerated.
- **cmd/prospeccao**: Excluded. Only router change (add `/servicos/{slug}` route).
- **scripts/**: Excluded.

## Pencil Design Notes

- Pencil designs (T006) are created using the Pencil MCP server (`mcp_call_tool` with `server_name: "pencil"`).
- Designs use brand tokens: Deep Navy `#031636` primary, Sóbrio Gold `#765a1a` secondary, Montserrat display, Inter body.
- 8 frames required (see T006 for list). "Login - Desktop" is reference only (already implemented).
- Designs are the visual source of truth. HTML implementation must match Pencil frames.
- Pencil file: `designs/prospeccao.pen`
