package handler

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

type instTestDB struct {
	queries        *db.Queries
	pool           *pgxpool.Pool
	instHandler    *InstitutionalHandler
	contactHandler *ContactHandler
	newsHandler    *NewsletterHandler
	tmpl           *template.Template
	limiter        *auth.RateLimiter
}

func setupInstTestDB(t *testing.T) *instTestDB {
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

	m, err := migrate.New("file://"+migrationsDir, databaseURL)
	if err != nil {
		t.Fatalf("migrate.New: %v", err)
	}
	_ = m.Down()
	if err := m.Up(); err != nil && err.Error() != "no change" {
		t.Fatalf("migrate up: %v", err)
	}
	defer func() { _, _ = m.Close() }()

	queries := db.New(pool)
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	limiter := auth.NewRateLimiter()

	tmpl := template.Must(loadTestTemplates(t))
	instHandler := NewInstitutionalHandler(queries, tmpl, log)
	contactHandler := NewContactHandler(queries, tmpl, log, limiter)
	newsHandler := NewNewsletterHandler(queries, tmpl, log, limiter)

	// Truncate tables for test isolation
	_, _ = pool.Exec(context.Background(), "TRUNCATE contact_submissions, newsletter_subscribers CASCADE")

	return &instTestDB{
		queries:        queries,
		pool:           pool,
		instHandler:    instHandler,
		contactHandler: contactHandler,
		newsHandler:    newsHandler,
		tmpl:           tmpl,
		limiter:        limiter,
	}
}

func (td *instTestDB) teardown(t *testing.T) {
	t.Helper()
	td.pool.Close()
}

func loadTestTemplates(t *testing.T) (*template.Template, error) {
	t.Helper()
	templateDir := findTemplateDir(t)
	tmpl, err := template.New("").Funcs(TemplateFuncs()).ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "partials", "*.html"))
	if err != nil {
		return nil, err
	}
	_, err = tmpl.ParseGlob(filepath.Join(templateDir, "fragments", "*.html"))
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func newInstRouter(td *instTestDB) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", td.instHandler.Home)
	r.Get("/quem-somos", td.instHandler.QuemSomos)
	r.Get("/servicos", td.instHandler.Servicos)
	r.Get("/servicos/{slug}", td.instHandler.ServicoDetalhe)
	r.Get("/nossos-clientes", td.instHandler.NossosClientes)
	r.Get("/segmentos", td.instHandler.Segmentos)
	r.Get("/equipe", td.instHandler.Equipe)
	r.Get("/parceiros", td.instHandler.Parceiros)
	r.Get("/responsabilidade-social", td.instHandler.ResponsabilidadeSocial)
	r.Get("/investidores", td.instHandler.Investidores)
	r.Get("/fale-conosco", td.instHandler.FaleConosco)
	r.Post("/fale-conosco", td.contactHandler.Submit)
	r.Post("/newsletter", td.newsHandler.Subscribe)
	r.NotFound(td.instHandler.NotFound)
	return r
}

func TestHomeGET(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Prospecção Brasil") {
		t.Error("expected 'Prospecção Brasil' in body")
	}
	if !strings.Contains(body, "Comercial") {
		t.Error("expected market copy 'Comercial' in body")
	}
	if strings.Contains(body, "carga cognitiva") {
		t.Error("forbidden copy 'carga cognitiva' found in home")
	}
	if strings.Contains(body, "pipeline") {
		t.Error("forbidden copy 'pipeline' found in home")
	}
	if !strings.Contains(body, "Servi") {
		t.Error("expected services section in body")
	}
	if !strings.Contains(body, "Soluções Estratégicas") {
		t.Error("expected services preview heading 'Soluções Estratégicas'")
	}
	// Verify metrics labels present (hero inline metrics)
	if !strings.Contains(body, "Pontos") {
		t.Error("expected metric label 'Pontos' in hero")
	}
	// Verify at least one hardcoded testimonial name present
	if !strings.Contains(body, "Ricardo Santos") {
		t.Error("expected testimonial from 'Ricardo Santos'")
	}
}

