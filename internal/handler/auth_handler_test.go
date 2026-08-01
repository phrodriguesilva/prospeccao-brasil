package handler

import (
	"context"
	"crypto/rand"
	"html/template"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

type handlerTestDB struct {
	queries *db.Queries
	pool    *pgxpool.Pool
	svc     *auth.Service
	handler *AuthHandler
	key     []byte
	tmpl    *template.Template
}

func setupHandlerTestDB(t *testing.T) *handlerTestDB {
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
	key := make([]byte, 32)
	rand.Read(key)
	limiter := auth.NewRateLimiter()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := auth.NewService(queries, pool, key, limiter, log)

	tmpl := template.Must(template.New("").Funcs(TemplateFuncs()).ParseGlob(filepath.Join(findTemplateDir(t), "*.html")))
	template.Must(tmpl.ParseGlob(filepath.Join(findTemplateDir(t), "partials", "*.html")))
	template.Must(tmpl.ParseGlob(filepath.Join(findTemplateDir(t), "fragments", "*.html")))
	handler := NewAuthHandler(svc, queries, tmpl, log, false, key)

	return &handlerTestDB{queries: queries, pool: pool, svc: svc, handler: handler, key: key, tmpl: tmpl}
}

func (td *handlerTestDB) teardown(t *testing.T) {
	t.Helper()
	td.pool.Close()
}

func (td *handlerTestDB) seed(t *testing.T) (tenantID, userID pgtype.UUID) {
	t.Helper()
	tenantID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, _ = td.queries.CreateTenant(context.Background(), db.CreateTenantParams{
		ID:   tenantID,
		Name: "Handler Test Tenant",
		Plan: "free",
	})
	hash, _ := auth.HashPassword("test123")
	userID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, _ = td.queries.CreateUser(context.Background(), db.CreateUserParams{
		ID:           userID,
		TenantID:     tenantID,
		Email:        "handler@prospeccao.com.br",
		FullName:     "Handler Test",
		Role:         "admin",
		PasswordHash: hash,
	})
	return tenantID, userID
}

func findTemplateDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		tmplDir := filepath.Join(dir, "internal", "template")
		if _, err := os.Stat(tmplDir); err == nil {
			return tmplDir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("template directory not found")
		}
		dir = parent
	}
}

func newRouter(h *AuthHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/login", h.LoginGET)
	r.Post("/login", h.LoginPOST)
	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return SessionValidation(h.queries, slog.Default())(next)
		})
		r.Post("/logout", h.LogoutPOST)
		r.Get("/admin", h.AdminGET)
	})
	return r
}

