package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

// PropertyHandler handles property CRUD operations.
type PropertyHandler struct {
	queries *db.Queries
	tmpl    *template.Template
	log     *slog.Logger
}

// NewPropertyHandler creates a new PropertyHandler.
func NewPropertyHandler(queries *db.Queries, tmpl *template.Template, log *slog.Logger) *PropertyHandler {
	return &PropertyHandler{queries: queries, tmpl: tmpl, log: log}
}

// propertyPageData is the template data for property pages.
type propertyPageData struct {
	Title        string
	ActivePage   string
	UserEmail    string
	UserRole     string
	Properties   []db.Property
	Property     *db.Property
	IsEdit       bool
	Form         propertyForm
	Errors       propertyErrors
	Page         int
	PerPage      int
	Total        int64
	TotalPages   int
	Filters      propertyFilters
	Prospections []prospectWithNames
	HasProspects bool
}

type propertyForm struct {
	Title       string
	Address     string
	City        string
	State       string
	ZipCode     string
	Price       string
	Status      string
	Type        string
	Bedrooms    string
	Bathrooms   string
	AreaSqm     string
	Description string
	Photos      string // one URL per line
}

type propertyErrors struct {
	Title   string
	Address string
	City    string
	State   string
	Price   string
	Type    string
	Status  string
	Generic string
}

type propertyFilters struct {
	Status string
	Type   string
	Search string
}

