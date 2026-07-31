# Quickstart Validation Guide: Internal System (SPEC-05)

**Date**: 2026-07-31

## Prerequisites

- Go 1.26+, PostgreSQL 16+, sqlc, golangci-lint, ast-grep
- `DATABASE_URL` set in `.env.local`
- `ENCRYPTION_KEY` set in `.env.local`
- `make check` passes before starting

## Setup

```bash
# 1. Apply migrations (schema already exists from SPEC-02)
make migrate

# 2. Generate sqlc (includes new queries)
make sqlc

# 3. Build CSS (if new classes added)
make build-css

# 4. Run the server
make dev
# Server starts on :8080
```

## Validation Scenarios

### Scenario 1: Dashboard (FR-001 to FR-004)

1. Go to `http://localhost:8080/login`
2. Login with admin credentials
3. Verify redirect to `/admin` (dashboard)
4. **Expected**: Dashboard shows counts (0 for fresh tenant), prospections by status (empty), recent prospections (empty state with CTA)

### Scenario 2: Create Property (FR-005 to FR-012)

1. From dashboard, click "Imóveis" in sidebar
2. **Expected**: Property list page with empty state
3. Click "Novo Imóvel"
4. Fill form: title="Sala Comercial Vila Mariana", address="Rua Vergueiro 1000", city="São Paulo", state="SP", price="500000", type="commercial", status="available"
5. Submit
6. **Expected**: Redirect to property detail page showing all entered data
7. Go back to `/properties`
8. **Expected**: Property appears in list

### Scenario 3: Filter Properties (FR-006)

1. Create 3 properties: 2 "available" + 1 "sold", 2 "commercial" + 1 "residential"
2. Go to `/properties?status=available`
3. **Expected**: Only 2 available properties shown
4. Go to `/properties?type=commercial`
5. **Expected**: Only 2 commercial properties shown
6. Go to `/properties?search=São+Paulo`
7. **Expected**: Only properties with "São Paulo" in title or city shown

### Scenario 4: Edit Property (FR-010)

1. Go to a property detail page
2. Click "Editar"
3. **Expected**: Form pre-filled with current values
4. Change price to "600000"
5. Submit
6. **Expected**: Redirect to detail page with new price

### Scenario 5: Soft Delete Property (FR-011)

1. Go to a property detail page
2. Click "Excluir"
3. **Expected**: Confirmation modal appears
4. Confirm
5. **Expected**: Redirect to `/properties`, deleted property not in list
6. Verify in DB: `SELECT deleted_at FROM properties WHERE id = '{id}'` -- deleted_at is set

### Scenario 6: Create Client (FR-013 to FR-020)

1. Go to `/clients`
2. Click "Novo Cliente"
3. Fill: name="João Silva", email="joao@example.com", phone="+55 11 99999-9999", status="lead"
4. Submit
5. **Expected**: Redirect to client detail page
6. Go to `/clients`
7. **Expected**: Client appears in list

### Scenario 7: Create Prospection (FR-021 to FR-028)

1. Ensure at least 1 client and 1 property exist
2. Go to `/prospections`
3. Click "Nova Prospecção"
4. Select client and property from dropdowns, set status="new"
5. Submit
6. **Expected**: Redirect to prospection detail page showing client name, property title, status badge

### Scenario 8: Update Prospection Status (FR-026)

1. Go to a prospection detail page
2. Click "Editar"
3. Change status from "new" to "contacting"
4. Set next_action_date to a future date
5. Submit
6. **Expected**: Status badge updated, next action date shown

### Scenario 9: Create Contact Log (FR-029 to FR-033)

1. Go to a client detail page
2. Click "Registrar Contato"
3. Fill: channel="phone", direction="outbound", subject="Follow-up", body="Cliente interessado"
4. Submit (HTMX)
5. **Expected**: Contact appears in log without page reload
6. Verify: no edit or delete buttons on the contact entry

### Scenario 10: Contact Log on Prospection (FR-029, FR-032)

1. Go to a prospection detail page
2. Click "Registrar Contato"
3. Fill: channel="email", direction="outbound", subject="Enviei informações"
4. Submit
5. **Expected**: Contact appears in prospection's contact log, also linked to the client

### Scenario 11: PDF Generation (FR-034 to FR-038)

1. Go to a property detail page with photos
2. Click "Gerar PDF"
3. **Expected**: PDF downloads with property details and photos
4. Verify: file is valid PDF (check with `file` command or open in viewer)
5. If chromedp not available: **Expected**: Error message shown, system does not crash

### Scenario 12: Auth Required (FR-042, FR-043)

1. Open a new incognito window (no session)
2. Visit `http://localhost:8080/properties`
3. **Expected**: Redirect to `/login`
4. Visit `http://localhost:8080/admin`
5. **Expected**: Redirect to `/login`

### Scenario 13: Tenant Isolation (FR-012, FR-020, FR-028, FR-033)

1. Create a second tenant in the DB with a property
2. Login as admin of first tenant
3. Try to access `/properties/{second-tenant-property-id}`
4. **Expected**: 404 (not found, because tenant_id filter excludes it)

### Scenario 14: Host-Based Routing (FR-039 to FR-041)

1. Send request with `Host: sistema.prospeccaobrasil.com` to `/properties`
2. **Expected**: 200 (internal system serves properties)
3. Send request with `Host: sistema.prospeccaobrasil.com` to `/quem-somos`
4. **Expected**: 404 (institutional page not on internal subdomain)
5. Send request with `Host: prospeccaobrasil.com` to `/properties`
6. **Expected**: 404 (internal page not on public domain)
7. Send request with `Host: localhost` to `/properties`
8. **Expected**: 200 (dev mode serves everything)

### Scenario 15: Pagination (FR-005, FR-013, FR-021)

1. Create 25 properties
2. Go to `/properties`
3. **Expected**: 20 properties on page 1, "Next" button enabled
4. Click "Next" or go to `/properties?page=2`
5. **Expected**: 5 properties on page 2, "Previous" button enabled

### Scenario 16: Make Check (SC-005 to SC-008)

```bash
make check
```
**Expected**: golangci-lint 0 issues, all tests pass, coverage >= 85%, build succeeds, ast-grep 0 errors.
