-- SPEC-06: Remove company column from contact_submissions.
ALTER TABLE contact_submissions
    DROP COLUMN IF EXISTS company;
