-- SPEC-04: Contact submissions and newsletter subscribers tables.
-- These are public institutional site forms (NOT tenant-scoped).

-- Contact form submissions from "Fale Conosco"
CREATE TABLE IF NOT EXISTS contact_submissions (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL,
    phone      VARCHAR(20)  NULL,
    subject    VARCHAR(255) NOT NULL,
    message    TEXT         NOT NULL,
    status     VARCHAR(20)  NOT NULL DEFAULT 'new'
        CHECK (status IN ('new', 'read', 'archived')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_contact_submissions_created_at ON contact_submissions (created_at DESC);
CREATE INDEX idx_contact_submissions_status ON contact_submissions (status);

-- Newsletter subscribers from footer form
CREATE TABLE IF NOT EXISTS newsletter_subscribers (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    subscribed_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    active        BOOLEAN      NOT NULL DEFAULT true
);

CREATE INDEX idx_newsletter_subscribers_active ON newsletter_subscribers (active) WHERE active = true;
