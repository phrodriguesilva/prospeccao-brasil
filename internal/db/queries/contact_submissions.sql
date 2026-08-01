-- name: CreateContactSubmission :one
INSERT INTO contact_submissions (id, name, email, phone, company, subject, message)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListContactSubmissions :many
SELECT * FROM contact_submissions
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetContactSubmissionByID :one
SELECT * FROM contact_submissions WHERE id = $1;

-- name: UpdateContactSubmissionStatus :exec
UPDATE contact_submissions
SET status = $2, updated_at = now()
WHERE id = $1;