func TestQuemSomosGET(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/quem-somos", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Quem Somos") {
		t.Error("expected 'Quem Somos' in body")
	}
	if !strings.Contains(body, "Luiz Cl") {
		t.Error("expected founder 'Luiz Cl'")
	}
	if !strings.Contains(body, "Shell") {
		t.Error("expected 'Shell' in founder bio")
	}
	if !strings.Contains(body, "Missão") {
		t.Error("expected 'Missão' section")
	}
	if !strings.Contains(body, "CRECI") {
		t.Error("expected 'CRECI' mention")
	}
	if strings.Contains(body, "carga cognitiva") {
		t.Error("forbidden copy 'carga cognitiva' found in quem-somos")
	}
	if strings.Contains(body, "plataforma") {
		t.Error("forbidden copy 'plataforma' found in quem-somos")
	}
}

func TestServicosGET(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/servicos", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Nossos") || !strings.Contains(body, "Servi") {
		t.Error("expected 'Nossos Serviços' in body")
	}
	// Check for at least 5 service cards (we have 5 services)
	count := strings.Count(body, "Saiba mais")
	if count < 5 {
		t.Errorf("expected at least 5 service cards, got %d 'Saiba mais' links", count)
	}
	// Verify no software copy
	if strings.Contains(body, "carga cognitiva") {
		t.Error("forbidden copy 'carga cognitiva' found in servicos")
	}
}

func TestServicoDetalheGET(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/servicos/expansao-de-redes", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Expansão de Redes") {
		t.Error("expected service title 'Expansão de Redes'")
	}
	if !strings.Contains(body, "Como") || !strings.Contains(body, "fazemos") {
		t.Error("expected methodology section 'Como fazemos'")
	}
}

func TestServicoDetalheNotFound(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/servicos/inexistente", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestServicoDetalheAll(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	slugs := []string{
		"expansao-de-redes",
		"built-to-suit",
		"strip-mall",
		"prospeccao-de-ponto",
		"conselho-consultivo",
		"sale-leaseback",
		"transferencia-de-pontos",
		"ingresso-em-mercados",
		"administracao-de-portfolios",
		"estruturacao-de-contratos",
		"avaliacao-estrategica",
	}
	for _, slug := range slugs {
		t.Run(slug, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/servicos/"+slug, nil)
			rr := httptest.NewRecorder()
			newInstRouter(td).ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected 200 for %s, got %d", slug, rr.Code)
			}
		})
	}
}

func TestSegmentosGET(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/segmentos", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Segmentos") {
		t.Error("expected 'Segmentos' in body")
	}
	if !strings.Contains(body, "Fast Food") {
		t.Error("expected segment 'Fast Food' in body")
	}
	if !strings.Contains(body, "Farmácias") {
		t.Error("expected segment 'Farmácias' in body")
	}
}

func TestEquipeGET(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/equipe", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Equipe") {
		t.Error("expected 'Equipe' in body")
	}
	if !strings.Contains(body, "Luiz Claudio") {
		t.Error("expected team member 'Luiz Claudio' in body")
	}
}

func TestParceirosGET(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/parceiros", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Parceiras") {
		t.Error("expected 'Parceiras' in body")
	}
	if !strings.Contains(body, "Burger King") {
		t.Error("expected client logo 'Burger King' in body")
	}
	if !strings.Contains(body, "burger-king.png") {
		t.Error("expected client logo path 'burger-king.png' in body")
	}
	if !strings.Contains(body, "Resultados") {
		t.Error("expected 'Resultados' section in body")
	}
	if !strings.Contains(body, "rihappy.png") {
		t.Error("expected result image 'rihappy.png' in body")
	}
}

func TestResponsabilidadeSocialGET(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/responsabilidade-social", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Responsabilidade") {
		t.Error("expected 'Responsabilidade' in body")
	}
	if !strings.Contains(body, "Gerando Falcões") {
		t.Error("expected social cause 'Gerando Falcões' in body")
	}
}

func TestInvestidoresGET(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/investidores", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Investidores") {
		t.Error("expected 'Investidores' in body")
	}
	if !strings.Contains(body, "Portfólios") {
		t.Error("expected investor service 'Portfólios' in body")
	}
}

func TestNossosClientesGET(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/nossos-clientes", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	// Should show testimonials, not empty state
	if strings.Contains(body, "Em breve") {
		t.Error("should not show empty state 'Em breve' -- testimonials are static")
	}
	if !strings.Contains(body, "Larissa Mello") && !strings.Contains(body, "Roberto Andrade") {
		t.Error("expected at least one testimonial name in body")
	}
	if !strings.Contains(body, "Pontos Comercializados") {
		t.Error("expected metrics strip on nossos-clientes")
	}
}

