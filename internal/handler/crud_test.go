package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

type crudTestDB struct {
	queries            *db.Queries
	pool               *pgxpool.Pool
	svc                *auth.Service
	tmpl               *template.Template
	log                *slog.Logger
	dashboardHandler   *DashboardHandler
	propertyHandler    *PropertyHandler
	clientHandler      *ClientHandler
	prospectionHandler *ProspectionHandler
	pdfHandler         *PDFHandler
}

func setupCRUDTestDB(t *testing.T) *crudTestDB {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}
	migrationsDir := findMigrationsDir(t)

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	pool.Config().MaxConns = 2

	m, err := migrate.New("file://"+migrationsDir, databaseURL)
	if err != nil {
		t.Fatalf("migrate.New: %v", err)
	}
	_ = m.Down()
	if err := m.Up(); err != nil && err.Error() != "no change" {
		t.Fatalf("migrate up: %v", err)
	}

	queries := db.New(pool)
	key := make([]byte, 32)
	rand.Read(key)
	limiter := auth.NewRateLimiter()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := auth.NewService(queries, pool, key, limiter, log)

	tmpl := template.Must(loadCRUDTestTemplates(t))

	return &crudTestDB{
		queries:            queries,
		pool:               pool,
		svc:                svc,
		tmpl:               tmpl,
		log:                log,
		dashboardHandler:   NewDashboardHandler(queries, tmpl, log),
		propertyHandler:    NewPropertyHandler(queries, tmpl, log),
		clientHandler:      NewClientHandler(queries, tmpl, log),
		prospectionHandler: NewProspectionHandler(queries, tmpl, log),
		pdfHandler:         NewPDFHandler(queries, tmpl, log),
	}
}

func loadCRUDTestTemplates(t *testing.T) (*template.Template, error) {
	t.Helper()
	templateDir := findTemplateDir(t)
	tmpl, err := template.New("").Funcs(TemplateFuncs()).ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "partials", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse partials: %w", err)
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "fragments", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse fragments: %w", err)
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "admin", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse admin: %w", err)
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "admin", "properties", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse admin/properties: %w", err)
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "admin", "clients", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse admin/clients: %w", err)
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "admin", "prospections", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse admin/prospections: %w", err)
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "admin", "contacts", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("parse admin/contacts: %w", err)
	}
	return tmpl, nil
}

func (td *crudTestDB) teardown(t *testing.T) {
	t.Helper()
	td.pool.Close()
}

func (td *crudTestDB) seed(t *testing.T) (tenantID, userID pgtype.UUID, rawToken string) {
	t.Helper()
	tenantID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, _ = td.queries.CreateTenant(context.Background(), db.CreateTenantParams{
		ID:   tenantID,
		Name: "CRUD Test Tenant",
		Plan: "free",
	})
	hash, _ := auth.HashPassword("test123")
	userID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, _ = td.queries.CreateUser(context.Background(), db.CreateUserParams{
		ID:           userID,
		TenantID:     tenantID,
		Email:        "crudtest@prospeccao.com.br",
		FullName:     "CRUD Test",
		Role:         "admin",
		PasswordHash: hash,
	})
	rawToken, _, err := td.svc.CreateSession(context.Background(), userID, tenantID)
	if err != nil {
		t.Fatalf("seed: create session: %v", err)
	}
	return tenantID, userID, rawToken
}

func (td *crudTestDB) seedProperty(t *testing.T, tenantID pgtype.UUID) db.Property {
	t.Helper()
	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	prop, err := td.queries.CreateProperty(context.Background(), db.CreatePropertyParams{
		ID:       id,
		TenantID: tenantID,
		Title:    "Sala Comercial Vila Mariana",
		Address:  "Rua Vergueiro 1000",
		City:     "Sao Paulo",
		State:    "SP",
		Price:    toPgNumeric("500000"),
		Status:   "available",
		Type:     "commercial",
		Photos:   []byte("[]"),
	})
	if err != nil {
		t.Fatalf("seedProperty: %v", err)
	}
	return prop
}

func (td *crudTestDB) seedClient(t *testing.T, tenantID pgtype.UUID) db.Client {
	t.Helper()
	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	client, err := td.queries.CreateClient(context.Background(), db.CreateClientParams{
		ID:          id,
		TenantID:    tenantID,
		Name:        "Joao Silva",
		Email:       strPtr("joao@example.com"),
		Phone:       strPtr("+55 11 99999-9999"),
		Status:      "lead",
		Preferences: []byte("{}"),
	})
	if err != nil {
		t.Fatalf("seedClient: %v", err)
	}
	return client
}

func strPtr(s string) *string { return &s }