func TestLoginGET(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/login", nil)
	rec := httptest.NewRecorder()
	td.handler.LoginGET(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("LoginGET: got %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "email") {
		t.Error("LoginGET should contain email field")
	}
}

func TestLoginPOSTCorrectCredentials(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	td.seed(t)

	form := url.Values{"email": {"handler@prospeccao.com.br"}, "password": {"test123"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	td.handler.LoginPOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("LoginPOST correct: got %d, want %d", rec.Code, http.StatusFound)
	}
	loc := rec.Header().Get("Location")
	if loc != "/2fa/setup" && loc != "/2fa/verify" {
		t.Errorf("Location = %q, want /2fa/setup or /2fa/verify", loc)
	}
	// Should have pending_session cookie
	hasPending := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.PendingSessionCookieName {
			hasPending = true
		}
	}
	if !hasPending {
		t.Error("should set pending_session cookie")
	}
}

func TestLoginPOSTWrongCredentials(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	td.seed(t)

	form := url.Values{"email": {"handler@prospeccao.com.br"}, "password": {"wrong"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	td.handler.LoginPOST(rec, req)

	if !strings.Contains(rec.Body.String(), "Email ou senha invalidos") {
		t.Error("should show generic error message")
	}
}

func TestAdminGETWithoutAuth(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	r := newRouter(td.handler)
	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("AdminGET without auth: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestAdminGETWithAuth(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	_, _, rawToken := td.seedWithSession(t)

	r := newRouter(td.handler)
	req := httptest.NewRequest("GET", "/admin", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("AdminGET with auth: got %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "handler@prospeccao.com.br") {
		t.Error("should show user email")
	}
}

func TestLogoutPOST(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	_, _, rawToken := td.seedWithSession(t)

	r := newRouter(td.handler)
	req := httptest.NewRequest("POST", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("LogoutPOST: got %d, want %d", rec.Code, http.StatusFound)
	}
	if rec.Header().Get("Location") != "/login" {
		t.Errorf("Location = %q, want /login", rec.Header().Get("Location"))
	}

	// Verify session is revoked
	req2 := httptest.NewRequest("GET", "/admin", nil)
	req2.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusFound {
		t.Errorf("after logout, AdminGET: got %d, want %d", rec2.Code, http.StatusFound)
	}
}

func TestTotpVerifyPOSTInvalidCode(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)

	// Enable TOTP for the user
	secret := "JBSWY3DPEHPK3PXP"
	enc, _ := auth.EncryptSecret(secret, td.key)
	encCopy := enc
	_, _ = td.queries.UpdateUserTOTP(context.Background(), db.UpdateUserTOTPParams{
		ID:          userID,
		TotpSecret:  &encCopy,
		TotpEnabled: true,
		TenantID:    tenantID,
	})

	// Create pending session cookie
	pending, _ := auth.PendingSessionCookie(userID.String(), tenantID.String(), td.key)

	form := url.Values{"code": {"000000"}}
	req := httptest.NewRequest("POST", "/2fa/verify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpVerifyPOST(rec, req)

	if !strings.Contains(rec.Body.String(), "Codigo TOTP invalido") {
		t.Error("should show TOTP error for invalid code")
	}
}

func TestTotpVerifyPOSTValidCode(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)

	// Enable TOTP with a real secret
	key2, _ := totp.Generate(totp.GenerateOpts{Issuer: "Test", AccountName: "handler@prospeccao.com.br"})
	enc, _ := auth.EncryptSecret(key2.Secret(), td.key)
	encCopy := enc
	_, _ = td.queries.UpdateUserTOTP(context.Background(), db.UpdateUserTOTPParams{
		ID:          userID,
		TotpSecret:  &encCopy,
		TotpEnabled: true,
		TenantID:    tenantID,
	})

	// Generate valid code
	code, _ := totp.GenerateCodeCustom(key2.Secret(), time.Now(), totp.ValidateOpts{
		Period: 30, Skew: 1, Digits: 6, Algorithm: 0,
	})

	pending, _ := auth.PendingSessionCookie(userID.String(), tenantID.String(), td.key)

	form := url.Values{"code": {code}}
	req := httptest.NewRequest("POST", "/2fa/verify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpVerifyPOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("valid TOTP: got %d, want %d", rec.Code, http.StatusFound)
	}
	if rec.Header().Get("Location") != "/admin" {
		t.Errorf("Location = %q, want /admin", rec.Header().Get("Location"))
	}
}

func TestTotpSetupGET(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	pending, _ := auth.PendingSessionCookie(userID.String(), tenantID.String(), td.key)

	req := httptest.NewRequest("GET", "/2fa/setup", nil)
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpSetupGET(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("TotpSetupGET: got %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "data:image/png;base64") {
		t.Error("should contain QR code image")
	}
}

func TestTotpSetupGETNoPendingCookie(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/2fa/setup", nil)
	rec := httptest.NewRecorder()
	td.handler.TotpSetupGET(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("TotpSetupGET no cookie: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestTotpSetupPOSTValidCode(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)

	// Enroll first to get a secret
	user, _ := td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})
	_, _ = td.svc.Enroll2FA(context.Background(), user)
	user, _ = td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})

	// Generate valid code
	secret, _ := auth.DecryptSecret(*user.TotpSecret, td.key)
	code, _ := totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period: 30, Skew: 1, Digits: 6, Algorithm: 0,
	})

	pending, _ := auth.PendingSessionCookie(userID.String(), tenantID.String(), td.key)
	form := url.Values{"code": {code}}
	req := httptest.NewRequest("POST", "/2fa/setup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpSetupPOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("TotpSetupPOST valid: got %d, want %d", rec.Code, http.StatusFound)
	}
	if rec.Header().Get("Location") != "/admin" {
		t.Errorf("Location = %q, want /admin", rec.Header().Get("Location"))
	}
}

func TestTotpVerifyGET(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	pending, _ := auth.PendingSessionCookie(userID.String(), tenantID.String(), td.key)

	req := httptest.NewRequest("GET", "/2fa/verify", nil)
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpVerifyGET(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("TotpVerifyGET: got %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "code") {
		t.Error("should contain code input field")
	}
}

func TestTotpVerifyGETNoPendingCookie(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/2fa/verify", nil)
	rec := httptest.NewRecorder()
	td.handler.TotpVerifyGET(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("TotpVerifyGET no cookie: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestLoginPOSTAccountLocked(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)

	// Lock the account
	lockedAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	_, _ = td.queries.UpdateUserLoginAttempts(context.Background(), db.UpdateUserLoginAttemptsParams{
		ID:                  userID,
		FailedLoginAttempts: 5,
		LockedAt:            lockedAt,
		TenantID:            tenantID,
	})

	form := url.Values{"email": {"handler@prospeccao.com.br"}, "password": {"test123"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	td.handler.LoginPOST(rec, req)

	if !strings.Contains(rec.Body.String(), "bloqueada") {
		t.Error("should show account locked message")
	}
}

func TestTotpSetupGETInvalidUser(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	// Create a pending cookie with a non-existent user UUID
	fakeUserID := uuid.New().String()
	tenantID, _ := td.seed(t)
	pending, _ := auth.PendingSessionCookie(fakeUserID, tenantID.String(), td.key)

	req := httptest.NewRequest("GET", "/2fa/setup", nil)
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpSetupGET(rec, req)

	// Should redirect to login (user not found)
	if rec.Code != http.StatusFound {
		t.Errorf("TotpSetupGET invalid user: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestTotpSetupPOSTInvalidUser(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	fakeUserID := uuid.New().String()
	tenantID, _ := td.seed(t)
	pending, _ := auth.PendingSessionCookie(fakeUserID, tenantID.String(), td.key)

	form := url.Values{"code": {"123456"}}
	req := httptest.NewRequest("POST", "/2fa/setup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpSetupPOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("TotpSetupPOST invalid user: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestTotpVerifyPOSTInvalidUser(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	fakeUserID := uuid.New().String()
	tenantID, _ := td.seed(t)
	pending, _ := auth.PendingSessionCookie(fakeUserID, tenantID.String(), td.key)

	form := url.Values{"code": {"123456"}}
	req := httptest.NewRequest("POST", "/2fa/verify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpVerifyPOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("TotpVerifyPOST invalid user: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestTotpSetupPOSTNoPendingCookie(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	form := url.Values{"code": {"123456"}}
	req := httptest.NewRequest("POST", "/2fa/setup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	td.handler.TotpSetupPOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("TotpSetupPOST no cookie: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestTotpVerifyPOSTNoPendingCookie(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	form := url.Values{"code": {"123456"}}
	req := httptest.NewRequest("POST", "/2fa/verify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	td.handler.TotpVerifyPOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("TotpVerifyPOST no cookie: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestLoginPOSTRateLimitedDirectly(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)
	td.seed(t)

	// Pre-exhaust the rate limiter
	limiter := td.svc.Limiter()
	for i := 0; i < 5; i++ {
		limiter.AllowBoth("10.0.0.1", "handler@prospeccao.com.br")
	}

	form := url.Values{"email": {"handler@prospeccao.com.br"}, "password": {"test123"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "10.0.0.1:12345"
	rec := httptest.NewRecorder()
	td.handler.LoginPOST(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("rate limited: got %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestLoginPOSTNoTenants(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	// Truncate tenants table
	_, err := td.pool.Exec(context.Background(), "TRUNCATE tenants CASCADE")
	if err != nil {
		t.Fatalf("truncate tenants: %v", err)
	}

	form := url.Values{"email": {"test@prospeccao.com.br"}, "password": {"test123"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	td.handler.LoginPOST(rec, req)

	// Should show internal error (200 with error message, not 302)
	if rec.Code != http.StatusOK {
		t.Errorf("LoginPOST no tenants: got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestSessionValidationNoTenants(t *testing.T) {
	td := setupMWTestDB(t)
	defer td.teardown(t)

	_, _, rawToken := td.seed(t)

	// Truncate tenants table (cascades to users + sessions)
	_, err := td.pool.Exec(context.Background(), "TRUNCATE tenants CASCADE")
	if err != nil {
		t.Fatalf("truncate tenants: %v", err)
	}

	mw := SessionValidation(td.queries, slog.Default())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with no tenants")
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("no tenants: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestRenderTemplateError(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	// Create a handler with a template that will fail on execution
	badTmpl := template.Must(template.New("login.html").Parse(`{{call .Error}}`))
	badHandler := NewAuthHandler(td.svc, td.queries, badTmpl, slog.Default(), false, td.key)

	req := httptest.NewRequest("GET", "/login", nil)
	rec := httptest.NewRecorder()
	badHandler.LoginGET(rec, req)

	// Should get internal server error from renderTemplate failure
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("renderTemplate error: got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestTotpSetupGETWithExistingSecret(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	// Set a TOTP secret but don't enable
	enc, _ := auth.EncryptSecret("JBSWY3DPEHPK3PXP", td.key)
	encCopy := enc
	_, _ = td.queries.UpdateUserTOTP(context.Background(), db.UpdateUserTOTPParams{
		ID:          userID,
		TotpSecret:  &encCopy,
		TotpEnabled: false,
		TenantID:    tenantID,
	})

	pending, _ := auth.PendingSessionCookie(userID.String(), tenantID.String(), td.key)
	req := httptest.NewRequest("GET", "/2fa/setup", nil)
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpSetupGET(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("TotpSetupGET with secret: got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTotpVerifyPOSTNoSecret(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	pending, _ := auth.PendingSessionCookie(userID.String(), tenantID.String(), td.key)

	form := url.Values{"code": {"123456"}}
	req := httptest.NewRequest("POST", "/2fa/verify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpVerifyPOST(rec, req)

	// User has no TOTP secret -- should show error
	if !strings.Contains(rec.Body.String(), "invalido") {
		t.Error("should show TOTP error for no secret")
	}
}

func TestLogoutPOSTAlreadyRevoked(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	tenantID, _, rawToken := td.seedWithSession(t)

	// Revoke the session first
	session, _ := td.queries.GetSessionByTokenHash(context.Background(), db.GetSessionByTokenHashParams{
		TokenHash: sha256Hex(rawToken),
		TenantID:  tenantID,
	})
	_ = td.queries.RevokeSessionByID(context.Background(), db.RevokeSessionByIDParams{
		ID:       session.ID,
		TenantID: tenantID,
	})

	// Now try to logout again -- should still redirect (Logout error is logged)
	r := newRouter(td.handler)
	req := httptest.NewRequest("POST", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// The session is already revoked, so SessionValidation will redirect to /login
	if rec.Code != http.StatusFound {
		t.Errorf("logout already revoked: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestClientIPFromRequestNoPort(t *testing.T) {
	req := &http.Request{RemoteAddr: "noport"}
	got := clientIPFromRequest(req)
	if got != "noport" {
		t.Errorf("clientIPFromRequest no port: got %q, want noport", got)
	}
}

func TestTotpSetupGETAlreadyEnrolled(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	// Enable TOTP
	enc, _ := auth.EncryptSecret("JBSWY3DPEHPK3PXP", td.key)
	encCopy := enc
	_, _ = td.queries.UpdateUserTOTP(context.Background(), db.UpdateUserTOTPParams{
		ID:          userID,
		TotpSecret:  &encCopy,
		TotpEnabled: true,
		TenantID:    tenantID,
	})

	pending, _ := auth.PendingSessionCookie(userID.String(), tenantID.String(), td.key)
	req := httptest.NewRequest("GET", "/2fa/setup", nil)
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpSetupGET(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("already enrolled: got %d, want %d", rec.Code, http.StatusFound)
	}
	if rec.Header().Get("Location") != "/2fa/verify" {
		t.Errorf("Location = %q, want /2fa/verify", rec.Header().Get("Location"))
	}
}

func TestTotpSetupPOSTInvalidCode(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	pending, _ := auth.PendingSessionCookie(userID.String(), tenantID.String(), td.key)

	form := url.Values{"code": {"000000"}}
	req := httptest.NewRequest("POST", "/2fa/setup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(pending)
	rec := httptest.NewRecorder()
	td.handler.TotpSetupPOST(rec, req)

	if !strings.Contains(rec.Body.String(), "invalido") {
		t.Error("should show TOTP error for invalid code")
	}
}

func TestAdminGETNoUserInContext(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("GET", "/admin", nil)
	// No user in context -- should redirect
	ctx := context.WithValue(req.Context(), auth.CtxRole, "admin")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	td.handler.AdminGET(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("AdminGET no user: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestLogoutPOSTNoSession(t *testing.T) {
	td := setupHandlerTestDB(t)
	defer td.teardown(t)

	req := httptest.NewRequest("POST", "/logout", nil)
	rec := httptest.NewRecorder()
	td.handler.LogoutPOST(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("LogoutPOST no session: got %d, want %d", rec.Code, http.StatusFound)
	}
}

// seedWithSession creates a tenant, user, and active session. Returns the raw
// session token for use in cookie.
func (td *handlerTestDB) seedWithSession(t *testing.T) (tenantID, userID pgtype.UUID, rawToken string) {
	t.Helper()
	tenantID, userID = td.seed(t)
	rawToken, _, err := td.svc.CreateSession(context.Background(), userID, tenantID)
	if err != nil {
		t.Fatalf("seedWithSession: %v", err)
	}
	return tenantID, userID, rawToken
}
