package handler

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

const uniqueViolationCode = "23505"

// NewsletterHandler handles newsletter signup from the footer form.
type NewsletterHandler struct {
	queries *db.Queries
	tmpl    *template.Template
	log     *slog.Logger
	limiter *auth.RateLimiter
}

// NewNewsletterHandler creates a new NewsletterHandler.
func NewNewsletterHandler(queries *db.Queries, tmpl *template.Template, log *slog.Logger, limiter *auth.RateLimiter) *NewsletterHandler {
	return &NewsletterHandler{
		queries: queries,
		tmpl:    tmpl,
		log:     log,
		limiter: limiter,
	}
}

// Subscribe processes the newsletter signup POST.
// Returns an HTMX fragment if HX-Request header is present, otherwise redirects.
func (h *NewsletterHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))

	// Rate limit
	ip := clientIPFromRequest(r)
	if !h.limiter.AllowBoth(ip, email) {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	// Validate email
	if _, err := mail.ParseAddress(email); err != nil {
		h.renderError(w, r, "Email inválido")
		return
	}

	// Try to create subscriber
	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, err := h.queries.CreateNewsletterSubscriber(r.Context(), db.CreateNewsletterSubscriberParams{
		ID:    id,
		Email: email,
	})

	if err != nil {
		// Check for unique constraint violation (email already subscribed)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			h.renderSuccess(w, r, "Você já está inscrito!")
			return
		}
		// Check if already exists via separate query (fallback for pgx error type)
		if _, qErr := h.queries.GetNewsletterSubscriberByEmail(r.Context(), email); qErr == nil {
			h.renderSuccess(w, r, "Você já está inscrito!")
			return
		}
		h.log.ErrorContext(r.Context(), "newsletter: create subscriber", "error", err)
		h.renderError(w, r, "Erro interno. Tente novamente.")
		return
	}

	h.log.InfoContext(r.Context(), "newsletter_subscriber_created", "id", uuid.UUID(id.Bytes).String(), "email", email)
	h.renderSuccess(w, r, "Inscrição confirmada!")
}

func (h *NewsletterHandler) renderSuccess(w http.ResponseWriter, r *http.Request, message string) {
	data := struct{ Message string }{Message: message}
	if isHTMX(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := h.tmpl.ExecuteTemplate(w, "newsletter_success.html", data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	http.Redirect(w, r, "/?newsletter=success", http.StatusSeeOther)
}

func (h *NewsletterHandler) renderError(w http.ResponseWriter, r *http.Request, message string) {
	data := struct{ Message string }{Message: message}
	if isHTMX(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := h.tmpl.ExecuteTemplate(w, "newsletter_error.html", data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	http.Redirect(w, r, "/?newsletter=error", http.StatusSeeOther)
}
