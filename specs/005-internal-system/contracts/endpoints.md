# Endpoint Contracts: Internal System (SPEC-05)

**Date**: 2026-07-31
**Base URL**: `https://sistema.prospeccaobrasil.com`
**Auth**: Session cookie (HttpOnly + SameSite=Strict + Secure). All endpoints except `/login`, `/2fa/*`, `/healthz` require `SessionValidation` + `RequireRole(admin)`.

## Dashboard

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/admin` | Yes | Dashboard with counts, prospections by status, recent prospections |

## Properties

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/properties` | Yes | Paginated property list (20/page). Query: `?page=1&per_page=20&status=available&type=commercial&search=São+Paulo` |
| GET | `/properties/new` | Yes | Property creation form |
| POST | `/properties` | Yes | Create property. Body: form-encoded. Redirects to `/properties/{id}` on success |
| GET | `/properties/{id}` | Yes | Property detail page. 404 if not found or wrong tenant |
| GET | `/properties/{id}/edit` | Yes | Property edit form (pre-filled) |
| POST | `/properties/{id}` | Yes | Update property. Body: form-encoded. Redirects to `/properties/{id}` |
| POST | `/properties/{id}/delete` | Yes | Soft-delete property. Redirects to `/properties` |

## Clients

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/clients` | Yes | Paginated client list (20/page). Query: `?page=1&per_page=20&status=lead&search=João` |
| GET | `/clients/new` | Yes | Client creation form |
| POST | `/clients` | Yes | Create client. Redirects to `/clients/{id}` |
| GET | `/clients/{id}` | Yes | Client detail page with prospections + contact log |
| GET | `/clients/{id}/edit` | Yes | Client edit form (pre-filled) |
| POST | `/clients/{id}` | Yes | Update client. Redirects to `/clients/{id}` |
| POST | `/clients/{id}/delete` | Yes | Soft-delete client. Redirects to `/clients` |

## Prospections

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/prospections` | Yes | Paginated prospection list (20/page). Query: `?page=1&per_page=20&status=negotiating` |
| GET | `/prospections/new` | Yes | Prospection creation form (client + property dropdowns) |
| POST | `/prospections` | Yes | Create prospection. Redirects to `/prospections/{id}` |
| GET | `/prospections/{id}` | Yes | Prospection detail with client, property, contact log |
| GET | `/prospections/{id}/edit` | Yes | Prospection edit form (status, notes, dates) |
| POST | `/prospections/{id}` | Yes | Update prospection. Redirects to `/prospections/{id}` |
| POST | `/prospections/{id}/delete` | Yes | Soft-delete prospection. Redirects to `/prospections` |

## Contacts

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/clients/{id}/contacts` | Yes | Create contact log entry for a client. HTMX: returns updated contact log fragment. Non-HTMX: redirects to `/clients/{id}` |
| POST | `/prospections/{id}/contacts` | Yes | Create contact log entry for a prospection. HTMX: returns updated contact log fragment. Non-HTMX: redirects to `/prospections/{id}` |

**Note**: Contacts are immutable. No GET (edit), PUT, or DELETE endpoints exist for individual contacts. Contacts are displayed inline on client/prospection detail pages.

## PDF

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/properties/{id}/pdf` | Yes | Generate PDF for property. Returns `Content-Type: application/pdf` with `Content-Disposition: attachment; filename="property-{id}.pdf"`. 500 if chromedp unavailable. |

## Response Formats

### HTML Pages (GET)
All GET endpoints return `text/html; charset=utf-8` with server-rendered templates.

### Form Submissions (POST)
- **Success**: HTTP 303 See Other (redirect to resource detail page) -- PRG pattern
- **Validation error**: HTTP 200 with form re-rendered (field errors shown)
- **HTMX success** (contacts): HTTP 200 with HTML fragment
- **HTMX error** (contacts): HTTP 200 with error fragment
- **Not found**: HTTP 404
- **Unauthorized**: HTTP 302 redirect to `/login`
- **Rate limited**: HTTP 429

### PDF
- **Success**: HTTP 200, `Content-Type: application/pdf`, `Content-Disposition: attachment`
- **Error**: HTTP 500 with HTML error page