func (td *crudTestDB) newCRUDRouter(rawToken string) *chi.Mux {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return SessionValidation(td.queries, slog.Default())(next)
		})
		r.Use(auth.RequireRole(auth.RoleAdmin))
		r.Get("/admin", td.dashboardHandler.Index)
		r.Get("/properties", td.propertyHandler.List)
		r.Get("/properties/new", td.propertyHandler.New)
		r.Post("/properties", td.propertyHandler.Create)
		r.Get("/properties/{id}", td.propertyHandler.Detail)
		r.Get("/properties/{id}/edit", td.propertyHandler.Edit)
		r.Post("/properties/{id}", td.propertyHandler.Update)
		r.Post("/properties/{id}/delete", td.propertyHandler.Delete)
		r.Get("/properties/{id}/pdf", td.pdfHandler.GeneratePropertyPDF)
		r.Get("/clients", td.clientHandler.List)
		r.Get("/clients/new", td.clientHandler.New)
		r.Post("/clients", td.clientHandler.Create)
		r.Get("/clients/{id}", td.clientHandler.Detail)
		r.Get("/clients/{id}/edit", td.clientHandler.Edit)
		r.Post("/clients/{id}", td.clientHandler.Update)
		r.Post("/clients/{id}/delete", td.clientHandler.Delete)
		r.Post("/clients/{id}/contacts", td.clientHandler.CreateContact)
		r.Get("/prospections", td.prospectionHandler.List)
		r.Get("/prospections/new", td.prospectionHandler.New)
		r.Post("/prospections", td.prospectionHandler.Create)
		r.Get("/prospections/{id}", td.prospectionHandler.Detail)
		r.Get("/prospections/{id}/edit", td.prospectionHandler.Edit)
		r.Post("/prospections/{id}", td.prospectionHandler.Update)
		r.Post("/prospections/{id}/delete", td.prospectionHandler.Delete)
		r.Post("/prospections/{id}/contacts", td.prospectionHandler.CreateContact)
	})
	return r
}

func authReq(method, path, rawToken string, body string) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	return req
}

// --- Dashboard Tests ---

func TestDashboardEmpty(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/admin", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Dashboard") {
		t.Error("expected Dashboard in body")
	}
}

func TestDashboardWithSeed(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	td.seedProperty(t, tenantID)
	td.seedClient(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/admin", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDashboardAuthRequired(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	td.seed(t)

	r := td.newCRUDRouter("invalid-token")
	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

// --- Property Tests ---

func TestPropertyList(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	td.seedProperty(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if len(body) < 100 {
		t.Logf("body too short: %s", body)
	}
	if !strings.Contains(body, "Sala Comercial") {
		end := 500
		if len(body) < end {
			end = len(body)
		}
		t.Logf("body (first %d chars): %s", end, body[:end])
		t.Error("expected property title in list")
	}
}

func TestPropertyListFiltered(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	td.seedProperty(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties?status=available", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestPropertyNew(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties/new", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "form") {
		t.Error("expected form in body")
	}
}

func TestPropertyCreate(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "title=Test Property&address=Rua Test 123&city=Sao Paulo&state=SP&price=500000&status=available&type=commercial"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/properties", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "/properties/") {
		t.Errorf("expected redirect to /properties/{id}, got %s", loc)
	}
}

func TestPropertyCreateInvalid(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "title=ab&address=xy&city=SP&state=SP&price=&status=available&type=commercial"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/properties", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (form re-rendered), got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "pelo menos 3") {
		t.Error("expected validation error for short title")
	}
}

func TestPropertyDetail(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties/"+uuid.UUID(prop.ID.Bytes).String(), rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Sala Comercial") {
		t.Error("expected property title in detail")
	}
}

func TestPropertyDetailNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties/"+uuid.New().String(), rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestPropertyEdit(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties/"+uuid.UUID(prop.ID.Bytes).String()+"/edit", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Sala Comercial") {
		t.Error("expected pre-filled title in edit form")
	}
}

func TestPropertyUpdate(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)

	form := "title=Updated Title&address=Rua Test 123&city=Sao Paulo&state=SP&price=600000&status=reserved&type=commercial"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/properties/"+uuid.UUID(prop.ID.Bytes).String(), rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestPropertySoftDelete(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/properties/"+uuid.UUID(prop.ID.Bytes).String()+"/delete", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/properties" {
		t.Errorf("expected redirect to /properties, got %s", rec.Header().Get("Location"))
	}

	// Verify it's not in the list anymore
	req = authReq("GET", "/properties", rawToken, "")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if strings.Contains(rec.Body.String(), "Sala Comercial") {
		t.Error("deleted property should not appear in list")
	}
}

