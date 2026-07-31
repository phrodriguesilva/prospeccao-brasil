# Tasks: Internal System (Properties, Clients, Prospecting CRUD + PDF)

**Input**: Design documents from `/specs/005-internal-system/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are included (85% coverage gate is mandatory per constitution).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add chromedp dependency, new sqlc queries, and internal layout template.

- [ ] T001 Add chromedp dependency: `go get github.com/chromedp/chromedp`. Verify `go.mod` has the new dependency. (FR-036, research.md R8)
- [ ] T002 Add new sqlc queries to `internal/db/queries/dashboard.sql`: `CountPropertiesByTenant`, `CountClientsByTenant`, `CountProspectsByTenant`, `CountProspectsByStatus`, `ListRecentProspectsWithDetails` (JOIN with clients + properties for names). Run `make sqlc` to generate typed Go code. (FR-001, FR-002, FR-003, data-model.md)
- [ ] T003 Add filtered/paginated sqlc queries to `internal/db/queries/properties.sql`: `ListPropertiesFiltered` (with status, type, search, limit, offset params), `CountPropertiesFiltered`. Run `make sqlc`. (FR-005, FR-006, research.md R2, R3)
- [ ] T004 [P] Add filtered/paginated sqlc queries to `internal/db/queries/clients.sql`: `ListClientsFiltered` (with status, search, limit, offset), `CountClientsFiltered`. Run `make sqlc`. (FR-013, FR-014, research.md R2, R3)
- [ ] T005 [P] Add filtered/paginated sqlc queries to `internal/db/queries/prospections.sql`: `ListProspectsFiltered` (with status, limit, offset), `CountProspectsFiltered`, `ListProspectsByClientWithDetails` (JOIN client + property names). Run `make sqlc`. (FR-021, FR-022, research.md R2, R3)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Internal system layout, sidebar nav, and router wiring. MUST be complete before any user story.

**CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T006 Create `internal/template/admin/_layout.html` -- internal system base layout with: `<head>` (meta, title block, app.css, htmx.min.js, alpine.min.js), sidebar navigation (Dashboard, Imóveis, Clientes, Prospecções links with active state via current path, user info at bottom with logout button, mobile hamburger via Alpine.js `x-data`/`x-show`), `{{template "content" .}}` block, no footer (internal system has no marketing footer). (FR-039 to FR-041, research.md R4)
- [ ] T007 Create `internal/template/partials/internal_nav.html` -- sidebar navigation partial included by `_layout.html` via `{{template "internal_nav" .}}`. Links: `/admin` (Dashboard), `/properties` (Imóveis), `/clients` (Clientes), `/prospections` (Prospecções). Active state based on request path. User email + role shown at bottom with logout form (POST to `/logout`). (research.md R4)
- [ ] T008 Update `cmd/prospeccao/main.go` -- extend `buildInternalRouter` with CRUD routes: `GET /admin` (dashboard), `GET/POST /properties`, `GET /properties/new`, `GET/POST /properties/{id}`, `GET /properties/{id}/edit`, `POST /properties/{id}/delete`, same pattern for `/clients` and `/prospections`, `POST /clients/{id}/contacts`, `POST /prospections/{id}/contacts`, `GET /properties/{id}/pdf`. All protected by `SessionValidation` + `RequireRole(admin)`. Update `loadTemplates` to parse `internal/template/admin/**/*.html`. (FR-039 to FR-044, contracts/endpoints.md)

**Checkpoint**: Foundation ready -- internal layout, sidebar, router wired. User story implementation can begin.

---

## Phase 3: User Story 1 - Admin Dashboard (Priority: P1) MVP

**Goal**: Dashboard with counts, prospections by status, and recent prospections.

**Independent Test**: Login and visit `/admin` -- verify counts (0 for fresh tenant), empty state with CTA.

### Implementation for User Story 1

- [ ] T009 [US1] Create `internal/handler/dashboard.go` -- `DashboardHandler` struct with `queries`, `tmpl`, `log`. `Index` method: gets tenant_id from session context, calls `CountPropertiesByTenant`, `CountClientsByTenant`, `CountProspectsByTenant`, `CountProspectsByStatus`, `ListRecentProspectsWithDetails`, renders `dashboard.html` with counts, status breakdown, recent prospections, and empty state flag. (FR-001 to FR-004, contracts/endpoints.md)
- [ ] T010 [P] [US1] Create `internal/template/admin/dashboard.html` -- uses `_layout.html`. Shows: 4 stat cards (Imóveis, Clientes, Prospecções, Taxa de Sucesso), prospections by status (badges with counts), recent prospections table (client name, property title, status badge, next action date), empty state with CTA "Cadastre seu primeiro imóvel" when all counts are 0. (FR-001 to FR-004)
- [ ] T011 [US1] Create `internal/handler/dashboard_test.go` -- integration tests: `TestDashboardEmpty` (0 counts, empty state CTA), `TestDashboardWithSeed` (seed 3 properties, 2 clients, 5 prospections, verify counts and recent prospections), `TestDashboardAuthRequired` (302 redirect to /login without session), `TestDashboardStatusBreakdown` (verify status counts match). Uses httptest + real Postgres + authenticated session. (SC-001, SC-004, SC-006, quickstart.md Scenario 1)

**Checkpoint**: Dashboard is functional and tested independently.

---

## Phase 4: User Story 2 - Property CRUD (Priority: P1)

**Goal**: Full property CRUD with list, filter, search, create, edit, soft-delete.

**Independent Test**: Create a property, view it in the list, edit it, soft-delete it.

### Implementation for User Story 2

- [ ] T012 [P] [US2] Create `internal/template/admin/properties/list.html` -- paginated table with columns: title, address, price (BRL formatted), status badge, type, actions (view, edit, delete). Filter bar (status dropdown, type dropdown, search input with HTMX `hx-get` for async filter). Pagination controls (Previous/Next + page number). Empty state. (FR-005, FR-006)
- [ ] T013 [P] [US2] Create `internal/template/admin/properties/form.html` -- shared create/edit form with fields: title, address, city, state, zip_code, price, status (dropdown), type (dropdown), bedrooms, bathrooms, area_sqm, description (textarea), photos (textarea with one URL per line, converted to JSON array on submit). Form posts to create or edit endpoint based on `{{.IsEdit}}` flag. Field-level error display. (FR-007, FR-008)
- [ ] T014 [P] [US2] Create `internal/template/admin/properties/detail.html` -- property detail page showing all fields, photos (grid), linked prospections (table with client name, status, next action), action buttons (Editar, Excluir, Gerar PDF). Soft-delete confirmation modal (Alpine.js). (FR-009, FR-034)
- [ ] T015 [US2] Create `internal/handler/property.go` -- `PropertyHandler` struct with `queries`, `tmpl`, `log`. Methods: `List` (paginated, filtered, parses query params), `New` (render form), `Create` (validate, persist, redirect), `Detail` (fetch by ID + tenant, 404 if not found), `Edit` (render pre-filled form), `Update` (validate, persist, redirect), `Delete` (soft-delete, redirect). All use tenant_id from session context. Validation: title min 3, address min 5, city min 2, state min 2, price > 0. (FR-005 to FR-012, contracts/endpoints.md)
- [ ] T016 [US2] Create `internal/handler/property_test.go` -- integration tests: `TestPropertyList` (200, paginated), `TestPropertyListFiltered` (status + type filter), `TestPropertyListSearch` (title/city search), `TestPropertyCreate` (valid form, redirect, DB record), `TestPropertyCreateInvalid` (short title, invalid price), `TestPropertyDetail` (200, all fields), `TestPropertyDetailNotFound` (404 for wrong tenant), `TestPropertyEdit` (pre-filled form), `TestPropertyUpdate` (change price, verify), `TestPropertySoftDelete` (deleted_at set, not in list), `TestPropertyAuthRequired` (302 without session). (SC-002, SC-005, SC-006, quickstart.md Scenarios 2-5, 12, 13)

**Checkpoint**: Property CRUD is functional and tested independently.

---

## Phase 5: User Story 3 - Client CRUD (Priority: P1)

**Goal**: Full client CRUD with list, filter, search, create, edit, soft-delete.

**Independent Test**: Create a client, view it in the list, edit it, soft-delete it.

### Implementation for User Story 3

- [ ] T017 [P] [US3] Create `internal/template/admin/clients/list.html` -- paginated table with columns: name, email, phone, status badge, budget (BRL), actions. Filter bar (status, search). Pagination. Empty state. (FR-013, FR-014)
- [ ] T018 [P] [US3] Create `internal/template/admin/clients/form.html` -- shared create/edit form with fields: name, email, phone, cpf_cnpj, address, budget, status (dropdown), preferences (JSON textarea, optional). Field-level error display. (FR-015, FR-016)
- [ ] T019 [P] [US3] Create `internal/template/admin/clients/detail.html` -- client detail showing all fields, linked prospections (table), contact log (chronological list with channel, direction, subject, body, timestamp), action buttons (Editar, Excluir, Registrar Contato). Soft-delete modal. (FR-017)
- [ ] T020 [US3] Create `internal/handler/client.go` -- `ClientHandler` struct. Methods: `List`, `New`, `Create`, `Detail`, `Edit`, `Update`, `Delete`. Same pattern as PropertyHandler. Validation: name min 2, email valid (net/mail.ParseAddress), budget >= 0. All queries include tenant_id. (FR-013 to FR-020, contracts/endpoints.md)
- [ ] T021 [US3] Create `internal/handler/client_test.go` -- integration tests: `TestClientList`, `TestClientListFiltered`, `TestClientListSearch`, `TestClientCreate`, `TestClientCreateInvalid` (short name, invalid email), `TestClientDetail`, `TestClientDetailNotFound`, `TestClientEdit`, `TestClientUpdate`, `TestClientSoftDelete`, `TestClientAuthRequired`. (SC-005, SC-006, quickstart.md Scenario 6)

**Checkpoint**: Client CRUD is functional and tested independently.

---

## Phase 6: User Story 4 - Prospecting CRUD (Priority: P1)

**Goal**: Full prospection CRUD linking clients to properties with status tracking.

**Independent Test**: Create a prospection (selecting client + property), view it, update status, soft-delete it.

### Implementation for User Story 4

- [ ] T022 [P] [US4] Create `internal/template/admin/prospections/list.html` -- paginated table with columns: client name, property title, status badge, next action date, actions. Filter bar (status). Pagination. Empty state with message "Nenhuma prospecção cadastrada" and CTA. (FR-021, FR-022)
- [ ] T023 [P] [US4] Create `internal/template/admin/prospections/form.html` -- shared create/edit form with: client dropdown (populated from ListClientsByTenant), property dropdown (populated from ListPropertiesByTenant), status dropdown (6 options), notes (textarea), contact_date (date input), next_action_date (date input). If no clients or properties exist, show message "Cadastre um cliente e um imóvel primeiro." (FR-023, FR-024)
- [ ] T024 [P] [US4] Create `internal/template/admin/prospections/detail.html` -- prospection detail showing: linked client info (name, email, phone -- clickable to client detail), linked property info (title, address, price -- clickable to property detail), status badge, notes, dates, contact log for this prospection, action buttons (Editar, Excluir, Registrar Contato). (FR-025)
- [ ] T025 [US4] Create `internal/handler/prospection.go` -- `ProspectionHandler` struct. Methods: `List`, `New`, `Create`, `Detail`, `Edit`, `Update`, `Delete`. New/Create: fetch client + property lists for dropdowns. Validation: client_id required (must exist in tenant), property_id required (must exist in tenant), status enum. All queries include tenant_id. (FR-021 to FR-028, contracts/endpoints.md)
- [ ] T026 [US4] Create `internal/handler/prospection_test.go` -- integration tests: `TestProspectionList`, `TestProspectionListFiltered`, `TestProspectionCreate` (valid, with client + property), `TestProspectionCreateNoClientsOrProperties` (shows message), `TestProspectionDetail`, `TestProspectionDetailNotFound`, `TestProspectionUpdateStatus`, `TestProspectionSoftDelete`, `TestProspectionAuthRequired`, `TestProspectionCrossTenant` (404 for wrong tenant). (SC-005, SC-006, quickstart.md Scenarios 7-8, 13)

**Checkpoint**: Prospecting CRUD is functional and tested independently.

---

## Phase 7: User Story 5 - Contact Log (Priority: P2)

**Goal**: Immutable contact log for clients and prospections with HTMX async creation.

**Independent Test**: Create a contact log entry for a client, verify it appears in the log, verify no edit/delete buttons.

### Implementation for User Story 5

- [ ] T027 [P] [US5] Create `internal/template/admin/contacts/_form.html` -- HTMX inline form fragment with fields: channel (dropdown: phone, email, whatsapp, in_person), direction (dropdown: inbound, outbound), subject (text, optional), body (textarea, optional), contacted_at (date, default today). Form uses `hx-post` to submit and `hx-target` to swap the contact log container. (FR-029, FR-030, research.md R5)
- [ ] T028 [P] [US5] Create `internal/template/admin/contacts/_log.html` -- contact log fragment showing chronological list of contacts with: channel icon, direction arrow, subject, body, timestamp (formatted as "dd/mm/yyyy HH:MM"). No edit or delete buttons (immutable). Empty state: "Nenhum contato registrado." (FR-031, FR-032)
- [ ] T029 [US5] Add contact creation methods to `internal/handler/client.go` and `internal/handler/prospection.go` -- `CreateContact` method on both handlers. Parses form, validates (channel enum, direction enum), persists via `queries.CreateContact` with tenant_id, client_id (from URL path), prospect_id (from URL path or null). HTMX: returns `_log.html` fragment with updated log. Non-HTMX: redirects to detail page. (FR-029 to FR-033, contracts/endpoints.md)
- [ ] T030 [US5] Create `internal/handler/contact_log_test.go` -- integration tests: `TestCreateContactForClient` (HTMX, 200, fragment, DB record), `TestCreateContactForProspection` (HTMX, 200, linked to prospect + client), `TestCreateContactInvalidChannel` (error), `TestCreateContactNoJS` (303 redirect), `TestContactLogImmutable` (no edit/delete in response), `TestContactLogDisplayed` (appears on client detail), `TestContactAuthRequired`. (SC-007, quickstart.md Scenarios 9-10)

**Checkpoint**: Contact log is functional, immutable, and tested.

---

## Phase 8: User Story 6 - PDF Generation (Priority: P2)

**Goal**: Generate professional PDF presentation for a property using chromedp.

**Independent Test**: Generate a PDF for a property with photos, verify file is valid PDF.

### Implementation for User Story 6

- [ ] T031 [P] [US6] Create `internal/template/admin/properties/pdf.html` -- standalone HTML template (no layout) for PDF rendering. Includes: property title (large), address, price (BRL formatted), type, area, bedrooms, bathrooms, description, photos (full-width images). Styled with inline CSS (chromedp needs self-contained HTML, no external CSS file). Print-friendly layout (A4, margins, page breaks between sections). (FR-035, research.md R1)
- [ ] T032 [US6] Create `internal/handler/pdf.go` -- `PDFHandler` struct with `queries`, `tmpl`, `log`. `GeneratePropertyPDF` method: fetch property by ID + tenant_id (404 if not found), render `pdf.html` template to a temp HTML file, use chromedp to navigate to `file://` URL and call `page.PrintToPDF`, return PDF as download (Content-Type: application/pdf, Content-Disposition: attachment). Timeout: 30 seconds. If chromedp/Chrome not available, return 500 with user-friendly error. (FR-034 to FR-038, research.md R1, R8)
- [ ] T033 [US6] Create `internal/handler/pdf_test.go` -- integration tests: `TestGeneratePDF` (skip if Chrome not available via `exec.LookPath("chromium-browser")` or `exec.LookPath("google-chrome")`, 200, Content-Type application/pdf, non-empty body, contains PDF magic bytes `%PDF`), `TestGeneratePDFNotFound` (404 for wrong tenant), `TestGeneratePDFAuthRequired` (302 without session), `TestGeneratePDFNoChrome` (mock/verify error handling when Chrome unavailable). (SC-003, quickstart.md Scenario 11)

