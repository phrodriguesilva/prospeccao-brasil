package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"prospeccaobrasil/internal/db"
)

// InstitutionalHandler handles public institutional site pages.
// All pages are public (no auth required) per FR-023.
type InstitutionalHandler struct {
	queries *db.Queries
	tmpl    *template.Template
	log     *slog.Logger
}

// NewInstitutionalHandler creates a new InstitutionalHandler.
func NewInstitutionalHandler(queries *db.Queries, tmpl *template.Template, log *slog.Logger) *InstitutionalHandler {
	return &InstitutionalHandler{
		queries: queries,
		tmpl:    tmpl,
		log:     log,
	}
}

// pageData is the base data passed to all institutional templates.
type pageData struct {
	ActivePage   string
	Success      bool
	Form         contactForm
	Errors       contactErrors
	Testimonials []testimonial
}

type contactForm struct {
	Name    string
	Email   string
	Phone   string
	Subject string
	Message string
}

type contactErrors struct {
	Name    string
	Email   string
	Subject string
	Message string
	Generic string
}

type testimonial struct {
	Name    string
	Company string
	Quote   string
	Metric  string
}

// Home renders the home page at GET /.
func (h *InstitutionalHandler) Home(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "home.html", pageData{ActivePage: "home"})
}

// QuemSomos renders the "Quem somos" page at GET /quem-somos.
func (h *InstitutionalHandler) QuemSomos(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "quem-somos.html", pageData{ActivePage: "quem-somos"})
}

// Servicos renders the "Servicos" page at GET /servicos.
func (h *InstitutionalHandler) Servicos(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "servicos.html", pageData{ActivePage: "servicos"})
}

// NossosClientes renders the "Nossos clientes" page at GET /nossos-clientes.
func (h *InstitutionalHandler) NossosClientes(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "nossos-clientes.html", pageData{ActivePage: "nossos-clientes"})
}

// FaleConosco renders the "Fale Conosco" page at GET /fale-conosco.
func (h *InstitutionalHandler) FaleConosco(w http.ResponseWriter, r *http.Request) {
	data := pageData{ActivePage: "fale-conosco"}
	if r.URL.Query().Get("success") == "1" {
		data.Success = true
	}
	h.renderPage(w, "fale-conosco.html", data)
}

// NotFound renders the 404 page for unmatched routes.
func (h *InstitutionalHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	h.renderPage(w, "404.html", pageData{})
}

// renderPage renders a page using the base.html layout.
func (h *InstitutionalHandler) renderPage(w http.ResponseWriter, name string, data pageData) {
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		h.log.Error("render page", "template", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
