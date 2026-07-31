-- SPEC-04: Down migration for contact_submissions and newsletter_subscribers.
-- NOTE: Down migrations are for development only. Production is forward-only.

DROP TABLE IF EXISTS newsletter_subscribers;
DROP TABLE IF EXISTS contact_submissions;
