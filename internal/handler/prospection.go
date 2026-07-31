package handler

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

// ProspectionHandler handles prospection CRUD operations.
type ProspectionHandler struct {
	queries *db.Queries
	tmpl    *template.Template
	log     *slog.Logger
}

// NewProspectionHandler creates a new ProspectionHandler.
func NewProspectionHandler(queries *db.Queries, tmpl *template.Template, log *slog.Logger) *ProspectionHandler {
	return &ProspectionHandler{queries: queries, tmpl: tmpl, log: log}
}

type prospectionPageData struct {
	Title         string
	ActivePage    string
	UserEmail     string
	UserRole      string
	Prospects     []db.Prospection
	Prospect      *db.Prospection
	IsEdit        bool
	Form          prospectionForm
	Errors        prospectionErrors
	Page          int
	PerPage       int
	Total         int64
	TotalPages    int
	Filters       prospectionFilters
	Clients       []db.Client
	Properties    []db.Property
	Client        *db.Client
	Property      *db.Property
	Contacts      []db.Contact
	HasContacts   bool
	HasClients    bool
	HasProperties bool
}

type prospectionForm struct {
	ClientID       string
	PropertyID     string
	Status         string
	Notes          string
	ContactDate    string
	NextActionDate string
}

type prospectionErrors struct {
	ClientID   string
	PropertyID string
	Status     string
	Generic    string
}

type prospectionFilters struct {
	Status string
}