func TestFaleConoscoGET(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/fale-conosco", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Fale Conosco") {
		t.Error("expected 'Fale Conosco' in body")
	}
	if !strings.Contains(body, `name="name"`) {
		t.Error("expected name field in form")
	}
	if !strings.Contains(body, `name="email"`) {
		t.Error("expected email field in form")
	}
	if !strings.Contains(body, `name="message"`) {
		t.Error("expected message field in form")
	}
	if !strings.Contains(body, `name="company"`) {
		t.Error("expected company field in form")
	}
	if !strings.Contains(body, "Ipanema") {
		t.Error("expected contact info 'Ipanema' on page")
	}
}

func TestNotFound(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "não encontrada") {
		t.Error("expected 'não encontrada' in 404 body")
	}
	if !strings.Contains(body, "Voltar ao Início") {
		t.Error("expected link back to home in 404 body")
	}
}

func TestNavActiveState(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	tests := []struct {
		path       string
		activePage string
	}{
		{"/", "home"},
		{"/quem-somos", "quem-somos"},
		{"/servicos", "servicos"},
		{"/nossos-clientes", "nossos-clientes"},
	}

	for _, tt := range tests {
		t.Run(tt.activePage, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()
			newInstRouter(td).ServeHTTP(rr, req)

			body := rr.Body.String()
			if !strings.Contains(body, "active") {
				t.Errorf("expected 'active' class on nav for %s", tt.path)
			}
		})
	}

	// Fale Conosco is a button (btn-primary), not a nav-link, so we check for its presence
	t.Run("fale-conosco", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/fale-conosco", nil)
		rr := httptest.NewRecorder()
		newInstRouter(td).ServeHTTP(rr, req)

		body := rr.Body.String()
		if !strings.Contains(body, "Fale Conosco") {
			t.Error("expected 'Fale Conosco' button in nav")
		}
	})
}

func TestContactSubmitValid(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{
		"company": {"Empresa Teste Ltda"},
		"name":    {"João Silva"},
		"email":   {"joao@example.com"},
		"phone":   {"+55 11 99999-9999"},
		"subject": {"Prospecção comercial"},
		"message": {"Gostaria de saber mais sobre imóveis comerciais em São Paulo"},
	}

	req := httptest.NewRequest("POST", "/fale-conosco", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "sucesso") {
		t.Errorf("expected success message in body, got: %s", body)
	}

	// Verify DB record
	submissions, err := td.queries.ListContactSubmissions(context.Background(), db.ListContactSubmissionsParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list submissions: %v", err)
	}
	if len(submissions) != 1 {
		t.Errorf("expected 1 submission, got %d", len(submissions))
	}
	if submissions[0].Name != "João Silva" {
		t.Errorf("expected name 'João Silva', got '%s'", submissions[0].Name)
	}
}

func TestContactSubmitInvalidEmail(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{
		"name":    {"João"},
		"email":   {"invalid-email"},
		"subject": {"Teste"},
		"message": {"Esta é uma mensagem de teste válida"},
	}

	req := httptest.NewRequest("POST", "/fale-conosco", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "inválido") {
		t.Error("expected 'inválido' error message in body")
	}
}

func TestContactSubmitShortMessage(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{
		"name":    {"João"},
		"email":   {"joao@example.com"},
		"subject": {"Teste"},
		"message": {"curta"},
	}

	req := httptest.NewRequest("POST", "/fale-conosco", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "pelo menos 10") {
		t.Error("expected 'pelo menos 10' error message in body")
	}
}

func TestContactSubmitMissingName(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{
		"name":    {"J"},
		"email":   {"joao@example.com"},
		"subject": {"Teste"},
		"message": {"Esta é uma mensagem de teste válida"},
	}

	req := httptest.NewRequest("POST", "/fale-conosco", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "pelo menos 2") {
		t.Error("expected 'pelo menos 2' error message for name")
	}
}

