# Research: Internal System (SPEC-05)

**Date**: 2026-07-31

## R1: chromedp for PDF generation

**Decision**: Use chromedp with headless Chrome to render an HTML template and print to PDF.

**Rationale**: The PRD specifies chromedp for PDF generation. chromedp drives a headless Chrome browser via the DevTools Protocol, allowing full CSS/HTML rendering (including Tailwind styles) before printing to PDF. This produces professional-looking documents without a separate PDF library.

**Alternatives considered**:
- **gofpdf / gopdf**: Pure Go PDF libraries. Rejected -- limited layout support, no CSS, difficult to include photos with proper sizing.
- **wkhtmltopdf**: External binary. Rejected -- requires separate installation, less maintained, security concerns.
- **weasyprint**: Python-based. Rejected -- adds Python dependency to a Go project.

**Implementation notes**:
- chromedp requires Chrome/Chromium installed on the server. On Ubuntu 24.04: `apt install chromium-browser`.
- Use `chromedp.NewExecAllocator` with `--headless --no-sandbox --disable-gpu` flags.
- Render an HTML template to a temp file, navigate to it via `file://` URL, then `page.PrintToPDF`.
- Timeout: 30 seconds max for PDF generation.
- If Chrome is not available, return a 500 error with a user-friendly message.

## R2: Pagination strategy

**Decision**: Offset-based pagination with `page` and `per_page` query parameters (default per_page=20, max 100).

**Rationale**: For the MVP scale (single admin, < 1000 records), offset pagination is simple and sufficient. Cursor pagination adds complexity (encoding/decoding cursors, handling sort order changes) that is not justified for this scale.

**Alternatives considered**:
- **Cursor pagination**: Better for large datasets and real-time feeds. Rejected for MVP -- premature optimization.
- **No pagination (load all)**: Rejected -- does not scale, poor UX with 100+ records.

**Implementation notes**:
- Query params: `?page=1&per_page=20`
- SQL: `LIMIT $limit OFFSET $offset` (sqlc params)
- UI: "Previous" / "Next" buttons + page number display
- Edge case: if `page * per_page > total`, show last page or empty state

## R3: Filtering and search

**Decision**: URL-synced filter state via query parameters. Server-side filtering in SQL.

**Rationale**: URL-synced filters allow bookmarking, sharing, and back-button support. Server-side filtering leverages PostgreSQL indexes and avoids client-side performance issues.

**Implementation notes**:
- Properties: `?status=available&type=commercial&search=São+Paulo`
- Clients: `?status=lead&search=João`
- Prospections: `?status=negotiating`
- Search uses `ILIKE` for case-insensitive matching on title/city (properties) or name/email (clients)
- New sqlc queries needed: `ListPropertiesFiltered`, `ListClientsFiltered`, `ListProspectsFiltered`

## R4: Internal system layout

**Decision**: Sidebar navigation layout for the internal system, separate from the institutional site's top nav.

**Rationale**: The internal system is a CRUD application (dashboard, lists, forms) which benefits from a persistent sidebar for navigation between entities. The institutional site is a marketing site with a top nav. Different layouts reflect different use cases.

**Implementation notes**:
- `_layout.html` in `internal/template/admin/` provides the sidebar + content area
- Sidebar links: Dashboard, Imóveis, Clientes, Prospecções
- Active state based on current path
- User info (email, role) at bottom of sidebar with logout button
- Mobile: sidebar collapses via Alpine.js `x-data`/`x-show`

## R5: Form validation pattern

**Decision**: Server-side validation with field-level error messages rendered in the template. HTMX for async form submission where appropriate (contact log), full page POST for create/edit (redirect after success).

**Rationale**: Server-side validation is the source of truth (never trust client-side). HTMX async is used for inline operations (contact log creation without page reload). Full page POST + redirect is used for create/edit to ensure the URL reflects the resource (PRG pattern -- Post/Redirect/Get).

**Implementation notes**:
- Validation in handler methods (not a separate validation layer -- YAGNI for MVP)
- Errors rendered as `<p class="form-error">` below each field
- PRG pattern: POST -> 303 See Other -> GET (resource detail page)
- HTMX for contact log: POST -> 200 with HTML fragment (updated log)

## R6: Dashboard count queries

**Decision**: New sqlc queries for dashboard counts: `CountPropertiesByTenant`, `CountClientsByTenant`, `CountProspectsByTenant`, `CountProspectsByStatus`, `ListRecentProspects`.

**Rationale**: The existing `ListPropertiesByTenant` returns all rows which is inefficient for counting. Dedicated count queries use `SELECT COUNT(*)` and are indexed.

**Implementation notes**:
- `CountPropertiesByTenant`: `SELECT COUNT(*) FROM properties WHERE tenant_id = $1 AND deleted_at IS NULL`
- `CountProspectsByStatus`: `SELECT status, COUNT(*) FROM prospections WHERE tenant_id = $1 AND deleted_at IS NULL GROUP BY status`
- `ListRecentProspects`: `SELECT * FROM prospections WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 5`
- Dashboard also needs client/property names for recent prospections -- use a JOIN query

## R7: Soft delete confirmation

**Decision**: Soft delete uses a confirmation modal (Alpine.js) before POSTing to the delete endpoint.

**Rationale**: Prevents accidental deletions. The modal is a simple Alpine.js component with a confirm button. No external modal library needed.

**Implementation notes**:
- Button: `<button @click="$dispatch('open-modal', {id: 'delete-{id}'})">Excluir</button>`
- Modal: Alpine.js `x-data` with `x-show` for visibility
- Confirm button: form POST to `/properties/{id}/delete`
- No undo -- soft delete is permanent from the UI perspective (recovery is a DB operation)

## R8: chromedp dependency and VPS installation

**Decision**: Add `github.com/chromedp/chromedp` as a Go dependency. Install Chromium on the VPS via `apt install chromium-browser`.

**Rationale**: chromedp is a pure Go library that communicates with Chrome via DevTools Protocol. It does not bundle Chrome -- Chrome must be installed separately on the server. On Ubuntu 24.04, `chromium-browser` is the standard package.

**Alternatives considered**:
- **rod (go-rod/rod)**: Similar to chromedp but higher-level. Rejected -- chromedp is specified in the PRD and is more widely used.
- **playwright-go**: Requires Node.js for installation. Rejected -- adds Node.js runtime dependency.

**Implementation notes**:
- `go get github.com/chromedp/chromedp`
- VPS: `apt install chromium-browser` (user handles deployment)
- CI: chromedp tests are skipped if Chrome is not available (env check)
- PDF handler: `chromedp.NewExecAllocator` with `--headless --no-sandbox --disable-gpu --disable-dev-shm-usage`
