# Feature Specification: Internal System (Properties, Clients, Prospecting CRUD + PDF)

**Feature Branch**: `005-internal-system`

**Created**: 2026-07-31

**Status**: Draft

**Input**: User description: "Internal system with property CRUD, client CRUD, prospecting CRUD (link client + property, status tracking), contact log, PDF generation via chromedp, admin dashboard. All internal pages require auth and tenant_id filtering. Host-based routing: sistema.* = internal, prospeccaobrasil.com = public."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Admin Dashboard (Priority: P1)

As the admin (Luiz Claudio), I want to log into the internal system and see a dashboard with an overview of my business: total properties, total clients, total prospections (by status), and recent activity (latest prospections and contacts). This gives me a quick pulse of my pipeline without clicking through multiple pages.

**Why this priority**: The dashboard is the landing page after login. It must exist before any CRUD because it is the navigation hub. Without it, the user has nowhere to land after authenticating.

**Independent Test**: Can be fully tested by logging in and verifying the dashboard shows counts (0 for a fresh tenant) and recent activity. Delivers immediate value as a navigation hub.

**Acceptance Scenarios**:

1. **Given** the user is authenticated and has 5 properties, 3 clients, and 2 prospections, **When** they visit `/admin` (the dashboard), **Then** they see "5 Imóveis", "3 Clientes", "2 Prospecções" and a breakdown of prospections by status (e.g., "1 Novo", "1 Em contato")
2. **Given** the user is not authenticated, **When** they visit `/admin`, **Then** they are redirected to `/login`
3. **Given** the user has 0 properties/clients/prospections, **When** they visit the dashboard, **Then** they see "0" counts and an empty-state message encouraging them to add their first property
4. **Given** the user has recent prospections, **When** they view the dashboard, **Then** they see the 5 most recent prospections with client name, property title, status badge, and next action date

---

### User Story 2 - Property CRUD (Priority: P1)

As the admin, I want to create, view, edit, and soft-delete properties (imóveis). Each property has a title, address, city, state, zip code, price, status (available/reserved/sold/inactive), type (residential/commercial/land/rural), bedrooms, bathrooms, area, description, and photos. I need to list properties with filters (status, type, search by title/city) and paginate the results.

**Why this priority**: Properties are the core entity. Without properties, there is nothing to prospect. This is the first CRUD the user needs.

**Independent Test**: Can be fully tested by creating a property, viewing it in the list, editing it, and soft-deleting it. Delivers value as a property catalog.

**Acceptance Scenarios**:

1. **Given** the user is authenticated, **When** they visit `/properties`, **Then** they see a paginated table of properties with columns: title, address, price, status badge, type, and action buttons (view, edit, delete)
2. **Given** the user is on the properties list, **When** they filter by status="available" and type="commercial", **Then** only available commercial properties are shown
3. **Given** the user is on the properties list, **When** they search "São Paulo" in the search box, **Then** properties with title or city matching "São Paulo" are shown
4. **Given** the user clicks "Novo Imóvel", **When** they fill the form with valid data and submit, **Then** the property is created and they are redirected to the property detail page
5. **Given** the user is on a property detail page, **When** they click "Editar", modify the price, and submit, **Then** the property is updated and the new price is shown
6. **Given** the user is on a property detail page, **When** they click "Excluir" and confirm, **Then** the property is soft-deleted (deleted_at set) and no longer appears in the list
7. **Given** the user is not authenticated, **When** they visit `/properties`, **Then** they are redirected to `/login`

---

### User Story 3 - Client CRUD (Priority: P1)

As the admin, I want to create, view, edit, and soft-delete clients. Each client has a name, email, phone, CPF/CNPJ, address, budget, preferences (JSON), and status (active/inactive/lead). I need to list clients with filters (status, search by name/email) and paginate the results.

**Why this priority**: Clients are the second core entity. Prospections link a client to a property, so clients must exist before prospections.

**Independent Test**: Can be fully tested by creating a client, viewing it in the list, editing it, and soft-deleting it. Delivers value as a client catalog.

**Acceptance Scenarios**:

