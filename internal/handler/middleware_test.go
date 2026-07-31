package handler

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

type mwTestDB struct {
	queries *db.Queries
	pool    *pgxpool.Pool
	svc     *auth.Service
	key     []byte
}

func setupMWTestDB(t *testing.T) *mwTestDB {
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

	return &mwTestDB{queries: queries, pool: pool, svc: svc, key: key}
}

func (td *mwTestDB) teardown(t *testing.T) {
	t.Helper()
	td.pool.Close()
}

func findMigrationsDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		migrationsDir := filepath.Join(dir, "migrations")
		if _, err := os.Stat(migrationsDir); err == nil {
			abs, _ := filepath.Abs(migrationsDir)
			return abs
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("migrations directory not found")
		}
		dir = parent
	}
}

func (td *mwTestDB) seed(t *testing.T) (tenantID, userID pgtype.UUID, rawToken string) {
	t.Helper()
	tenantID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, _ = td.queries.CreateTenant(context.Background(), db.CreateTenantParams{
		ID:   tenantID,
		Name: "MW Test Tenant",
		Plan: "free",
	})
	hash, _ := auth.HashPassword("test123")
	userID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, _ = td.queries.CreateUser(context.Background(), db.CreateUserParams{
		ID:           userID,
		TenantID:     tenantID,
		Email:        "mwtest@prospeccao.com.br",
		FullName:     "MW Test",
		Role:         "admin",
		PasswordHash: hash,
	})
	rawToken, _, err := td.svc.CreateSession(context.Background(), userID, tenantID)
	if err != nil {
		t.Fatalf("seed: create session: %v", err)
	}
	return tenantID, userID, rawToken
}

func TestSessionValidationNoCookie(t *testing.T) {
	td := setupMWTestDB(t)
	defer td.teardown(t)

	mw := SessionValidation(td.queries, slog.Default())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("no cookie: got %d, want %d", rec.Code, http.StatusFound)
	}
	if rec.Header().Get("Location") != "/login" {
		t.Errorf("Location = %q, want /login", rec.Header().Get("Location"))
	}
}

func TestSessionValidationValidCookie(t *testing.T) {
	td := setupMWTestDB(t)
	defer td.teardown(t)

	_, _, rawToken := td.seed(t)

	called := false
	mw := SessionValidation(td.queries, slog.Default())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		role, ok := auth.RoleFromContext(r.Context())
		if !ok || role != "admin" {
			t.Errorf("role = %q, %v, want admin", role, ok)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should be called with valid cookie")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("valid cookie: got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestSessionValidationExpiredSession(t *testing.T) {
	td := setupMWTestDB(t)
	defer td.teardown(t)

	tenantID, _, rawToken := td.seed(t)

	// Expire the session by updating expires_at in the past
	session, _ := td.queries.GetSessionByTokenHash(context.Background(), db.GetSessionByTokenHashParams{
		TokenHash: sha256Hex(rawToken),
		TenantID:  tenantID,
	})
	_, err := td.pool.Exec(context.Background(),
		"UPDATE sessions SET expires_at = now() - interval '1 hour' WHERE id = $1", session.ID)
	if err != nil {
		t.Fatalf("expire session: %v", err)
	}

	mw := SessionValidation(td.queries, slog.Default())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for expired session")
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expired session: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestSessionValidationRevokedSession(t *testing.T) {
	td := setupMWTestDB(t)
	defer td.teardown(t)

	tenantID, _, rawToken := td.seed(t)

	// Revoke the session
	session, _ := td.queries.GetSessionByTokenHash(context.Background(), db.GetSessionByTokenHashParams{
		TokenHash: sha256Hex(rawToken),
		TenantID:  tenantID,
	})
	_ = td.queries.RevokeSessionByID(context.Background(), db.RevokeSessionByIDParams{
		ID:       session.ID,
		TenantID: tenantID,
	})

	mw := SessionValidation(td.queries, slog.Default())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for revoked session")
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("revoked session: got %d, want %d", rec.Code, http.StatusFound)
	}
}

func TestSessionValidationSoftDeletedUser(t *testing.T) {
	td := setupMWTestDB(t)
	defer td.teardown(t)

	_, userID, rawToken := td.seed(t)

	// Soft-delete the user via direct SQL (no sqlc query for this)
	_, err := td.pool.Exec(context.Background(),
		"UPDATE users SET deleted_at = now() WHERE id = $1", userID)
	if err != nil {
		t.Fatalf("soft-delete user: %v", err)
	}

	mw := SessionValidation(td.queries, slog.Default())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for soft-deleted user")
	}))

	req := httptest.NewRequest("GET", "/admin", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: rawToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("soft-deleted user: got %d, want %d", rec.Code, http.StatusFound)
	}
}
