package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

// ContactHandler handles the "Fale Conosco" form submission.
type ContactHandler struct {
	queries *db.Queries
	tmpl    *template.Template
	log     *slog.Logger
	limiter *auth.RateLimiter
}

// NewContactHandler creates a new ContactHandler.
func NewContactHandler(queries *db.Queries, tmpl *template.Template, log *slog.Logger, limiter *auth.RateLimiter) *ContactHandler {
	return &ContactHandler{
		queries: queries,
		tmpl:    tmpl,
		log:     log,
		limiter: limiter,
	}
}

// Submit processes the contact form POST.
// Returns an HTMX fragment if HX-Request header is present, otherwise redirects.
func (h *ContactHandler) Submit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	form := contactForm{
		Company: strings.TrimSpace(r.FormValue("company")),
		Name:    strings.TrimSpace(r.FormValue("name")),
		Email:   strings.TrimSpace(r.FormValue("email")),
		Phone:   strings.TrimSpace(r.FormValue("phone")),
		Subject: strings.TrimSpace(r.FormValue("subject")),
		Message: strings.TrimSpace(r.FormValue("message")),
	}

	// Rate limit
	ip := clientIPFromRequest(r)
	if !h.limiter.AllowBoth(ip, form.Email) {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	// Validate
	errors := h.validate(form)
	if hasErrors(errors) {
		h.renderError(w, r, form, errors)
		return
	}

	// Persist
	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	var phonePtr *string
	if form.Phone != "" {
		phonePtr = &form.Phone
	}
	var companyPtr *string
	if form.Company != "" {
		companyPtr = &form.Company
	}
	_, err := h.queries.CreateContactSubmission(r.Context(), db.CreateContactSubmissionParams{
		ID:      id,
		Name:    form.Name,
		Email:   form.Email,
		Phone:   phonePtr,
		Company: companyPtr,
		Subject: form.Subject,
		Message: form.Message,
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "contact: create submission", "error", err)
		h.renderError(w, r, form, contactErrors{Generic: "Erro interno. Tente novamente."})
		return
	}

	h.log.InfoContext(r.Context(), "contact_submission_created", "id", uuid.UUID(id.Bytes).String(), "email", form.Email)

	// Return success
	if isHTMX(r) {
		if err := h.tmpl.ExecuteTemplate(w, "contact_success.html", nil); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	http.Redirect(w, r, "/fale-conosco?success=1", http.StatusSeeOther)
}

func (h *ContactHandler) validate(form contactForm) contactErrors {
	var e contactErrors
	if len(form.Name) < 2 {
		e.Name = "Nome deve ter pelo menos 2 caracteres"
	}
	if _, err := mail.ParseAddress(form.Email); err != nil {
		e.Email = "Email inválido"
	}
	if len(form.Subject) < 2 {
		e.Subject = "Assunto deve ter pelo menos 2 caracteres"
	}
	if len(form.Message) < 10 {
		e.Message = "Mensagem deve ter pelo menos 10 caracteres"
	}
	return e
}

func (h *ContactHandler) renderError(w http.ResponseWriter, r *http.Request, form contactForm, errors contactErrors) {
	data := pageData{
		ActivePage: "fale-conosco",
		Form:       form,
		Errors:     errors,
	}
	if isHTMX(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := h.tmpl.ExecuteTemplate(w, "contact_error.html", data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	// Non-HTMX: render the full page with errors
	if err := h.tmpl.ExecuteTemplate(w, "fale-conosco.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func hasErrors(e contactErrors) bool {
	return e.Name != "" || e.Email != "" || e.Subject != "" || e.Message != "" || e.Generic != ""
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