**Checkpoint**: PDF generation is functional and tested.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final verification, CI, and cleanup.

- [ ] T034 [P] Update `internal/template/admin/_layout.html` -- ensure all design system classes (btn, badge, card, form-input) from SPEC-04 are used consistently. Add internal-specific styles (sidebar, stat cards, data tables with zebra striping). Run `make build-css`. (FR-001 to FR-007 from SPEC-04)
- [ ] T035 Run `make check` and verify: golangci-lint 0 issues, all tests pass, coverage >= 85% (excluding internal/db, cmd/prospeccao, scripts), build succeeds, ast-grep 0 errors. Fix any failures. (SC-005 to SC-008, quickstart.md Scenario 16)
- [ ] T036 Run all 16 quickstart.md validation scenarios and verify each passes. Document any failures and fix. (All FRs, quickstart.md)
- [ ] T037 Verify ast-grep rules pass: `go-handler-missing-auth.yml` (all internal handlers have SessionValidation), `go-missing-tenant-filter.yml` (all DB queries include tenant_id). Fix any violations. (Constitution II, IV)
- [ ] T038 Commit all changes. Push to main. Verify CI passes via `gh run watch`. Check that the Test step runs integration tests against Postgres service container and the coverage gate passes. (SC-008)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- can start immediately
- **Foundational (Phase 2)**: Depends on Setup (T001-T005) -- BLOCKS all user stories
- **User Stories (Phase 3-8)**: All depend on Foundational phase (T006-T008) completion
  - US1 (Dashboard) can start after Foundational
  - US2 (Property CRUD) can start after Foundational -- independent of US1
  - US3 (Client CRUD) can start after Foundational -- independent of US1, US2
  - US4 (Prospecting CRUD) depends on US2 + US3 (needs properties + clients to exist for dropdowns)
  - US5 (Contact Log) depends on US3 + US4 (needs clients + prospections to attach contacts)
  - US6 (PDF) depends on US2 (needs property detail page with "Gerar PDF" button)