1. **Given** the user is authenticated, **When** they visit `/clients`, **Then** they see a paginated table of clients with columns: name, email, phone, status badge, budget, and action buttons (view, edit, delete)
2. **Given** the user is on the clients list, **When** they filter by status="lead", **Then** only lead clients are shown
3. **Given** the user is on the clients list, **When** they search "João" in the search box, **Then** clients with name or email matching "João" are shown
4. **Given** the user clicks "Novo Cliente", **When** they fill the form with valid data and submit, **Then** the client is created and they are redirected to the client detail page
5. **Given** the user is on a client detail page, **When** they click "Editar", modify the phone, and submit, **Then** the client is updated and the new phone is shown
6. **Given** the user is on a client detail page, **When** they click "Excluir" and confirm, **Then** the client is soft-deleted and no longer appears in the list
7. **Given** the user views a client detail page, **Then** they see the client's prospections and contact log (recent interactions)

---

### User Story 4 - Prospecting CRUD (Priority: P1)

As the admin, I want to create, view, edit, and soft-delete prospections. A prospection links a client to a property with a status (new/contacting/visiting/negotiating/closed_won/closed_lost), notes, contact date, and next action date. I need to list prospections with filters (status, client, property) and see the pipeline as a Kanban-style board or filtered list.

**Why this priority**: Prospections are the pipeline. This is the core business value -- tracking which client is interested in which property and at what stage.

**Independent Test**: Can be fully tested by creating a prospection (selecting a client and property), viewing it, updating its status, and soft-deleting it. Delivers value as a pipeline tracker.

**Acceptance Scenarios**:

1. **Given** the user is authenticated and has at least 1 client and 1 property, **When** they visit `/prospections`, **Then** they see a list of prospections with columns: client name, property title, status badge, next action date, and action buttons
2. **Given** the user clicks "Nova Prospecção", **When** they select a client and property, set status to "new", and submit, **Then** the prospection is created
3. **Given** the user is on a prospection detail page, **When** they change the status from "new" to "contacting" and submit, **Then** the status is updated and the status badge reflects the change
4. **Given** the user is on the prospections list, **When** they filter by status="negotiating", **Then** only negotiating prospections are shown
5. **Given** the user is on a prospection detail page, **When** they set a next action date, **Then** the date is saved and visible
6. **Given** the user views a prospection detail page, **Then** they see the linked client info, linked property info, status timeline, and contact log for this prospection

---

### User Story 5 - Contact Log (Priority: P2)

As the admin, I want to log interactions (contacts) with clients. Each contact has a channel (phone/email/whatsapp/in_person), direction (inbound/outbound), subject, body, and contacted_at timestamp. Contacts are immutable (no edit/delete) and can be linked to a specific prospection or be a standalone client interaction.

**Why this priority**: The contact log is the history of interactions. It is P2 because the pipeline (US4) works without it, but it adds significant value for follow-up tracking.

**Independent Test**: Can be fully tested by creating a contact log entry for a client, viewing it in the client's contact history, and verifying it is immutable (no edit/delete buttons).

**Acceptance Scenarios**:

1. **Given** the user is on a client detail page, **When** they click "Registrar Contato", fill the form (channel, direction, subject, body), and submit, **Then** the contact is created and appears in the client's contact log
2. **Given** the user is on a prospection detail page, **When** they click "Registrar Contato", fill the form, and submit, **Then** the contact is created and linked to both the client and the prospection
3. **Given** the user views a client's contact log, **Then** they see a chronological list of contacts with channel icon, direction, subject, body, and timestamp
4. **Given** a contact has been created, **When** the user views it, **Then** there are no edit or delete buttons (contacts are immutable)

---

### User Story 6 - PDF Report Generation (Priority: P2)

As the admin, I want to generate a professional PDF presentation document for a property. The PDF includes the property details (title, address, price, type, area, bedrooms, bathrooms, description) and photos. This is used to present the property to clients.

**Why this priority**: PDF generation is a key differentiator but requires chromedp (headless Chrome) which adds complexity. It is P2 because the CRUD (US2-US4) works without it, but it is a core value proposition of the platform.

**Independent Test**: Can be fully tested by creating a property with photos, generating the PDF, and verifying the PDF file is valid (non-empty, correct content-type, contains property title).

**Acceptance Scenarios**:

1. **Given** the user is on a property detail page, **When** they click "Gerar PDF", **Then** a PDF is generated and downloaded with the property details
2. **Given** the property has photos, **When** the PDF is generated, **Then** the photos are included in the PDF
3. **Given** the property has no photos, **When** the PDF is generated, **Then** a placeholder is shown where photos would be
4. **Given** the PDF generation fails (e.g., chromedp not available), **When** the user clicks "Gerar PDF", **Then** an error message is shown and the system does not crash

---

### Edge Cases