func TestContactSubmitNoJS(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{
		"name":    {"João Silva"},
		"email":   {"joao@example.com"},
		"subject": {"Teste"},
		"message": {"Esta é uma mensagem de teste válida e longa"},
	}

	req := httptest.NewRequest("POST", "/fale-conosco", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect, got %d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	if !strings.Contains(loc, "success=1") {
		t.Errorf("expected redirect with success=1, got %s", loc)
	}
}

func TestNewsletterSubscribeNew(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{"email": {"new@example.com"}}

	req := httptest.NewRequest("POST", "/newsletter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "confirmada") {
		t.Errorf("expected 'confirmada' in body, got: %s", body)
	}

	// Verify DB record
	sub, err := td.queries.GetNewsletterSubscriberByEmail(context.Background(), "new@example.com")
	if err != nil {
		t.Fatalf("get subscriber: %v", err)
	}
	if !sub.Active {
		t.Error("expected subscriber to be active")
	}
}

func TestNewsletterSubscribeDuplicate(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{"email": {"duplicate@example.com"}}

	// First subscription
	req1 := httptest.NewRequest("POST", "/newsletter", strings.NewReader(form.Encode()))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req1.Header.Set("HX-Request", "true")
	rr1 := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr1, req1)

	// Second subscription (same email)
	req2 := httptest.NewRequest("POST", "/newsletter", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("HX-Request", "true")
	rr2 := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr2, req2)

	body := rr2.Body.String()
	if !strings.Contains(body, "já está inscrito") {
		t.Errorf("expected 'já está inscrito' in body, got: %s", body)
	}

	// Verify no duplicate
	subs, err := td.queries.ListActiveNewsletterSubscribers(context.Background())
	if err != nil {
		t.Fatalf("list subscribers: %v", err)
	}
	count := 0
	for _, s := range subs {
		if s.Email == "duplicate@example.com" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 subscriber, got %d", count)
	}
}

func TestNewsletterSubscribeInvalidEmail(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{"email": {"not-an-email"}}

	req := httptest.NewRequest("POST", "/newsletter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "inválido") {
		t.Error("expected 'inválido' error message in body")
	}
}

func TestNewsletterSubscribeNoJS(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{"email": {"nojs@example.com"}}

	req := httptest.NewRequest("POST", "/newsletter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect, got %d", rr.Code)
	}
}

func TestContactSubmitRateLimited(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{
		"name":    {"Test"},
		"email":   {"test@example.com"},
		"subject": {"Test"},
		"message": {"This is a test message for rate limiting validation"},
	}

	// Make 5 rapid submissions (should succeed)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/fale-conosco", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")
		req.RemoteAddr = "10.0.0.1:12345"
		rr := httptest.NewRecorder()
		newInstRouter(td).ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("attempt %d: expected 200, got %d", i+1, rr.Code)
		}
	}

	// 6th should be rate limited
	req := httptest.NewRequest("POST", "/fale-conosco", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req.RemoteAddr = "10.0.0.1:12345"
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on 6th attempt, got %d", rr.Code)
	}
}

func TestFaleConoscoSuccessParam(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/fale-conosco?success=1", nil)
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "sucesso") {
		t.Error("expected success message when success=1 query param is present")
	}
}

func TestContactSubmitInvalidForm(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	// Send malformed form data
	req := httptest.NewRequest("POST", "/fale-conosco", strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed form, got %d", rr.Code)
	}
}

func TestContactSubmitErrorNoJS(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{
		"name":    {"J"},
		"email":   {"bad"},
		"subject": {"Test"},
		"message": {"short"},
	}

	// Non-HTMX request with validation errors -- should render full page
	req := httptest.NewRequest("POST", "/fale-conosco", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "inválido") && !strings.Contains(body, "pelo menos") {
		t.Error("expected validation error in rendered page")
	}
}

func TestNewsletterSubscribeRateLimited(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{"email": {"ratetest@example.com"}}

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/newsletter", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")
		req.RemoteAddr = "10.0.0.2:12345"
		rr := httptest.NewRecorder()
		newInstRouter(td).ServeHTTP(rr, req)
	}

	req := httptest.NewRequest("POST", "/newsletter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req.RemoteAddr = "10.0.0.2:12345"
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
}

func TestNewsletterInvalidForm(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	// Send a request with no Content-Type to trigger ParseForm error
	req := httptest.NewRequest("POST", "/newsletter", strings.NewReader("this is not url-encoded"))
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	// ParseForm may or may not fail depending on content type
	// Just verify the handler doesn't crash
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusOK && rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 400/200/303, got %d", rr.Code)
	}
}

func TestNewsletterErrorNoJS(t *testing.T) {
	td := setupInstTestDB(t)
	defer td.teardown(t)

	form := url.Values{"email": {"bad-email"}}

	req := httptest.NewRequest("POST", "/newsletter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	newInstRouter(td).ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect for no-JS error, got %d", rr.Code)
	}
}