- **Polish (Phase 9)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (Dashboard)**: No dependencies on other stories
- **US2 (Property CRUD)**: No dependencies on other stories
- **US3 (Client CRUD)**: No dependencies on other stories
- **US4 (Prospecting CRUD)**: Depends on US2 + US3 (dropdowns need data)
- **US5 (Contact Log)**: Depends on US3 + US4 (contacts attach to clients/prospections)
- **US6 (PDF)**: Depends on US2 (property detail page must exist)

### Parallel Opportunities

- T003, T004, T005 can run in parallel (different sqlc query files)
- T012, T013, T014 can run in parallel (different property templates)
- T017, T018, T019 can run in parallel (different client templates)
- T022, T023, T024 can run in parallel (different prospection templates)
- T027, T028 can run in parallel (different contact fragments)
- T031 can run in parallel with T032 (template vs handler)
- US2, US3 can be developed in parallel (independent entities)

---

## Implementation Strategy

### MVP First (Dashboard + Property CRUD)

1. Complete Phase 1: Setup (sqlc queries + chromedp)
2. Complete Phase 2: Foundational (layout + router)
3. Complete Phase 3: US1 Dashboard
4. Complete Phase 4: US2 Property CRUD
5. **STOP and VALIDATE**: Test dashboard + property CRUD independently

### Incremental Delivery

1. Setup + Foundational -> Foundation ready
2. Add US1 Dashboard -> Test -> Deploy
3. Add US2 Property CRUD -> Test -> Deploy
4. Add US3 Client CRUD -> Test -> Deploy
5. Add US4 Prospecting CRUD -> Test -> Deploy
6. Add US5 Contact Log -> Test -> Deploy
7. Add US6 PDF Generation -> Test -> Deploy
8. Polish -> Final CI verification -> Deploy
