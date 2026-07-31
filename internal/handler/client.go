package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

// ClientHandler handles client CRUD operations.
type ClientHandler struct {
	queries *db.Queries
	tmpl    *template.Template
	log     *slog.Logger
}

// NewClientHandler creates a new ClientHandler.
func NewClientHandler(queries *db.Queries, tmpl *template.Template, log *slog.Logger) *ClientHandler {
	return &ClientHandler{queries: queries, tmpl: tmpl, log: log}
}

type clientPageData struct {
	Title        string
	ActivePage   string
	UserEmail    string
	UserRole     string
	Clients      []db.Client
	Client       *db.Client
	IsEdit       bool
	Form         clientForm
	Errors       clientErrors
	Page         int
	PerPage      int
	Total        int64
	TotalPages   int
	Filters      clientFilters
	Prospections []prospectWithNames
	Contacts     []db.Contact
	HasContacts  bool
}

type clientForm struct {
	Name        string
	Email       string
	Phone       string
	CpfCnpj     string
	Address     string
	Budget      string
	Preferences string
	Status      string
}

type clientErrors struct {
	Name    string
	Email   string
	Budget  string
	Generic string
}

type clientFilters struct {
	Status string
	Search string
}

func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	page, perPage := parsePagination(r, 20)
	filters := clientFilters{
		Status: r.URL.Query().Get("status"),
		Search: r.URL.Query().Get("search"),
	}

	var statusVal, searchVal string
	if filters.Status != "" {
		statusVal = filters.Status
	}
	if filters.Search != "" {
		searchVal = filters.Search
	}

	clients, err := h.queries.ListClientsFiltered(r.Context(), db.ListClientsFilteredParams{
		TenantID: user.TenantID,
		Column2:  statusVal,
		Column3:  searchVal,
		Limit:    int32(perPage),
		Offset:   int32((page - 1) * perPage),
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "client: list", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	total, err := h.queries.CountClientsFiltered(r.Context(), db.CountClientsFilteredParams{
		TenantID: user.TenantID,
		Column2:  statusVal,
		Column3:  searchVal,
	})
	if err != nil {
		total = 0
	}

	data := clientPageData{
		Title:      "Clientes",
		ActivePage: "clients",
		UserEmail:  user.Email,
		UserRole:   user.Role,
		Clients:    clients,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
		Filters:    filters,
	}

	if err := h.tmpl.ExecuteTemplate(w, "clients_list.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "client: render list", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *ClientHandler) New(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	data := clientPageData{
		Title:      "Novo Cliente",
		ActivePage: "clients",
		UserEmail:  user.Email,
		UserRole:   user.Role,
		IsEdit:     false,
		Form:       clientForm{Status: "lead"},
	}

	if err := h.tmpl.ExecuteTemplate(w, "clients_form.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "client: render form", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	form := parseClientForm(r)
	errors := validateClientForm(form)
	if hasClientErrors(errors) {
		h.renderClientFormError(w, r, user, form, errors, false, nil)
		return
	}

	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	client, err := h.queries.CreateClient(r.Context(), db.CreateClientParams{
		ID:          id,
		TenantID:    user.TenantID,
		Name:        form.Name,
		Email:       toPgText(form.Email),
		Phone:       toPgText(form.Phone),
		CpfCnpj:     toPgText(form.CpfCnpj),
		Address:     toPgText(form.Address),
		Budget:      toPgNumeric(form.Budget),
		Preferences: parsePreferencesJSON(form.Preferences),
		Status:      form.Status,
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "client: create", "error", err)
		errors.Generic = "Erro ao criar cliente."
		h.renderClientFormError(w, r, user, form, errors, false, nil)
		return
	}

	h.log.InfoContext(r.Context(), "client_created", "id", uuid.UUID(client.ID.Bytes).String(), "name", form.Name)
	http.Redirect(w, r, fmt.Sprintf("/clients/%s", uuid.UUID(client.ID.Bytes).String()), http.StatusSeeOther)
}

func (h *ClientHandler) Detail(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := parseChiUUID(r, "id")
	client, err := h.queries.GetClientByID(r.Context(), db.GetClientByIDParams{
		ID:       id,
		TenantID: user.TenantID,
	})
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Fetch prospections for this client
	prospectRows, _ := h.queries.ListProspectsByClientWithDetails(r.Context(), db.ListProspectsByClientWithDetailsParams{
		ClientID: id,
		TenantID: user.TenantID,
	})

	prospections := make([]prospectWithNames, 0, len(prospectRows))
	for _, p := range prospectRows {
		prospections = append(prospections, prospectWithNames{
			ID:             p.ID,
			Status:         p.Status,
			NextActionDate: p.NextActionDate,
			PropertyName:   p.PropertyTitle,
		})
	}

	// Fetch contact log
	contacts, _ := h.queries.ListContactsByClient(r.Context(), db.ListContactsByClientParams{
		ClientID: id,
		TenantID: user.TenantID,
	})

	data := clientPageData{
		Title:        client.Name,
		ActivePage:   "clients",
		UserEmail:    user.Email,
		UserRole:     user.Role,
		Client:       &client,
		Prospections: prospections,
		Contacts:     contacts,
		HasContacts:  len(contacts) > 0,
	}

	if err := h.tmpl.ExecuteTemplate(w, "clients_detail.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "client: render detail", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *ClientHandler) Edit(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := parseChiUUID(r, "id")
	client, err := h.queries.GetClientByID(r.Context(), db.GetClientByIDParams{
		ID:       id,
		TenantID: user.TenantID,
	})
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data := clientPageData{
		Title:      "Editar Cliente",
		ActivePage: "clients",
		UserEmail:  user.Email,
		UserRole:   user.Role,
		IsEdit:     true,
		Client:     &client,
		Form:       clientToForm(client),
	}

	if err := h.tmpl.ExecuteTemplate(w, "clients_form.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "client: render edit form", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := parseChiUUID(r, "id")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	form := parseClientForm(r)
	errors := validateClientForm(form)
	if hasClientErrors(errors) {
		client, _ := h.queries.GetClientByID(r.Context(), db.GetClientByIDParams{ID: id, TenantID: user.TenantID})
		h.renderClientFormError(w, r, user, form, errors, true, &client)
		return
	}

	client, err := h.queries.UpdateClient(r.Context(), db.UpdateClientParams{
		ID:          id,
		TenantID:    user.TenantID,
		Name:        form.Name,
		Email:       toPgText(form.Email),
		Phone:       toPgText(form.Phone),
		CpfCnpj:     toPgText(form.CpfCnpj),
		Address:     toPgText(form.Address),
		Budget:      toPgNumeric(form.Budget),
		Preferences: parsePreferencesJSON(form.Preferences),
		Status:      form.Status,
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "client: update", "error", err)
		errors.Generic = "Erro ao atualizar cliente."
		h.renderClientFormError(w, r, user, form, errors, true, nil)
		return
	}

	h.log.InfoContext(r.Context(), "client_updated", "id", uuid.UUID(client.ID.Bytes).String())
	http.Redirect(w, r, fmt.Sprintf("/clients/%s", uuid.UUID(client.ID.Bytes).String()), http.StatusSeeOther)
}

func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := parseChiUUID(r, "id")
	if err := h.queries.SoftDeleteClient(r.Context(), db.SoftDeleteClientParams{
		ID:       id,
		TenantID: user.TenantID,
	}); err != nil {
		h.log.ErrorContext(r.Context(), "client: delete", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.log.InfoContext(r.Context(), "client_deleted", "id", uuid.UUID(id.Bytes).String())
	http.Redirect(w, r, "/clients", http.StatusSeeOther)
}

// CreateContact handles creating a contact log entry for a client.
func (h *ClientHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	clientID := parseChiUUID(r, "id")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	channel := strings.TrimSpace(r.FormValue("channel"))
	direction := strings.TrimSpace(r.FormValue("direction"))
	subject := strings.TrimSpace(r.FormValue("subject"))
	body := strings.TrimSpace(r.FormValue("body"))

	if channel == "" || direction == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, err := h.queries.CreateContact(r.Context(), db.CreateContactParams{
		ID:          id,
		TenantID:    user.TenantID,
		ClientID:    clientID,
		ProspectID:  pgtype.UUID{Valid: false},
		Channel:     channel,
		Direction:   direction,
		Subject:     toPgText(subject),
		Body:        toPgText(body),
		ContactedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "contact: create for client", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.log.InfoContext(r.Context(), "contact_created", "id", uuid.UUID(id.Bytes).String(), "client_id", uuid.UUID(clientID.Bytes).String())

	if isHTMX(r) {
		contacts, _ := h.queries.ListContactsByClient(r.Context(), db.ListContactsByClientParams{
			ClientID: clientID,
			TenantID: user.TenantID,
		})
		data := struct {
			Contacts    []db.Contact
			HasContacts bool
		}{Contacts: contacts, HasContacts: len(contacts) > 0}
		if err := h.tmpl.ExecuteTemplate(w, "contacts_log.html", data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/clients/%s", uuid.UUID(clientID.Bytes).String()), http.StatusSeeOther)
}

func (h *ClientHandler) renderClientFormError(w http.ResponseWriter, r *http.Request, user *db.User, form clientForm, errors clientErrors, isEdit bool, client *db.Client) {
	data := clientPageData{
		Title:      cond(isEdit, "Editar Cliente", "Novo Cliente"),
		ActivePage: "clients",
		UserEmail:  user.Email,
		UserRole:   user.Role,
		IsEdit:     isEdit,
		Client:     client,
		Form:       form,
		Errors:     errors,
	}
	if err := h.tmpl.ExecuteTemplate(w, "clients_form.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "client: render form error", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func parseClientForm(r *http.Request) clientForm {
	return clientForm{
		Name:        strings.TrimSpace(r.FormValue("name")),
		Email:       strings.TrimSpace(r.FormValue("email")),
		Phone:       strings.TrimSpace(r.FormValue("phone")),
		CpfCnpj:     strings.TrimSpace(r.FormValue("cpf_cnpj")),
		Address:     strings.TrimSpace(r.FormValue("address")),
		Budget:      strings.TrimSpace(r.FormValue("budget")),
		Preferences: strings.TrimSpace(r.FormValue("preferences")),
		Status:      strings.TrimSpace(r.FormValue("status")),
	}
}

func validateClientForm(form clientForm) clientErrors {
	var e clientErrors
	if len(form.Name) < 2 {
		e.Name = "Nome deve ter pelo menos 2 caracteres"
	}
	if form.Email != "" {
		if _, err := mail.ParseAddress(form.Email); err != nil {
			e.Email = "Email inválido"
		}
	}
	return e
}

func hasClientErrors(e clientErrors) bool {
	return e.Name != "" || e.Email != "" || e.Budget != "" || e.Generic != ""
}

func clientToForm(c db.Client) clientForm {
	var prefs string
	if len(c.Preferences) > 0 && string(c.Preferences) != "{}" {
		var m map[string]any
		if err := json.Unmarshal(c.Preferences, &m); err == nil && len(m) > 0 {
			prefs = string(c.Preferences)
		}
	}
	return clientForm{
		Name:        c.Name,
		Email:       fromPgText(c.Email),
		Phone:       fromPgText(c.Phone),
		CpfCnpj:     fromPgText(c.CpfCnpj),
		Address:     fromPgText(c.Address),
		Budget:      fromPgNumeric(c.Budget),
		Preferences: prefs,
		Status:      c.Status,
	}
}

func parsePreferencesJSON(s string) []byte {
	if s == "" {
		return []byte("{}")
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return []byte("{}")
	}
	data, _ := json.Marshal(m)
	return data
}
