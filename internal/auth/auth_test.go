package auth

import (
	"context"
	"crypto/rand"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"prospeccaobrasil/internal/db"
)

type testDB struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	svc     *Service
	key     []byte
}

func setupTestDB(t *testing.T) *testDB {
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
	limiter := NewRateLimiter()
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := NewService(queries, pool, key, limiter, log)

	td := &testDB{pool: pool, queries: queries, svc: svc, key: key}
	return td
}

func (td *testDB) teardown(t *testing.T) {
	t.Helper()
	td.pool.Close()
}

func (td *testDB) seed(t *testing.T) (tenantID pgtype.UUID, userID pgtype.UUID) {
	t.Helper()
	tenantID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, err := td.queries.CreateTenant(context.Background(), db.CreateTenantParams{
		ID:   tenantID,
		Name: "Test Tenant",
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("seed tenant: %v", err)
	}
	hash, _ := HashPassword("test123")
	userID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	_, err = td.queries.CreateUser(context.Background(), db.CreateUserParams{
		ID:           userID,
		TenantID:     tenantID,
		Email:        "admin@prospeccao.com.br",
		FullName:     "Admin",
		Role:         "admin",
		PasswordHash: hash,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return tenantID, userID
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

func TestLoginCorrectCredentials(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, _ := td.seed(t)
	result, err := td.svc.Login(context.Background(), "admin@prospeccao.com.br", "test123", tenantID)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if result.Need2FASetup != true {
		t.Error("Need2FASetup should be true for new user")
	}
	if result.User.Email != "admin@prospeccao.com.br" {
		t.Errorf("Email = %q", result.User.Email)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, _ := td.seed(t)
	_, err := td.svc.Login(context.Background(), "admin@prospeccao.com.br", "wrong", tenantID)
	if err != ErrInvalidCredentials {
		t.Errorf("Login wrong password: got %v, want ErrInvalidCredentials", err)
	}

	// Verify failed_login_attempts was incremented
	user, _ := td.queries.GetUserForAuth(context.Background(), db.GetUserForAuthParams{
		Email:    "admin@prospeccao.com.br",
		TenantID: tenantID,
	})
	if user.FailedLoginAttempts != 1 {
		t.Errorf("FailedLoginAttempts = %d, want 1", user.FailedLoginAttempts)
	}
}

func TestLoginAccountLockout(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, _ := td.seed(t)

	// 5 failed attempts
	for i := 0; i < 5; i++ {
		_, _ = td.svc.Login(context.Background(), "admin@prospeccao.com.br", "wrong", tenantID)
	}

	// 6th attempt should return locked
	_, err := td.svc.Login(context.Background(), "admin@prospeccao.com.br", "test123", tenantID)
	if err != ErrAccountLocked {
		t.Errorf("Login after lockout: got %v, want ErrAccountLocked", err)
	}
}

func TestLoginResetsAttemptsOnSuccess(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, _ := td.seed(t)

	// 2 failed attempts
	_, _ = td.svc.Login(context.Background(), "admin@prospeccao.com.br", "wrong", tenantID)
	_, _ = td.svc.Login(context.Background(), "admin@prospeccao.com.br", "wrong", tenantID)

	// Successful login
	_, err := td.svc.Login(context.Background(), "admin@prospeccao.com.br", "test123", tenantID)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	user, _ := td.queries.GetUserForAuth(context.Background(), db.GetUserForAuthParams{
		Email:    "admin@prospeccao.com.br",
		TenantID: tenantID,
	})
	if user.FailedLoginAttempts != 0 {
		t.Errorf("FailedLoginAttempts = %d, want 0", user.FailedLoginAttempts)
	}
}

func TestComplete2FASetupValidCode(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	user, _ := td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})

	// Enroll
	_, err := td.svc.Enroll2FA(context.Background(), user)
	if err != nil {
		t.Fatalf("Enroll2FA: %v", err)
	}

	// Get updated user with secret
	user, _ = td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})

	// Decrypt secret and generate valid code
	secret, _ := DecryptSecret(*user.TotpSecret, td.key)
	code := generateTOTPCode(t, secret)

	err = td.svc.Complete2FASetup(context.Background(), user, code)
	if err != nil {
		t.Fatalf("Complete2FASetup: %v", err)
	}

	// Verify totp_enabled
	user, _ = td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})
	if !user.TotpEnabled {
		t.Error("totp_enabled should be true")
	}
}

