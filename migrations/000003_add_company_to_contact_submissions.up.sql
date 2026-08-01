-- SPEC-06: Add optional company field to contact_submissions
-- for the "Empresa" field in the redesigned Fale Conosco form.
ALTER TABLE contact_submissions
    ADD COLUMN company VARCHAR(255) NULL;