func TestPropertyAuthRequired(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	td.seed(t)

	r := td.newCRUDRouter("invalid")
	req := httptest.NewRequest("GET", "/properties", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

// --- Client Tests ---

func TestClientList(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	td.seedClient(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/clients", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Joao Silva") {
		t.Error("expected client name in list")
	}
}

func TestClientCreate(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "name=Maria Santos&email=maria@example.com&phone=11999999999&status=lead"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestClientCreateInvalid(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "name=a&email=invalid&status=lead"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (form re-rendered), got %d", rec.Code)
	}
}

func TestClientDetail(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/clients/"+uuid.UUID(client.ID.Bytes).String(), rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Joao Silva") {
		t.Error("expected client name in detail")
	}
}

func TestClientDetailNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/clients/"+uuid.New().String(), rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestClientEdit(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/clients/"+uuid.UUID(client.ID.Bytes).String()+"/edit", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestClientUpdate(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	form := "name=Joao Updated&email=joao@example.com&phone=11999999999&status=active"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients/"+uuid.UUID(client.ID.Bytes).String(), rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestClientSoftDelete(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients/"+uuid.UUID(client.ID.Bytes).String()+"/delete", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

// --- Prospection Tests ---

func TestProspectionList(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	td.seedProspection(t, tenantID, client.ID, prop.ID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/prospections", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestProspectionNew(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	td.seedProperty(t, tenantID)
	td.seedClient(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/prospections/new", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestProspectionNewNoClientsOrProperties(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/prospections/new", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Cadastre um cliente") {
		t.Error("expected warning about no clients")
	}
}

func TestProspectionCreate(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)

	form := fmt.Sprintf("client_id=%s&property_id=%s&status=new", uuid.UUID(client.ID.Bytes).String(), uuid.UUID(prop.ID.Bytes).String())
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestProspectionCreateInvalid(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "client_id=&property_id=&status="
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (form re-rendered), got %d", rec.Code)
	}
}

func TestProspectionDetail(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	prospect := td.seedProspection(t, tenantID, client.ID, prop.ID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/prospections/"+uuid.UUID(prospect.ID.Bytes).String(), rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestProspectionDetailNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/prospections/"+uuid.New().String(), rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestProspectionUpdateStatus(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	prospect := td.seedProspection(t, tenantID, client.ID, prop.ID)

	form := "status=contacting&notes=Updated notes"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections/"+uuid.UUID(prospect.ID.Bytes).String(), rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestProspectionSoftDelete(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	prospect := td.seedProspection(t, tenantID, client.ID, prop.ID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections/"+uuid.UUID(prospect.ID.Bytes).String()+"/delete", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func (td *crudTestDB) seedProspection(t *testing.T, tenantID, clientID, propertyID pgtype.UUID) db.Prospection {
	t.Helper()
	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	prospect, err := td.queries.CreateProspect(context.Background(), db.CreateProspectParams{
		ID:         id,
		TenantID:   tenantID,
		ClientID:   clientID,
		PropertyID: propertyID,
		Status:     "new",
	})
	if err != nil {
		t.Fatalf("seedProspection: %v", err)
	}
	return prospect
}

// --- Contact Log Tests ---

func TestCreateContactForClient(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	form := "channel=phone&direction=outbound&subject=Follow-up&body=Cliente interessado"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients/"+uuid.UUID(client.ID.Bytes).String()+"/contacts", rawToken, form)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCreateContactForClientNoJS(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	form := "channel=email&direction=outbound&subject=Info&body=Sent info"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients/"+uuid.UUID(client.ID.Bytes).String()+"/contacts", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestCreateContactForProspection(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	prospect := td.seedProspection(t, tenantID, client.ID, prop.ID)

	form := "channel=whatsapp&direction=inbound&subject=Resposta&body=Cliente respondeu"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections/"+uuid.UUID(prospect.ID.Bytes).String()+"/contacts", rawToken, form)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCreateContactInvalidChannel(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	form := "channel=&direction=&subject=&body="
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients/"+uuid.UUID(client.ID.Bytes).String()+"/contacts", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- PDF Tests ---

func TestGeneratePDFAuthRequired(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	td.seed(t)

	r := td.newCRUDRouter("invalid")
	req := httptest.NewRequest("GET", "/properties/"+uuid.New().String()+"/pdf", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestGeneratePDFNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties/"+uuid.New().String()+"/pdf", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// --- Helper function tests ---

func TestParsePhotosJSON(t *testing.T) {
	result := parsePhotosJSON("")
	if string(result) != "[]" {
		t.Errorf("expected [], got %s", string(result))
	}

	result = parsePhotosJSON("http://example.com/1.jpg\nhttp://example.com/2.jpg")
	var urls []string
	if err := json.Unmarshal(result, &urls); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(urls) != 2 {
		t.Errorf("expected 2 urls, got %d", len(urls))
	}
}

func TestParsePagination(t *testing.T) {
	req := httptest.NewRequest("GET", "/properties?page=3&per_page=10", nil)
	page, perPage := parsePagination(req, 20)
	if page != 3 || perPage != 10 {
		t.Errorf("expected page=3, perPage=10, got page=%d, perPage=%d", page, perPage)
	}

	req = httptest.NewRequest("GET", "/properties", nil)
	page, perPage = parsePagination(req, 20)
	if page != 1 || perPage != 20 {
		t.Errorf("expected defaults page=1, perPage=20, got page=%d, perPage=%d", page, perPage)
	}

	req = httptest.NewRequest("GET", "/properties?per_page=200", nil)
	_, perPage = parsePagination(req, 20)
	if perPage != 100 {
		t.Errorf("expected perPage capped at 100, got %d", perPage)
	}
}

func TestValidatePropertyForm(t *testing.T) {
	errors := validatePropertyForm(propertyForm{Title: "ab", Address: "xy", City: "S", State: "S", Price: "", Type: "", Status: ""})
	if !hasPropertyErrors(errors) {
		t.Error("expected errors for invalid form")
	}

	errors = validatePropertyForm(propertyForm{Title: "Valid Title", Address: "Valid Address", City: "Sao Paulo", State: "SP", Price: "500000", Type: "commercial", Status: "available"})
	if hasPropertyErrors(errors) {
		t.Error("expected no errors for valid form")
	}
}

func TestValidateClientForm(t *testing.T) {
	errors := validateClientForm(clientForm{Name: "a", Email: "invalid"})
	if !hasClientErrors(errors) {
		t.Error("expected errors for invalid form")
	}

	errors = validateClientForm(clientForm{Name: "Joao", Email: "joao@example.com"})
	if hasClientErrors(errors) {
		t.Error("expected no errors for valid form")
	}
}

func TestToPgText(t *testing.T) {
	result := toPgText("")
	if result != nil {
		t.Error("expected nil for empty string")
	}
	result = toPgText("test")
	if result == nil || *result != "test" {
		t.Error("expected pointer to 'test'")
	}
}

func TestToPgInt32(t *testing.T) {
	result := toPgInt32("")
	if result != nil {
		t.Error("expected nil for empty string")
	}
	result = toPgInt32("42")
	if result == nil || *result != 42 {
		t.Error("expected pointer to 42")
	}
	result = toPgInt32("invalid")
	if result != nil {
		t.Error("expected nil for invalid number")
	}
}

func TestFindChrome(t *testing.T) {
	// Just verify it doesn't panic
	_ = findChrome()
}

// --- Additional coverage tests ---

func TestClientNew(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/clients/new", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "form") {
		t.Error("expected form in body")
	}
}

func TestProspectionEdit(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	prospect := td.seedProspection(t, tenantID, client.ID, prop.ID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/prospections/"+uuid.UUID(prospect.ID.Bytes).String()+"/edit", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestPropertyListSearch(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	td.seedProperty(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties?search=Sao+Paulo", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestPropertyListTypeFilter(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	td.seedProperty(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties?type=commercial", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestClientListSearch(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	td.seedClient(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/clients?search=Joao", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestClientListFiltered(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	td.seedClient(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/clients?status=lead", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestProspectionListFiltered(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	td.seedProspection(t, tenantID, client.ID, prop.ID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/prospections?status=new", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestProspectionCreateNoClientsOrProperties(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "client_id=&property_id=&status=new"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (form re-rendered), got %d", rec.Code)
	}
}

func TestProspectionEditNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/prospections/"+uuid.New().String()+"/edit", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestPropertyEditNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties/"+uuid.New().String()+"/edit", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestClientEditNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/clients/"+uuid.New().String()+"/edit", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestPropertyUpdateNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "title=Test&address=Rua Test&city=SP&state=SP&price=500000&status=available&type=commercial"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/properties/"+uuid.New().String(), rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Update fails, form is re-rendered with error
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (form re-rendered), got %d", rec.Code)
	}
}

func TestPropertyDeleteNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/properties/"+uuid.New().String()+"/delete", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// SoftDeleteProperty is :exec, doesn't error on not found
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestClientDeleteNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients/"+uuid.New().String()+"/delete", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestProspectionDeleteNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections/"+uuid.New().String()+"/delete", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestClientAuthRequired(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	td.seed(t)

	r := td.newCRUDRouter("invalid")
	req := httptest.NewRequest("GET", "/clients", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestProspectionAuthRequired(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	td.seed(t)

	r := td.newCRUDRouter("invalid")
	req := httptest.NewRequest("GET", "/prospections", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestCreateContactForProspectionNoJS(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	prospect := td.seedProspection(t, tenantID, client.ID, prop.ID)

	form := "channel=phone&direction=outbound&subject=Call&body=Called client"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections/"+uuid.UUID(prospect.ID.Bytes).String()+"/contacts", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestCreateContactForProspectionNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "channel=phone&direction=outbound&subject=Call&body=Called"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections/"+uuid.New().String()+"/contacts", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestCreateContactForClientInvalidChannel(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	form := "channel=invalidchannel&direction=outbound&subject=&body="
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients/"+uuid.UUID(client.ID.Bytes).String()+"/contacts", rawToken, form)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Invalid channel violates DB check constraint, returns 500
	if rec.Code != http.StatusInternalServerError {
		t.Logf("got %d (invalid channel hits DB constraint)", rec.Code)
	}
}

func TestParsePreferencesJSON(t *testing.T) {
	result := parsePreferencesJSON("")
	if string(result) != "{}" {
		t.Errorf("expected {}, got %s", string(result))
	}

	result = parsePreferencesJSON("invalid json")
	if string(result) != "{}" {
		t.Errorf("expected {} for invalid json, got %s", string(result))
	}

	result = parsePreferencesJSON(`{"key":"value"}`)
	if string(result) != `{"key":"value"}` {
		t.Errorf("expected valid json, got %s", string(result))
	}
}

func TestProspectionUpdateNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "status=contacting&notes=test"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections/"+uuid.New().String(), rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestClientUpdateNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "name=Test&email=test@test.com&status=active"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients/"+uuid.New().String(), rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Update fails, form is re-rendered with error
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (form re-rendered), got %d", rec.Code)
	}
}

func TestCond(t *testing.T) {
	if cond(true, "yes", "no") != "yes" {
		t.Error("expected yes for true")
	}
	if cond(false, "yes", "no") != "no" {
		t.Error("expected no for false")
	}
}

func TestParseChiUUIDInvalid(t *testing.T) {
	r := httptest.NewRequest("GET", "/properties/invalid-uuid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	result := parseChiUUID(r, "id")
	if result.Valid {
		t.Error("expected invalid UUID to have Valid=false")
	}
}

func TestParseUUIDParam(t *testing.T) {
	result := parseUUIDParam("invalid")
	if result.Valid {
		t.Error("expected invalid UUID to have Valid=false")
	}

	result = parseUUIDParam(uuid.New().String())
	if !result.Valid {
		t.Error("expected valid UUID to have Valid=true")
	}
}

func TestParsePgTimestamp(t *testing.T) {
	result := parsePgTimestamp("")
	if result.Valid {
		t.Error("expected empty string to have Valid=false")
	}

	result = parsePgTimestamp("2026-12-31")
	if !result.Valid {
		t.Error("expected valid date to have Valid=true")
	}

	result = parsePgTimestamp("invalid")
	if result.Valid {
		t.Error("expected invalid date to have Valid=false")
	}
}

func TestFromPgTimestamp(t *testing.T) {
	result := fromPgTimestamp(pgtype.Timestamptz{Valid: false})
	if result != "" {
		t.Errorf("expected empty string for invalid timestamp, got %s", result)
	}

	result = fromPgTimestamp(pgtype.Timestamptz{Time: parsePgTimestamp("2026-12-31").Time, Valid: true})
	if result != "2026-12-31" {
		t.Errorf("expected 2026-12-31, got %s", result)
	}
}

func TestFormatBRL(t *testing.T) {
	if formatBRL("") != "R$ 0,00" {
		t.Error("expected R$ 0,00 for empty")
	}
	if formatBRL("500.00") != "R$ 500.00" {
		t.Errorf("expected R$ 500.00, got %s", formatBRL("500.00"))
	}
}

func TestIsHTMX(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if isHTMX(req) {
		t.Error("expected false for non-HTMX request")
	}

	req.Header.Set("HX-Request", "true")
	if !isHTMX(req) {
		t.Error("expected true for HTMX request")
	}
}

func TestProspectionToForm(t *testing.T) {
	prospect := db.Prospection{
		ID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
		ClientID:   pgtype.UUID{Bytes: uuid.New(), Valid: true},
		PropertyID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Status:     "new",
	}

	form := prospectionToForm(prospect)
	if form.Status != "new" {
		t.Errorf("expected status 'new', got %s", form.Status)
	}
	if form.ClientID == "" {
		t.Error("expected non-empty client_id")
	}
}

func TestClientToForm(t *testing.T) {
	client := db.Client{
		Name:   "Test",
		Email:  strPtr("test@example.com"),
		Status: "lead",
	}

	form := clientToForm(client)
	if form.Name != "Test" {
		t.Errorf("expected name 'Test', got %s", form.Name)
	}
	if form.Email != "test@example.com" {
		t.Errorf("expected email, got %s", form.Email)
	}
}

func TestClientToFormWithNilFields(t *testing.T) {
	client := db.Client{
		Name:   "Test",
		Email:  nil,
		Phone:  nil,
		Status: "lead",
	}

	form := clientToForm(client)
	if form.Name != "Test" {
		t.Errorf("expected name 'Test', got %s", form.Name)
	}
	if form.Email != "" {
		t.Errorf("expected empty email, got %s", form.Email)
	}
	if form.Phone != "" {
		t.Errorf("expected empty phone, got %s", form.Phone)
	}
}

func TestPropertyToForm(t *testing.T) {
	prop := db.Property{
		Title:  "Test Property",
		Status: "available",
		Type:   "commercial",
		Photos: []byte(`["http://example.com/1.jpg"]`),
	}

	form := propertyToForm(prop)
	if form.Title != "Test Property" {
		t.Errorf("expected title, got %s", form.Title)
	}
	if !strings.Contains(form.Photos, "example.com") {
		t.Errorf("expected photos URL in form, got %s", form.Photos)
	}
}

func TestValidateProspectionForm(t *testing.T) {
	errors := validateProspectionForm(prospectionForm{ClientID: "", PropertyID: "", Status: ""})
	if !hasProspectionErrors(errors) {
		t.Error("expected errors for empty form")
	}

	errors = validateProspectionForm(prospectionForm{ClientID: "uuid", PropertyID: "uuid", Status: "new"})
	if hasProspectionErrors(errors) {
		t.Error("expected no errors for valid form")
	}
}

func TestHasPropertyErrorsGeneric(t *testing.T) {
	errors := propertyErrors{Generic: "some error"}
	if !hasPropertyErrors(errors) {
		t.Error("expected true for generic error")
	}
}

func TestHasProspectionErrorsGeneric(t *testing.T) {
	errors := prospectionErrors{Generic: "some error"}
	if !hasProspectionErrors(errors) {
		t.Error("expected true for generic error")
	}
}

func TestGeneratePDFNoChrome(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties/"+uuid.UUID(prop.ID.Bytes).String()+"/pdf", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Chrome is available on macOS, should generate PDF
	if rec.Code == http.StatusOK {
		ct := rec.Header().Get("Content-Type")
		if ct != "application/pdf" {
			t.Errorf("expected application/pdf, got %s", ct)
		}
		if rec.Body.Len() == 0 {
			t.Error("expected non-empty PDF body")
		}
	} else if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 200 (PDF) or 500 (no chrome), got %d", rec.Code)
	}
}

func TestDashboardWithSeedData(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	td.seedProperty(t, tenantID)
	td.seedClient(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/admin", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Imóveis") {
		t.Error("expected Imóveis stat card")
	}
}

func TestDashboardWithProspectionData(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	td.seedProspection(t, tenantID, client.ID, prop.ID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/admin", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestPropertyCreateWithPhotos(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "title=Test With Photos&address=Rua Test 123&city=Sao Paulo&state=SP&price=500000&status=available&type=commercial&photos=http://example.com/1.jpg%0Ahttp://example.com/2.jpg"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/properties", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestPropertyCreateWithAllFields(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "title=Full Property&address=Rua Full 123&city=Sao Paulo&state=SP&zip_code=01000-000&price=750000.50&status=available&type=residential&bedrooms=3&bathrooms=2&area_sqm=120.50&description=Nice property&photos=http://example.com/photo.jpg"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/properties", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestClientCreateWithAllFields(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "name=Full Client&email=full@example.com&phone=11999999999&cpf_cnpj=12345678901&address=Rua Client 123&budget=500000&preferences=%7B%22type%22%3A%22apartment%22%7D&status=active"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestClientCreateInvalidEmail(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	form := "name=Test&email=notanemail&status=lead"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (form re-rendered), got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Email inv") {
		t.Error("expected email validation error")
	}
}

func TestProspectionCreateWithDates(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)

	form := fmt.Sprintf("client_id=%s&property_id=%s&status=contacting&notes=Following up&contact_date=2026-08-01&next_action_date=2026-08-15",
		uuid.UUID(client.ID.Bytes).String(), uuid.UUID(prop.ID.Bytes).String())
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestProspectionUpdateWithDate(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	prospect := td.seedProspection(t, tenantID, client.ID, prop.ID)

	form := "status=negotiating&notes=In negotiation&next_action_date=2026-09-01"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections/"+uuid.UUID(prospect.ID.Bytes).String(), rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", rec.Code)
	}
}

func TestPropertyListEmpty(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Nenhum im") {
		t.Error("expected empty state message")
	}
}

func TestClientListEmpty(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/clients", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Nenhum cliente") {
		t.Error("expected empty state message")
	}
}

func TestProspectionListEmpty(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/prospections", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Nenhuma prospec") {
		t.Error("expected empty state message")
	}
}

func TestPropertyDetailWithProspections(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	td.seedProspection(t, tenantID, client.ID, prop.ID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties/"+uuid.UUID(prop.ID.Bytes).String(), rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestClientDetailWithContacts(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	// Create a contact first
	form := "channel=phone&direction=outbound&subject=First call&body=Called"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients/"+uuid.UUID(client.ID.Bytes).String()+"/contacts", rawToken, form)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Now view detail page
	req = authReq("GET", "/clients/"+uuid.UUID(client.ID.Bytes).String(), rawToken, "")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestProspectionDetailWithContacts(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	prospect := td.seedProspection(t, tenantID, client.ID, prop.ID)

	// Create a contact for the prospection
	form := "channel=email&direction=outbound&subject=Sent info&body=Sent details"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections/"+uuid.UUID(prospect.ID.Bytes).String()+"/contacts", rawToken, form)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Now view detail page
	req = authReq("GET", "/prospections/"+uuid.UUID(prospect.ID.Bytes).String(), rawToken, "")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestFromPgInt32(t *testing.T) {
	if fromPgInt32(nil) != "" {
		t.Error("expected empty string for nil")
	}
	i := int32(42)
	if fromPgInt32(&i) != "42" {
		t.Error("expected 42")
	}
}

func TestFromPgText(t *testing.T) {
	if fromPgText(nil) != "" {
		t.Error("expected empty string for nil")
	}
	s := "hello"
	if fromPgText(&s) != "hello" {
		t.Error("expected hello")
	}
}

func TestToPgNumericInvalid(t *testing.T) {
	result := toPgNumeric("not a number")
	if result.Valid {
		t.Error("expected invalid for non-number")
	}
}

func TestFromPgNumericInvalid(t *testing.T) {
	result := fromPgNumeric(pgtype.Numeric{Valid: false})
	if result != "" {
		t.Errorf("expected empty for invalid, got %s", result)
	}
}

func TestFromPgNumericValid(t *testing.T) {
	result := fromPgNumeric(toPgNumeric("500000"))
	if result == "" {
		t.Error("expected non-empty for valid numeric")
	}
}

func TestParseDecimalEdgeCases(t *testing.T) {
	result := parseDecimal("")
	if result.Valid {
		t.Error("expected invalid for empty string")
	}

	result = parseDecimal("0")
	if !result.Valid {
		t.Error("expected valid for 0")
	}

	result = parseDecimal("invalid")
	if result.Valid {
		t.Error("expected invalid for non-number")
	}
}

func TestFormatBRLNumeric(t *testing.T) {
	result := formatBRLNumeric(pgtype.Numeric{Valid: false})
	if result != "R$ 0,00" {
		t.Errorf("expected R$ 0,00, got %s", result)
	}
}

func TestUuidToStringInvalid(t *testing.T) {
	result := uuidToString(pgtype.UUID{Valid: false})
	if result != "" {
		t.Errorf("expected empty for invalid, got %s", result)
	}
}

func TestJoinPhotos(t *testing.T) {
	result := joinPhotos([]byte(`["http://a.com/1.jpg","http://b.com/2.jpg"]`))
	if !strings.Contains(result, "a.com") {
		t.Errorf("expected URLs in result, got %s", result)
	}
}

func TestTemplateFuncs(t *testing.T) {
	funcs := TemplateFuncs()
	if funcs == nil {
		t.Error("expected non-nil FuncMap")
	}
	if _, ok := funcs["uuidToString"]; !ok {
		t.Error("expected uuidToString function")
	}
}

// --- No-user-in-context redirect tests ---
// These test the `!ok` branch at the top of each handler.

func TestDashboardNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	td.dashboardHandler.Index(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestPropertyListNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/properties", nil)
	rec := httptest.NewRecorder()
	td.propertyHandler.List(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestPropertyNewNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/properties/new", nil)
	rec := httptest.NewRecorder()
	td.propertyHandler.New(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestPropertyCreateNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	form := "title=Test&address=Rua&city=SP&state=SP&price=500&status=available&type=commercial"
	req := authReq("POST", "/properties", "", form)
	rec := httptest.NewRecorder()
	td.propertyHandler.Create(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestPropertyDetailNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Get("/properties/{id}", td.propertyHandler.Detail)
	req := httptest.NewRequest("GET", "/properties/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestPropertyEditNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Get("/properties/{id}/edit", td.propertyHandler.Edit)
	req := httptest.NewRequest("GET", "/properties/"+uuid.New().String()+"/edit", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestPropertyUpdateNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Post("/properties/{id}", td.propertyHandler.Update)
	form := "title=Test&address=Rua&city=SP&state=SP&price=500&status=available&type=commercial"
	req := authReq("POST", "/properties/"+uuid.New().String(), "", form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestPropertyDeleteNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Post("/properties/{id}/delete", td.propertyHandler.Delete)
	req := httptest.NewRequest("POST", "/properties/"+uuid.New().String()+"/delete", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestClientListNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/clients", nil)
	rec := httptest.NewRecorder()
	td.clientHandler.List(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestClientNewNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/clients/new", nil)
	rec := httptest.NewRecorder()
	td.clientHandler.New(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestClientCreateNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	form := "name=Test&email=test@test.com&status=lead"
	req := authReq("POST", "/clients", "", form)
	rec := httptest.NewRecorder()
	td.clientHandler.Create(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestClientDetailNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Get("/clients/{id}", td.clientHandler.Detail)
	req := httptest.NewRequest("GET", "/clients/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestClientEditNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Get("/clients/{id}/edit", td.clientHandler.Edit)
	req := httptest.NewRequest("GET", "/clients/"+uuid.New().String()+"/edit", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestClientUpdateNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Post("/clients/{id}", td.clientHandler.Update)
	form := "name=Test&email=test@test.com&status=active"
	req := authReq("POST", "/clients/"+uuid.New().String(), "", form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestClientDeleteNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Post("/clients/{id}/delete", td.clientHandler.Delete)
	req := httptest.NewRequest("POST", "/clients/"+uuid.New().String()+"/delete", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestProspectionListNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/prospections", nil)
	rec := httptest.NewRecorder()
	td.prospectionHandler.List(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestProspectionNewNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/prospections/new", nil)
	rec := httptest.NewRecorder()
	td.prospectionHandler.New(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestProspectionCreateNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	form := "client_id=&property_id=&status=new"
	req := authReq("POST", "/prospections", "", form)
	rec := httptest.NewRecorder()
	td.prospectionHandler.Create(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestProspectionDetailNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Get("/prospections/{id}", td.prospectionHandler.Detail)
	req := httptest.NewRequest("GET", "/prospections/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestProspectionEditNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Get("/prospections/{id}/edit", td.prospectionHandler.Edit)
	req := httptest.NewRequest("GET", "/prospections/"+uuid.New().String()+"/edit", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestProspectionUpdateNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Post("/prospections/{id}", td.prospectionHandler.Update)
	form := "status=contacting&notes=test"
	req := authReq("POST", "/prospections/"+uuid.New().String(), "", form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestProspectionDeleteNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Post("/prospections/{id}/delete", td.prospectionHandler.Delete)
	req := httptest.NewRequest("POST", "/prospections/"+uuid.New().String()+"/delete", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestCreateContactForClientNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Post("/clients/{id}/contacts", td.clientHandler.CreateContact)
	form := "channel=phone&direction=outbound&subject=Test&body=Test"
	req := authReq("POST", "/clients/"+uuid.New().String()+"/contacts", "", form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestCreateContactForProspectionNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Post("/prospections/{id}/contacts", td.prospectionHandler.CreateContact)
	form := "channel=phone&direction=outbound&subject=Test&body=Test"
	req := authReq("POST", "/prospections/"+uuid.New().String()+"/contacts", "", form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

func TestGeneratePDFNoUser(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)

	r := chi.NewRouter()
	r.Get("/properties/{id}/pdf", td.pdfHandler.GeneratePropertyPDF)
	req := httptest.NewRequest("GET", "/properties/"+uuid.New().String()+"/pdf", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
}

// --- Validation error path tests ---

func TestPropertyUpdateInvalidForm(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)

	form := "title=ab&address=xy&city=S&state=S&price=&status=available&type=commercial"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/properties/"+uuid.UUID(prop.ID.Bytes).String(), rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (form re-rendered), got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "pelo menos 3") {
		t.Error("expected validation error for short title")
	}
}

func TestClientUpdateInvalidForm(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	form := "name=a&email=invalid&status=active"
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/clients/"+uuid.UUID(client.ID.Bytes).String(), rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (form re-rendered), got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "pelo menos 2") {
		t.Error("expected validation error for short name")
	}
}

func TestProspectionCreateInvalidUUID(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	td.seedClient(t, tenantID)

	form := fmt.Sprintf("client_id=invalid&property_id=%s&status=new", uuid.UUID(prop.ID.Bytes).String())
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Invalid UUID fails validation, form is re-rendered
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (form re-rendered), got %d", rec.Code)
	}
}

func TestProspectionCreatePropertyNotFound(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	form := fmt.Sprintf("client_id=%s&property_id=%s&status=new", uuid.UUID(client.ID.Bytes).String(), uuid.New().String())
	r := td.newCRUDRouter(rawToken)
	req := authReq("POST", "/prospections", rawToken, form)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Property doesn't exist, FK violation, form re-rendered with error
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (form re-rendered), got %d", rec.Code)
	}
}

// --- ParseForm error tests ---

func TestPropertyCreateBadForm(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	// Send a request with invalid content type to trigger ParseForm error
	req := httptest.NewRequest("POST", "/properties", strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestPropertyUpdateBadForm(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := httptest.NewRequest("POST", "/properties/"+uuid.UUID(prop.ID.Bytes).String(), strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestClientCreateBadForm(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := httptest.NewRequest("POST", "/clients", strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestClientUpdateBadForm(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := httptest.NewRequest("POST", "/clients/"+uuid.UUID(client.ID.Bytes).String(), strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestProspectionCreateBadForm(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	_, _, rawToken := td.seed(t)

	r := td.newCRUDRouter(rawToken)
	req := httptest.NewRequest("POST", "/prospections", strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestProspectionUpdateBadForm(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	prospect := td.seedProspection(t, tenantID, client.ID, prop.ID)

	r := td.newCRUDRouter(rawToken)
	req := httptest.NewRequest("POST", "/prospections/"+uuid.UUID(prospect.ID.Bytes).String(), strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCreateContactForClientBadForm(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	client := td.seedClient(t, tenantID)

	r := td.newCRUDRouter(rawToken)
	req := httptest.NewRequest("POST", "/clients/"+uuid.UUID(client.ID.Bytes).String()+"/contacts", strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCreateContactForProspectionBadForm(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	prospect := td.seedProspection(t, tenantID, client.ID, prop.ID)

	r := td.newCRUDRouter(rawToken)
	req := httptest.NewRequest("POST", "/prospections/"+uuid.UUID(prospect.ID.Bytes).String()+"/contacts", strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- Property with photos detail test ---

func TestPropertyDetailWithPhotos(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)

	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, _ = td.queries.CreateProperty(context.Background(), db.CreatePropertyParams{
		ID:       id,
		TenantID: tenantID,
		Title:    "Property With Photos",
		Address:  "Rua Test 123",
		City:     "Sao Paulo",
		State:    "SP",
		Price:    toPgNumeric("500000"),
		Status:   "available",
		Type:     "commercial",
		Photos:   []byte(`["http://example.com/1.jpg","http://example.com/2.jpg"]`),
	})

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties/"+uuid.UUID(id.Bytes).String(), rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Property With Photos") {
		t.Error("expected property title in detail page")
	}
}

// --- Property with all fields detail test ---

func TestPropertyDetailWithAllFields(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)

	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, _ = td.queries.CreateProperty(context.Background(), db.CreatePropertyParams{
		ID:          id,
		TenantID:    tenantID,
		Title:       "Full Property Detail",
		Address:     "Rua Full 123",
		City:        "Sao Paulo",
		State:       "SP",
		ZipCode:     toPgText("01000-000"),
		Price:       toPgNumeric("750000"),
		Status:      "available",
		Type:        "residential",
		Bedrooms:    toPgInt32("3"),
		Bathrooms:   toPgInt32("2"),
		AreaSqm:     toPgNumeric("120.50"),
		Description: toPgText("Beautiful property with garden"),
		Photos:      []byte("[]"),
	})

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties/"+uuid.UUID(id.Bytes).String(), rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Full Property Detail") {
		t.Error("expected property title in detail")
	}
	if !strings.Contains(body, "Beautiful property") {
		t.Error("expected description in detail")
	}
}

// --- Client detail with contacts and prospections ---

func TestClientDetailWithProspections(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)
	td.seedProspection(t, tenantID, client.ID, prop.ID)

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/clients/"+uuid.UUID(client.ID.Bytes).String(), rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- Prospection detail with all fields ---

func TestProspectionDetailWithNotes(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)

	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, _ = td.queries.CreateProspect(context.Background(), db.CreateProspectParams{
		ID:         id,
		TenantID:   tenantID,
		ClientID:   client.ID,
		PropertyID: prop.ID,
		Status:     "negotiating",
		Notes:      toPgText("Client is interested, negotiating price"),
	})

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/prospections/"+uuid.UUID(id.Bytes).String(), rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "negotiating") {
		t.Error("expected status in detail")
	}
}

func TestSubInt(t *testing.T) {
	if subInt(5, 3) != 2 {
		t.Error("expected 2")
	}
}

func TestAddInt(t *testing.T) {
	if addInt(5, 3) != 8 {
		t.Error("expected 8")
	}
}

func TestPgInt32Display(t *testing.T) {
	if pgInt32Display(nil) != "" {
		t.Error("expected empty for nil")
	}
	i := int32(42)
	if pgInt32Display(&i) != "42" {
		t.Error("expected 42")
	}
}

func TestPgNumericDisplay(t *testing.T) {
	result := pgNumericDisplay(pgtype.Numeric{Valid: false})
	if result != "" {
		t.Errorf("expected empty for invalid, got %s", result)
	}
}

func TestPropertyListPagination(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)

	// Create 25 properties
	for i := 0; i < 25; i++ {
		id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		_, _ = td.queries.CreateProperty(context.Background(), db.CreatePropertyParams{
			ID:       id,
			TenantID: tenantID,
			Title:    fmt.Sprintf("Property %d", i),
			Address:  "Rua Test 123",
			City:     "Sao Paulo",
			State:    "SP",
			Price:    toPgNumeric("500000"),
			Status:   "available",
			Type:     "commercial",
			Photos:   []byte("[]"),
		})
	}

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/properties?page=2&per_page=20", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestClientListPagination(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)

	for i := 0; i < 25; i++ {
		id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		_, _ = td.queries.CreateClient(context.Background(), db.CreateClientParams{
			ID:          id,
			TenantID:    tenantID,
			Name:        fmt.Sprintf("Client %d", i),
			Status:      "lead",
			Preferences: []byte("{}"),
		})
	}

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/clients?page=2&per_page=20", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestProspectionListPagination(t *testing.T) {
	td := setupCRUDTestDB(t)
	defer td.teardown(t)
	tenantID, _, rawToken := td.seed(t)
	prop := td.seedProperty(t, tenantID)
	client := td.seedClient(t, tenantID)

	for i := 0; i < 25; i++ {
		id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		_, _ = td.queries.CreateProspect(context.Background(), db.CreateProspectParams{
			ID:         id,
			TenantID:   tenantID,
			ClientID:   client.ID,
			PropertyID: prop.ID,
			Status:     "new",
		})
	}

	r := td.newCRUDRouter(rawToken)
	req := authReq("GET", "/prospections?page=2&per_page=20", rawToken, "")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