- What happens when a user tries to access a property from another tenant? The query includes `tenant_id = $2` so it returns no rows (404).
- What happens when a user tries to create a prospection without any clients or properties? The form shows a message: "Cadastre um cliente e um imóvel primeiro."
- What happens when a user soft-deletes a client that has prospections? The client is soft-deleted but prospections remain (they reference the client_id; the client just won't appear in lists). The prospection detail page shows "Cliente removido" for the client name.
- What happens when chromedp is not installed on the server? PDF generation returns a 500 error with a user-friendly message. The rest of the system works normally.
- What happens when a user navigates to `/properties?page=999` (beyond last page)? The system shows the last available page or an empty state.
- What happens when a user submits a form with invalid data (e.g., negative price)? Server-side validation rejects it and shows field-level errors.

## Requirements *(mandatory)*

### Functional Requirements

**Dashboard**

- **FR-001**: System MUST display a dashboard at `/admin` (internal system) showing total counts: properties, clients, prospections
- **FR-002**: Dashboard MUST show prospections grouped by status (new, contacting, visiting, negotiating, closed_won, closed_lost)
- **FR-003**: Dashboard MUST show the 5 most recent prospections with client name, property title, status badge, and next action date
- **FR-004**: Dashboard MUST show an empty state with a CTA when there are 0 properties/clients/prospections

**Property CRUD**

- **FR-005**: System MUST provide a paginated property list at `/properties` (20 per page) with columns: title, address, price, status badge, type, actions
- **FR-006**: Property list MUST support filtering by status, type, and text search (title or city) via query parameters
- **FR-007**: System MUST provide a property creation form at `/properties/new` with fields: title, address, city, state, zip_code, price, status, type, bedrooms, bathrooms, area_sqm, description, photos (URL list)
- **FR-008**: System MUST validate property form: title (min 3), address (min 5), city (min 2), state (min 2), price (> 0), type (enum), status (enum)
- **FR-009**: System MUST provide a property detail page at `/properties/{id}` showing all fields, photos, linked prospections, and action buttons
- **FR-010**: System MUST provide a property edit form at `/properties/{id}/edit` pre-filled with current values
- **FR-011**: System MUST soft-delete properties (set `deleted_at`) when the user clicks "Excluir" and confirms; soft-deleted properties do not appear in lists
- **FR-012**: All property queries MUST include `tenant_id` filter for multi-tenant isolation

**Client CRUD**

- **FR-013**: System MUST provide a paginated client list at `/clients` (20 per page) with columns: name, email, phone, status badge, budget, actions
- **FR-014**: Client list MUST support filtering by status and text search (name or email) via query parameters
- **FR-015**: System MUST provide a client creation form at `/clients/new` with fields: name, email, phone, cpf_cnpj, address, budget, preferences, status
- **FR-016**: System MUST validate client form: name (min 2), email (valid format), phone (optional), cpf_cnpj (optional), budget (>= 0)
- **FR-017**: System MUST provide a client detail page at `/clients/{id}` showing all fields, linked prospections, and contact log
- **FR-018**: System MUST provide a client edit form at `/clients/{id}/edit` pre-filled with current values
- **FR-019**: System MUST soft-delete clients (set `deleted_at`) when the user clicks "Excluir" and confirms
- **FR-020**: All client queries MUST include `tenant_id` filter for multi-tenant isolation

**Prospecting CRUD**

- **FR-021**: System MUST provide a paginated prospection list at `/prospections` (20 per page) with columns: client name, property title, status badge, next action date, actions
- **FR-022**: Prospection list MUST support filtering by status via query parameters
- **FR-023**: System MUST provide a prospection creation form at `/prospections/new` with a client dropdown, property dropdown, status, notes, contact_date, next_action_date
- **FR-024**: System MUST validate prospection form: client_id (required, must exist in tenant), property_id (required, must exist in tenant), status (enum)
- **FR-025**: System MUST provide a prospection detail page at `/prospections/{id}` showing linked client, linked property, status, notes, dates, and contact log
- **FR-026**: System MUST provide a prospection edit form at `/prospections/{id}/edit` allowing status, notes, and date updates
- **FR-027**: System MUST soft-delete prospections (set `deleted_at`) when the user clicks "Excluir" and confirms
- **FR-028**: All prospection queries MUST include `tenant_id` filter for multi-tenant isolation

**Contact Log**

- **FR-029**: System MUST provide a contact creation form accessible from client detail and prospection detail pages with fields: channel (enum), direction (enum), subject, body, contacted_at
- **FR-030**: System MUST validate contact form: channel (enum), direction (enum), subject (optional), body (optional), contacted_at (default now)
- **FR-031**: Contacts MUST be immutable -- no edit or delete operations are available in the UI or API
- **FR-032**: System MUST display a chronological contact log on client detail and prospection detail pages
- **FR-033**: All contact queries MUST include `tenant_id` filter for multi-tenant isolation

**PDF Generation**

- **FR-034**: System MUST provide a "Gerar PDF" button on the property detail page that generates a PDF presentation document
- **FR-035**: PDF MUST include: property title, address, price (formatted as BRL), type, area, bedrooms, bathrooms, description, and photos
- **FR-036**: PDF generation MUST use chromedp (headless Chrome) to render an HTML template and convert to PDF
- **FR-037**: System MUST return the PDF as a download (Content-Type: application/pdf, Content-Disposition: attachment)
- **FR-038**: System MUST handle PDF generation errors gracefully (chromedp not available, timeout) with a user-friendly error message

**Host-Based Routing**

- **FR-039**: Requests to `sistema.prospeccaobrasil.com` MUST serve only the internal system (auth + admin + CRUD); institutional pages return 404
- **FR-040**: Requests to `prospeccaobrasil.com` (and .com.br, www) MUST serve only the institutional site; internal system pages return 404
- **FR-041**: Requests to localhost (dev/test) MUST serve both public and internal routes (dev mode)

**Auth & Security**

- **FR-042**: All internal system pages (except `/login`, `/2fa/*`, `/healthz`) MUST require authentication via `SessionValidation` middleware
- **FR-043**: All internal system pages MUST require the `admin` role via `RequireRole(auth.RoleAdmin)` middleware
- **FR-044**: System MUST log all CRUD operations (create, update, soft-delete) via `slog` with entity type, entity ID, and user ID

### Key Entities *(include if feature involves data)*

- **Property**: A real estate listing (title, address, price, status, type, photos). Tenant-scoped. Soft-deletable. Links to prospections.
- **Client**: A person or company interested in properties (name, email, phone, CPF/CNPJ, budget, preferences). Tenant-scoped. Soft-deletable. PII under LGPD. Links to prospections and contacts.
- **Prospection**: A pipeline entry linking a client to a property with a status lifecycle (new -> contacting -> visiting -> negotiating -> closed_won/closed_lost). Tenant-scoped. Soft-deletable.
- **Contact**: An immutable interaction log entry (channel, direction, subject, body, contacted_at). Tenant-scoped. Linked to a client and optionally a prospection. Never deleted (LGPD audit trail).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Admin can create a property, client, and prospection in under 3 minutes total (3 forms, minimal fields)
- **SC-002**: Property list loads in under 500ms for 100 properties (paginated, indexed queries)
- **SC-003**: PDF generation completes in under 10 seconds for a property with up to 10 photos
- **SC-004**: Dashboard shows accurate counts and recent activity within 1 second
- **SC-005**: All CRUD operations enforce tenant_id isolation (cross-tenant access returns 404, verified by tests)
- **SC-006**: All internal pages redirect unauthenticated users to /login (verified by tests)
- **SC-007**: Contact log entries are immutable (no edit/delete UI or API, verified by tests)
- **SC-008**: Host-based routing correctly separates sistema.* (internal) from prospeccaobrasil.com (public), verified by tests

## Assumptions

- The database schema (properties, clients, prospections, contacts tables) already exists from SPEC-02 and does not need new migrations.
- The sqlc queries for CRUD operations already exist from SPEC-02 and may need additions for filtering/pagination.
- chromedp (headless Chrome) will be installed on the VPS for PDF generation. If not available, PDF generation returns an error but the rest of the system works.
- The MVP is single-admin (Luiz Claudio), so there is no multi-user UI, no role management, no assignment. All prospections/clients/properties belong to the single tenant.
- The Host-based routing quick fix (committed before this spec) already separates sistema.* from public domains. This spec extends the internal router with CRUD routes.
- Tailwind CSS design system classes (btn, badge, card, form-input, nav-link) from SPEC-04 are reused for the internal system UI.
- Pagination uses offset-based pagination (page + per_page query params) for MVP. Cursor pagination is a future enhancement.
- Photos are stored as a JSONB array of URLs in the properties table. No file upload is implemented in MVP -- the user pastes photo URLs.
- The internal system uses the same Go binary, same template engine, same static file server as the institutional site. No separate frontend.
