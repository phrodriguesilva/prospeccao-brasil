# Data Model: Institutional Site & Design System

**Date**: 2026-07-31
**Spec**: [spec.md](spec.md)

## New Tables

### contact_submissions

Stores messages submitted via the "Fale Conosco" form. NOT tenant-scoped
(public institutional form, no tenant context).

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | UUID | PK, DEFAULT gen_random_uuid() | |
| name | VARCHAR(255) | NOT NULL | Visitor's full name |
| email | VARCHAR(255) | NOT NULL | Visitor's email (for reply) |
| phone | VARCHAR(20) | NULL | Optional, Brazilian format |
| subject | VARCHAR(255) | NOT NULL | Subject line |
| message | TEXT | NOT NULL | Message body (max 5000 chars app-level) |
| status | VARCHAR(20) | NOT NULL DEFAULT 'new', CHECK IN ('new','read','archived') | Admin workflow status |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() | |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() | |

**Indexes**: `idx_contact_submissions_created_at` (for chronological listing
in the internal system), `idx_contact_submissions_status` (for filtering
by status).

**Migration**: `migrations/2_contact_newsletter.sql` (forward-only)

### newsletter_subscribers

Stores email addresses subscribed to the newsletter. NOT tenant-scoped.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | UUID | PK, DEFAULT gen_random_uuid() | |
| email | VARCHAR(255) | NOT NULL, UNIQUE | Unique constraint = idempotency |
| subscribed_at | TIMESTAMPTZ | NOT NULL DEFAULT now() | |
| active | BOOLEAN | NOT NULL DEFAULT true | Soft unsubscribe (active=false) |

**Indexes**: `idx_newsletter_subscribers_email` (UNIQUE, for idempotency
check and lookup), `idx_newsletter_subscribers_active` (for listing
active subscribers).

**Migration**: `migrations/2_contact_newsletter.sql` (same file as
contact_submissions)

## Existing Tables (no changes)

No changes to existing tables (tenants, users, sessions, properties,
clients, prospections, contacts, audit_log). The institutional site
tables are independent of the internal system tables.

## sqlc Queries

### contact_submissions.sql

```sql
-- name: CreateContactSubmission :one
INSERT INTO contact_submissions (id, name, email, phone, subject, message)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListContactSubmissions :many
SELECT * FROM contact_submissions ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: GetContactSubmissionByID :one
SELECT * FROM contact_submissions WHERE id = $1;

-- name: UpdateContactSubmissionStatus :exec
UPDATE contact_submissions SET status = $2, updated_at = now() WHERE id = $1;
```

### newsletter_subscribers.sql

```sql
-- name: CreateNewsletterSubscriber :one
INSERT INTO newsletter_subscribers (id, email)
VALUES ($1, $2)
RETURNING *;

-- name: GetNewsletterSubscriberByEmail :one
SELECT * FROM newsletter_subscribers WHERE email = $1;

-- name: ListActiveNewsletterSubscribers :many
SELECT * FROM newsletter_subscribers WHERE active = true ORDER BY subscribed_at DESC;
```

## State Transitions

### contact_submissions.status

```
new --(admin reads)--> read --(admin archives)--> archived
```

No reverse transitions (forward-only workflow). The admin can also
archive directly from `new` (skip `read`).

### newsletter_subscribers.active

```
true --(unsubscribe)--> false
```

No reverse transition in MVP (resubscribing creates a new row or the
admin reactivates). The UNIQUE constraint on email means resubscribing
after unsubscribe will hit the unique violation -- the handler catches
it and reactivates the existing record (sets active=true).

## Validation Rules (server-side, enforced in Go handler)

| Field | Rule | Implementation |
|-------|------|----------------|
| name | required, 2-255 chars | `len(strings.TrimSpace(name)) >= 2` |
| email | required, valid format | `net/mail.ParseAddress(email)` |
| phone | optional, 10-20 chars | `len(phone) == 0 \|\| (len(phone) >= 10 && len(phone) <= 20)` |
| subject | required, 2-255 chars | `len(strings.TrimSpace(subject)) >= 2` |
| message | required, 10-5000 chars | `len(strings.TrimSpace(message)) >= 10` |
| newsletter email | required, valid format | `net/mail.ParseAddress(email)` |