// List shows a paginated, filtered list of properties.
func (h *PropertyHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	page, perPage := parsePagination(r, 20)
	filters := propertyFilters{
		Status: r.URL.Query().Get("status"),
		Type:   r.URL.Query().Get("type"),
		Search: r.URL.Query().Get("search"),
	}

	var statusVal, typeVal, searchVal string
	if filters.Status != "" {
		statusVal = filters.Status
	}
	if filters.Type != "" {
		typeVal = filters.Type
	}
	if filters.Search != "" {
		searchVal = filters.Search
	}

	props, err := h.queries.ListPropertiesFiltered(r.Context(), db.ListPropertiesFilteredParams{
		TenantID: user.TenantID,
		Column2:  statusVal,
		Column3:  typeVal,
		Column4:  searchVal,
		Limit:    int32(perPage),
		Offset:   int32((page - 1) * perPage),
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "property: list", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	total, err := h.queries.CountPropertiesFiltered(r.Context(), db.CountPropertiesFilteredParams{
		TenantID: user.TenantID,
		Column2:  statusVal,
		Column3:  typeVal,
		Column4:  searchVal,
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "property: count", "error", err)
		total = 0
	}

	data := propertyPageData{
		Title:      "Imóveis",
		ActivePage: "properties",
		UserEmail:  user.Email,
		UserRole:   user.Role,
		Properties: props,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
		Filters:    filters,
	}

	if err := h.tmpl.ExecuteTemplate(w, "properties_list.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "property: render list", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// New renders the property creation form.
func (h *PropertyHandler) New(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	data := propertyPageData{
		Title:      "Novo Imóvel",
		ActivePage: "properties",
		UserEmail:  user.Email,
		UserRole:   user.Role,
		IsEdit:     false,
		Form:       propertyForm{Status: "available", Type: "commercial"},
	}

	if err := h.tmpl.ExecuteTemplate(w, "properties_form.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "property: render form", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create handles property creation POST.
func (h *PropertyHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	form := parsePropertyForm(r)
	errors := validatePropertyForm(form)
	if hasPropertyErrors(errors) {
		h.renderFormError(w, r, user, form, errors, false)
		return
	}

	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	prop, err := h.queries.CreateProperty(r.Context(), db.CreatePropertyParams{
		ID:          id,
		TenantID:    user.TenantID,
		Title:       form.Title,
		Address:     form.Address,
		City:        form.City,
		State:       form.State,
		ZipCode:     toPgText(form.ZipCode),
		Price:       parseDecimal(form.Price),
		Status:      form.Status,
		Type:        form.Type,
		Bedrooms:    toPgInt32(form.Bedrooms),
		Bathrooms:   toPgInt32(form.Bathrooms),
		AreaSqm:     toPgNumeric(form.AreaSqm),
		Description: toPgText(form.Description),
		Photos:      parsePhotosJSON(form.Photos),
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "property: create", "error", err)
		errors.Generic = "Erro ao criar imóvel. Tente novamente."
		h.renderFormError(w, r, user, form, errors, false)
		return
	}

	h.log.InfoContext(r.Context(), "property_created", "id", uuid.UUID(prop.ID.Bytes).String(), "title", form.Title)
	http.Redirect(w, r, fmt.Sprintf("/properties/%s", uuid.UUID(prop.ID.Bytes).String()), http.StatusSeeOther)
}

// Detail shows a single property.
func (h *PropertyHandler) Detail(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := parseChiUUID(r, "id")
	prop, err := h.queries.GetPropertyByID(r.Context(), db.GetPropertyByIDParams{
		ID:       id,
		TenantID: user.TenantID,
	})
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Fetch linked prospections
	prospectRows, _ := h.queries.ListProspectsByProperty(r.Context(), db.ListProspectsByPropertyParams{
		PropertyID: id,
		TenantID:   user.TenantID,
	})

	prospections := make([]prospectWithNames, 0, len(prospectRows))
	for _, p := range prospectRows {
		client, _ := h.queries.GetClientByID(r.Context(), db.GetClientByIDParams{
			ID:       p.ClientID,
			TenantID: user.TenantID,
		})
		prospections = append(prospections, prospectWithNames{
			ID:             p.ID,
			Status:         p.Status,
			NextActionDate: p.NextActionDate,
			ClientName:     client.Name,
		})
	}

	data := propertyPageData{
		Title:        prop.Title,
		ActivePage:   "properties",
		UserEmail:    user.Email,
		UserRole:     user.Role,
		Property:     &prop,
		Prospections: prospections,
		HasProspects: len(prospections) > 0,
	}

	if err := h.tmpl.ExecuteTemplate(w, "properties_detail.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "property: render detail", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Edit renders the property edit form pre-filled with current values.
func (h *PropertyHandler) Edit(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := parseChiUUID(r, "id")
	prop, err := h.queries.GetPropertyByID(r.Context(), db.GetPropertyByIDParams{
		ID:       id,
		TenantID: user.TenantID,
	})
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data := propertyPageData{
		Title:      "Editar Imóvel",
		ActivePage: "properties",
		UserEmail:  user.Email,
		UserRole:   user.Role,
		IsEdit:     true,
		Property:   &prop,
		Form:       propertyToForm(prop),
	}

	if err := h.tmpl.ExecuteTemplate(w, "properties_form.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "property: render edit form", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Update handles property update POST.
func (h *PropertyHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	form := parsePropertyForm(r)
	errors := validatePropertyForm(form)
	if hasPropertyErrors(errors) {
		prop, _ := h.queries.GetPropertyByID(r.Context(), db.GetPropertyByIDParams{ID: id, TenantID: user.TenantID})
		h.renderFormError(w, r, user, form, errors, true)
		_ = prop
		return
	}

	prop, err := h.queries.UpdateProperty(r.Context(), db.UpdatePropertyParams{
		ID:          id,
		TenantID:    user.TenantID,
		Title:       form.Title,
		Address:     form.Address,
		City:        form.City,
		State:       form.State,
		ZipCode:     toPgText(form.ZipCode),
		Price:       parseDecimal(form.Price),
		Status:      form.Status,
		Type:        form.Type,
		Bedrooms:    toPgInt32(form.Bedrooms),
		Bathrooms:   toPgInt32(form.Bathrooms),
		AreaSqm:     toPgNumeric(form.AreaSqm),
		Description: toPgText(form.Description),
		Photos:      parsePhotosJSON(form.Photos),
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "property: update", "error", err)
		errors.Generic = "Erro ao atualizar imóvel."
		h.renderFormError(w, r, user, form, errors, true)
		return
	}

	h.log.InfoContext(r.Context(), "property_updated", "id", uuid.UUID(prop.ID.Bytes).String())
	http.Redirect(w, r, fmt.Sprintf("/properties/%s", uuid.UUID(prop.ID.Bytes).String()), http.StatusSeeOther)
}

// Delete soft-deletes a property.
func (h *PropertyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := parseChiUUID(r, "id")
	if err := h.queries.SoftDeleteProperty(r.Context(), db.SoftDeletePropertyParams{
		ID:       id,
		TenantID: user.TenantID,
	}); err != nil {
		h.log.ErrorContext(r.Context(), "property: delete", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.log.InfoContext(r.Context(), "property_deleted", "id", uuid.UUID(id.Bytes).String())
	http.Redirect(w, r, "/properties", http.StatusSeeOther)
}

func (h *PropertyHandler) renderFormError(w http.ResponseWriter, r *http.Request, user *db.User, form propertyForm, errors propertyErrors, isEdit bool) {
	data := propertyPageData{
		Title:      cond(isEdit, "Editar Imóvel", "Novo Imóvel"),
		ActivePage: "properties",
		UserEmail:  user.Email,
		UserRole:   user.Role,
		IsEdit:     isEdit,
		Form:       form,
		Errors:     errors,
	}
	if err := h.tmpl.ExecuteTemplate(w, "properties_form.html", data); err != nil {
		h.log.ErrorContext(r.Context(), "property: render form error", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func parsePropertyForm(r *http.Request) propertyForm {
	return propertyForm{
		Title:       strings.TrimSpace(r.FormValue("title")),
		Address:     strings.TrimSpace(r.FormValue("address")),
		City:        strings.TrimSpace(r.FormValue("city")),
		State:       strings.TrimSpace(r.FormValue("state")),
		ZipCode:     strings.TrimSpace(r.FormValue("zip_code")),
		Price:       strings.TrimSpace(r.FormValue("price")),
		Status:      strings.TrimSpace(r.FormValue("status")),
		Type:        strings.TrimSpace(r.FormValue("type")),
		Bedrooms:    strings.TrimSpace(r.FormValue("bedrooms")),
		Bathrooms:   strings.TrimSpace(r.FormValue("bathrooms")),
		AreaSqm:     strings.TrimSpace(r.FormValue("area_sqm")),
		Description: strings.TrimSpace(r.FormValue("description")),
		Photos:      strings.TrimSpace(r.FormValue("photos")),
	}
}

func validatePropertyForm(form propertyForm) propertyErrors {
	var e propertyErrors
	if len(form.Title) < 3 {
		e.Title = "Título deve ter pelo menos 3 caracteres"
	}
	if len(form.Address) < 5 {
		e.Address = "Endereço deve ter pelo menos 5 caracteres"
	}
	if len(form.City) < 2 {
		e.City = "Cidade deve ter pelo menos 2 caracteres"
	}
	if len(form.State) < 2 {
		e.State = "Estado deve ter pelo menos 2 caracteres"
	}
	if form.Price == "" {
		e.Price = "Preço é obrigatório"
	} else {
		priceVal := parseDecimal(form.Price)
		if !priceVal.Valid {
			e.Price = "Preço inválido"
		}
	}
	if form.Type == "" {
		e.Type = "Tipo é obrigatório"
	}
	if form.Status == "" {
		e.Status = "Status é obrigatório"
	}
	return e
}

func hasPropertyErrors(e propertyErrors) bool {
	return e.Title != "" || e.Address != "" || e.City != "" || e.State != "" || e.Price != "" || e.Type != "" || e.Status != "" || e.Generic != ""
}

func propertyToForm(prop db.Property) propertyForm {
	var photos string
	if len(prop.Photos) > 0 {
		var urls []string
		_ = json.Unmarshal(prop.Photos, &urls)
		photos = strings.Join(urls, "\n")
	}
	return propertyForm{
		Title:       prop.Title,
		Address:     prop.Address,
		City:        prop.City,
		State:       prop.State,
		ZipCode:     fromPgText(prop.ZipCode),
		Price:       formatDecimal(prop.Price),
		Status:      prop.Status,
		Type:        prop.Type,
		Bedrooms:    fromPgInt32(prop.Bedrooms),
		Bathrooms:   fromPgInt32(prop.Bathrooms),
		AreaSqm:     fromPgNumeric(prop.AreaSqm),
		Description: fromPgText(prop.Description),
		Photos:      photos,
	}
}

// parsePhotosJSON converts a textarea (one URL per line) to a JSON array.
func parsePhotosJSON(text string) []byte {
	if text == "" {
		return []byte("[]")
	}
	lines := strings.Split(text, "\n")
	var urls []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			urls = append(urls, line)
		}
	}
	data, _ := json.Marshal(urls)
	if data == nil {
		return []byte("[]")
	}
	return data
}

// parsePagination extracts page and per_page from query params.
func parsePagination(r *http.Request, defaultPerPage int) (page, perPage int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}

// parseChiUUID extracts a UUID from a chi URL param.
func parseChiUUID(r *http.Request, key string) pgtype.UUID {
	idStr := chi.URLParam(r, key)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

// prospectWithNames is a prospection with client/property names for display.
type prospectWithNames struct {
	ID             pgtype.UUID
	Status         string
	NextActionDate pgtype.Timestamptz
	ClientName     string
	PropertyName   string
}

func cond(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}