func (h *ProspectionHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	page, perPage := parsePagination(r, 20)
	filters := prospectionFilters{
		Status: r.URL.Query().Get("status"),
	}

	var statusVal string
	if filters.Status != "" {
		statusVal = filters.Status
	}

	prospects, err := h.queries.ListProspectsFiltered(r.Context(), db.ListProspectsFilteredParams{
		TenantID: user.TenantID,
		Column2:  statusVal,
		Limit:    int32(perPage),
		Offset:   int32((page - 1) * perPage),
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "prospection: list", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	total, _ := h.queries.CountProspectsFiltered(r.Context(), db.CountProspectsFilteredParams{
		TenantID: user.TenantID,
		Column2:  statusVal,
	})

	// Enrich with client/property names
	rows := make([]prospectListRow, 0, len(prospects))
	for _, p := range prospects {
		client, _ := h.queries.GetClientByID(r.Context(), db.GetClientByIDParams{ID: p.ClientID, TenantID: user.TenantID})
		prop, _ := h.queries.GetPropertyByID(r.Context(), db.GetPropertyByIDParams{ID: p.PropertyID, TenantID: user.TenantID})
		rows = append(rows, prospectListRow{
			Prospect:      p,
			ClientName:    client.Name,
			PropertyTitle: prop.Title,
		})
	}

	data := prospectionPageData{
		Title:      "Prospecções",
		ActivePage: "prospections",
		UserEmail:  user.Email,
		UserRole:   user.Role,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
		Filters:    filters,
	}

	templateData := struct {
		prospectionPageData
		Rows []prospectListRow
	}{
		prospectionPageData: data,
		Rows:                rows,
	}

	if err := h.tmpl.ExecuteTemplate(w, "prospections_list.html", templateData); err != nil {
		h.log.ErrorContext(r.Context(), "prospection: render list", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// prospectListRow is a prospection with client/property names for list display.
type prospectListRow struct {
	Prospect      db.Prospection
	ClientName    string
	PropertyTitle string
}

func (h *ProspectionHandler) New(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	clients, _ := h.queries.ListClientsByTenant(r.Context(), user.TenantID)
	properties, _ := h.queries.ListPropertiesByTenant(r.Context(), user.TenantID)

	data := prospectionPageData{
		Title:         "Nova Prospecção",
		ActivePage:    "prospections",
		UserEmail:     user.Email,
		UserRole:      user.Role,
		IsEdit:        false,
		Form:          prospectionForm{Status: "new"},
		Clients:       clients,
		Properties:    properties,
		HasClients:    len(clients) > 0,
		HasProperties: len(properties) > 0,
	}

	if err := h.tmpl.ExecuteTemplate(w, "prospections_form.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "prospection: render form", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *ProspectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	form := parseProspectionForm(r)
	errors := validateProspectionForm(form)
	if hasProspectionErrors(errors) {
		clients, _ := h.queries.ListClientsByTenant(r.Context(), user.TenantID)
		properties, _ := h.queries.ListPropertiesByTenant(r.Context(), user.TenantID)
		h.renderProspectionFormError(w, r, user, form, errors, false, clients, properties)
		return
	}

	clientID := parseUUIDParam(form.ClientID)
	propertyID := parseUUIDParam(form.PropertyID)

	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	prospect, err := h.queries.CreateProspect(r.Context(), db.CreateProspectParams{
		ID:             id,
		TenantID:       user.TenantID,
		ClientID:       clientID,
		PropertyID:     propertyID,
		Status:         form.Status,
		Notes:          toPgText(form.Notes),
		ContactDate:    parsePgTimestamp(form.ContactDate),
		NextActionDate: parsePgTimestamp(form.NextActionDate),
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "prospection: create", "error", err)
		errors.Generic = "Erro ao criar prospecção."
		clients, _ := h.queries.ListClientsByTenant(r.Context(), user.TenantID)
		properties, _ := h.queries.ListPropertiesByTenant(r.Context(), user.TenantID)
		h.renderProspectionFormError(w, r, user, form, errors, false, clients, properties)
		return
	}

	h.log.InfoContext(r.Context(), "prospection_created", "id", uuid.UUID(prospect.ID.Bytes).String())
	http.Redirect(w, r, fmt.Sprintf("/prospections/%s", uuid.UUID(prospect.ID.Bytes).String()), http.StatusSeeOther)
}

func (h *ProspectionHandler) Detail(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := parseChiUUID(r, "id")
	prospect, err := h.queries.GetProspectByID(r.Context(), db.GetProspectByIDParams{
		ID:       id,
		TenantID: user.TenantID,
	})
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	client, _ := h.queries.GetClientByID(r.Context(), db.GetClientByIDParams{ID: prospect.ClientID, TenantID: user.TenantID})
	property, _ := h.queries.GetPropertyByID(r.Context(), db.GetPropertyByIDParams{ID: prospect.PropertyID, TenantID: user.TenantID})

	contacts, _ := h.queries.ListContactsByProspect(r.Context(), db.ListContactsByProspectParams{
		ProspectID: id,
		TenantID:   user.TenantID,
	})

	data := prospectionPageData{
		Title:       "Prospecção",
		ActivePage:  "prospections",
		UserEmail:   user.Email,
		UserRole:    user.Role,
		Prospect:    &prospect,
		Client:      &client,
		Property:    &property,
		Contacts:    contacts,
		HasContacts: len(contacts) > 0,
	}

	if err := h.tmpl.ExecuteTemplate(w, "prospections_detail.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "prospection: render detail", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *ProspectionHandler) Edit(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := parseChiUUID(r, "id")
	prospect, err := h.queries.GetProspectByID(r.Context(), db.GetProspectByIDParams{
		ID:       id,
		TenantID: user.TenantID,
	})
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data := prospectionPageData{
		Title:      "Editar Prospecção",
		ActivePage: "prospections",
		UserEmail:  user.Email,
		UserRole:   user.Role,
		IsEdit:     true,
		Prospect:   &prospect,
		Form:       prospectionToForm(prospect),
	}

	if err := h.tmpl.ExecuteTemplate(w, "prospections_form.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "prospection: render edit form", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *ProspectionHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	form := parseProspectionForm(r)
	if form.Status == "" {
		form.Status = "new"
	}

	prospect, err := h.queries.UpdateProspect(r.Context(), db.UpdateProspectParams{
		ID:             id,
		TenantID:       user.TenantID,
		Status:         form.Status,
		Notes:          toPgText(form.Notes),
		NextActionDate: parsePgTimestamp(form.NextActionDate),
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "prospection: update", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.log.InfoContext(r.Context(), "prospection_updated", "id", uuid.UUID(prospect.ID.Bytes).String())
	http.Redirect(w, r, fmt.Sprintf("/prospections/%s", uuid.UUID(prospect.ID.Bytes).String()), http.StatusSeeOther)
}

func (h *ProspectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := parseChiUUID(r, "id")
	if err := h.queries.SoftDeleteProspect(r.Context(), db.SoftDeleteProspectParams{
		ID:       id,
		TenantID: user.TenantID,
	}); err != nil {
		h.log.ErrorContext(r.Context(), "prospection: delete", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.log.InfoContext(r.Context(), "prospection_deleted", "id", uuid.UUID(id.Bytes).String())
	http.Redirect(w, r, "/prospections", http.StatusSeeOther)
}

// CreateContact handles creating a contact log entry for a prospection.
func (h *ProspectionHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	prospectID := parseChiUUID(r, "id")

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

	// Get the prospect to find the client_id
	prospect, err := h.queries.GetProspectByID(r.Context(), db.GetProspectByIDParams{
		ID:       prospectID,
		TenantID: user.TenantID,
	})
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, err = h.queries.CreateContact(r.Context(), db.CreateContactParams{
		ID:          id,
		TenantID:    user.TenantID,
		ClientID:    prospect.ClientID,
		ProspectID:  prospectID,
		Channel:     channel,
		Direction:   direction,
		Subject:     toPgText(subject),
		Body:        toPgText(body),
		ContactedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "contact: create for prospect", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.log.InfoContext(r.Context(), "contact_created", "id", uuid.UUID(id.Bytes).String(), "prospect_id", uuid.UUID(prospectID.Bytes).String())

	if isHTMX(r) {
		contacts, _ := h.queries.ListContactsByProspect(r.Context(), db.ListContactsByProspectParams{
			ProspectID: prospectID,
			TenantID:   user.TenantID,
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

	http.Redirect(w, r, fmt.Sprintf("/prospections/%s", uuid.UUID(prospectID.Bytes).String()), http.StatusSeeOther)
}

func (h *ProspectionHandler) renderProspectionFormError(w http.ResponseWriter, r *http.Request, user *db.User, form prospectionForm, errors prospectionErrors, isEdit bool, clients []db.Client, properties []db.Property) {
	data := prospectionPageData{
		Title:         cond(isEdit, "Editar Prospecção", "Nova Prospecção"),
		ActivePage:    "prospections",
		UserEmail:     user.Email,
		UserRole:      user.Role,
		IsEdit:        isEdit,
		Form:          form,
		Errors:        errors,
		Clients:       clients,
		Properties:    properties,
		HasClients:    len(clients) > 0,
		HasProperties: len(properties) > 0,
	}
	if err := h.tmpl.ExecuteTemplate(w, "prospections_form.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "prospection: render form error", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func parseProspectionForm(r *http.Request) prospectionForm {
	return prospectionForm{
		ClientID:       strings.TrimSpace(r.FormValue("client_id")),
		PropertyID:     strings.TrimSpace(r.FormValue("property_id")),
		Status:         strings.TrimSpace(r.FormValue("status")),
		Notes:          strings.TrimSpace(r.FormValue("notes")),
		ContactDate:    strings.TrimSpace(r.FormValue("contact_date")),
		NextActionDate: strings.TrimSpace(r.FormValue("next_action_date")),
	}
}

func validateProspectionForm(form prospectionForm) prospectionErrors {
	var e prospectionErrors
	if form.ClientID == "" {
		e.ClientID = "Cliente é obrigatório"
	}
	if form.PropertyID == "" {
		e.PropertyID = "Imóvel é obrigatório"
	}
	if form.Status == "" {
		e.Status = "Status é obrigatório"
	}
	return e
}

func hasProspectionErrors(e prospectionErrors) bool {
	return e.ClientID != "" || e.PropertyID != "" || e.Status != "" || e.Generic != ""
}

func prospectionToForm(p db.Prospection) prospectionForm {
	return prospectionForm{
		ClientID:       uuid.UUID(p.ClientID.Bytes).String(),
		PropertyID:     uuid.UUID(p.PropertyID.Bytes).String(),
		Status:         p.Status,
		Notes:          fromPgText(p.Notes),
		ContactDate:    fromPgTimestamp(p.ContactDate),
		NextActionDate: fromPgTimestamp(p.NextActionDate),
	}
}

func parseUUIDParam(s string) pgtype.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func parsePgTimestamp(s string) pgtype.Timestamptz {
	if s == "" {
		return pgtype.Timestamptz{Valid: false}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func fromPgTimestamp(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02")
}
