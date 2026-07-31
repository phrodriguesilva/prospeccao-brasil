-- name: CreateNewsletterSubscriber :one
INSERT INTO newsletter_subscribers (id, email)
VALUES ($1, $2)
RETURNING *;

-- name: GetNewsletterSubscriberByEmail :one
SELECT * FROM newsletter_subscribers WHERE email = $1;

-- name: ListActiveNewsletterSubscribers :many
SELECT * FROM newsletter_subscribers
WHERE active = true
ORDER BY subscribed_at DESC;