func TestComplete2FASetupInvalidCode(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	user, _ := td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})

	_, _ = td.svc.Enroll2FA(context.Background(), user)
	user, _ = td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})

	err := td.svc.Complete2FASetup(context.Background(), user, "000000")
	if err != ErrInvalidTOTP {
		t.Errorf("Complete2FASetup invalid: got %v, want ErrInvalidTOTP", err)
	}
}

func TestCreateSession(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	raw, session, err := td.svc.CreateSession(context.Background(), userID, tenantID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if raw == "" {
		t.Error("raw token should not be empty")
	}
	if !session.ID.Valid {
		t.Error("session ID should be valid")
	}
}

func TestLogout(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	_, session, _ := td.svc.CreateSession(context.Background(), userID, tenantID)

	err := td.svc.Logout(context.Background(), session.ID, tenantID)
	if err != nil {
		t.Fatalf("Logout: %v", err)
	}

	// Verify session is revoked (GetSessionWithUser should fail)
	_, err = td.queries.GetSessionWithUser(context.Background(), db.GetSessionWithUserParams{
		TokenHash: session.TokenHash,
		TenantID:  tenantID,
	})
	if err == nil {
		t.Error("session should be revoked (not found)")
	}
}

func TestVerify2FAValidCode(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	user, _ := td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})

	// Enroll and complete setup
	_, _ = td.svc.Enroll2FA(context.Background(), user)
	user, _ = td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})
	secret, _ := DecryptSecret(*user.TotpSecret, td.key)
	code := generateTOTPCode(t, secret)
	_ = td.svc.Complete2FASetup(context.Background(), user, code)

	// Now test Verify2FA
	user, _ = td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})
	validCode := generateTOTPCode(t, secret)
	err := td.svc.Verify2FA(context.Background(), user, validCode)
	if err != nil {
		t.Fatalf("Verify2FA valid: %v", err)
	}
}

func TestVerify2FAInvalidCode(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	user, _ := td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})
	_, _ = td.svc.Enroll2FA(context.Background(), user)
	user, _ = td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})

	err := td.svc.Verify2FA(context.Background(), user, "000000")
	if err != ErrInvalidTOTP {
		t.Errorf("Verify2FA invalid: got %v, want ErrInvalidTOTP", err)
	}
}

func TestVerify2FANoSecret(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)
	user, _ := td.queries.GetUserByID(context.Background(), db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})

	err := td.svc.Verify2FA(context.Background(), user, "123456")
	if err != ErrNoTOTPSecret {
		t.Errorf("Verify2FA no secret: got %v, want ErrNoTOTPSecret", err)
	}
}

func TestLoginNonExistentUser(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, _ := td.seed(t)
	_, err := td.svc.Login(context.Background(), "nobody@prospeccao.com.br", "test123", tenantID)
	if err != ErrInvalidCredentials {
		t.Errorf("Login non-existent: got %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginExpiredLockReset(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	tenantID, userID := td.seed(t)

	// Lock the account 20 minutes ago (lock should be expired)
	lockedAt := pgtype.Timestamptz{Time: time.Now().Add(-20 * time.Minute), Valid: true}
	_, _ = td.queries.UpdateUserLoginAttempts(context.Background(), db.UpdateUserLoginAttemptsParams{
		ID:                  userID,
		FailedLoginAttempts: 5,
		LockedAt:            lockedAt,
		TenantID:            tenantID,
	})

	// Login with correct password -- should succeed (lock expired, reset)
	result, err := td.svc.Login(context.Background(), "admin@prospeccao.com.br", "test123", tenantID)
	if err != nil {
		t.Fatalf("Login with expired lock: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}

	// Verify failed_login_attempts was reset
	user, _ := td.queries.GetUserForAuth(context.Background(), db.GetUserForAuthParams{
		Email:    "admin@prospeccao.com.br",
		TenantID: tenantID,
	})
	if user.FailedLoginAttempts != 0 {
		t.Errorf("FailedLoginAttempts = %d, want 0 (reset after expired lock)", user.FailedLoginAttempts)
	}
	if user.LockedAt.Valid {
		t.Error("LockedAt should be reset to NULL")
	}
}

func TestLimiterAccessor(t *testing.T) {
	td := setupTestDB(t)
	defer td.teardown(t)

	if td.svc.Limiter() == nil {
		t.Error("Limiter() should not return nil")
	}
}

// generateTOTPCode generates a valid TOTP code for the current time.
func generateTOTPCode(t *testing.T, secret string) string {
	t.Helper()
	// Use pquerna/otp to generate a code
	return generateTOTPCodeAt(secret, time.Now())
}

func generateTOTPCodeAt(secret string, now time.Time) string {
	// Import would be circular, so use a helper
	return totpCode(secret, now)
}
