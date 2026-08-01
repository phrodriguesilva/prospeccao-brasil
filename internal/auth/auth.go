package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"prospeccaobrasil/internal/db"
)

// LoginResult indicates what the caller should do after a login attempt.
type LoginResult struct {
	User          db.User
	Need2FASetup  bool // totp_enabled=false -> redirect to /2fa/setup
	Need2FAVerify bool // totp_enabled=true -> redirect to /2fa/verify
	Skip2FA       bool // 2FA disabled globally -> create session directly
}

// Service provides authentication operations (login, 2FA, session management).
type Service struct {
	queries    *db.Queries
	pool       *pgxpool.Pool
	key        []byte // AES-256-GCM key for totp_secret encryption
	limiter    *RateLimiter
	log        *slog.Logger
	sessionTTL time.Duration
	require2FA bool // if false, login skips 2FA and creates session directly
}

// NewService creates a new auth Service.
func NewService(queries *db.Queries, pool *pgxpool.Pool, key []byte, limiter *RateLimiter, log *slog.Logger) *Service {
	return &Service{
		queries:    queries,
		pool:       pool,
		key:        key,
		limiter:    limiter,
		log:        log,
		sessionTTL: SessionMaxAge * time.Second,
		require2FA: true,
	}
}

// SetRequire2FA configures whether 2FA is required at login.
// When false, Login returns Skip2FA=true and the handler creates a session
// directly without redirecting to /2fa/setup or /2fa/verify.
func (s *Service) SetRequire2FA(required bool) {
	s.require2FA = required
}

// Limiter returns the rate limiter (for handler use).
func (s *Service) Limiter() *RateLimiter {
	return s.limiter
}

// Login validates email + password and returns a LoginResult indicating
// the next step (2FA setup, 2FA verify, or error). Does NOT create a session
// -- the caller creates the session after 2FA is completed.
func (s *Service) Login(ctx context.Context, email, password string, tenantID pgtype.UUID) (*LoginResult, error) {
	user, err := s.queries.GetUserForAuth(ctx, db.GetUserForAuthParams{
		Email:    email,
		TenantID: tenantID,
	})
	if err != nil {
		s.log.InfoContext(ctx, "login_failed", "email", email, "reason", "user_not_found")
		return nil, ErrInvalidCredentials
	}

	// Check if account is locked
	if user.LockedAt.Valid {
		if time.Since(user.LockedAt.Time) < 15*time.Minute {
			s.log.InfoContext(ctx, "login_failed", "email", email, "reason", "account_locked")
			return nil, ErrAccountLocked
		}
		// Lock expired -- reset
		if err := s.queries.ResetFailedLoginAttempts(ctx, db.ResetFailedLoginAttemptsParams{
			ID:       user.ID,
			TenantID: user.TenantID,
		}); err != nil {
			return nil, fmt.Errorf("login: reset lock: %w", err)
		}
		user.FailedLoginAttempts = 0
		user.LockedAt = pgtype.Timestamptz{}
	}

	// Verify password
	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		attempts := user.FailedLoginAttempts + 1
		var lockedAt pgtype.Timestamptz
		if attempts >= 5 {
			lockedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
			s.log.InfoContext(ctx, "account_locked", "email", email, "attempts", attempts)
		}
		if _, err := s.queries.UpdateUserLoginAttempts(ctx, db.UpdateUserLoginAttemptsParams{
			ID:                  user.ID,
			FailedLoginAttempts: attempts,
			LockedAt:            lockedAt,
			TenantID:            user.TenantID,
		}); err != nil {
			return nil, fmt.Errorf("login: update attempts: %w", err)
		}
		s.log.InfoContext(ctx, "login_failed", "email", email, "reason", "wrong_password", "attempts", attempts)
		return nil, ErrInvalidCredentials
	}

	// Password correct -- reset failed attempts
	if user.FailedLoginAttempts > 0 {
		if err := s.queries.ResetFailedLoginAttempts(ctx, db.ResetFailedLoginAttemptsParams{
			ID:       user.ID,
			TenantID: user.TenantID,
		}); err != nil {
			return nil, fmt.Errorf("login: reset attempts: %w", err)
		}
	}

	s.log.InfoContext(ctx, "login_success", "email", email, "user_id", user.ID)

	result := &LoginResult{User: user}
	if !s.require2FA {
		result.Skip2FA = true
		return result, nil
	}
	if user.TotpEnabled {
		result.Need2FAVerify = true
	} else {
		result.Need2FASetup = true
	}
	return result, nil
}

// Enroll2FA generates a TOTP secret for the user, encrypts it, stores it in
// the DB (totp_enabled=false), and returns the QR code PNG for display.
func (s *Service) Enroll2FA(ctx context.Context, user db.User) ([]byte, error) {
	secret, qrPNG, err := GenerateTOTPSecret(user.Email)
	if err != nil {
		return nil, fmt.Errorf("enroll 2fa: %w", err)
	}
	encrypted, err := EncryptSecret(secret, s.key)
	if err != nil {
		return nil, fmt.Errorf("enroll 2fa: encrypt: %w", err)
	}
	encCopy := encrypted
	if _, err := s.queries.UpdateUserTOTP(ctx, db.UpdateUserTOTPParams{
		ID:          user.ID,
		TotpSecret:  &encCopy,
		TotpEnabled: false, // enabled only after verification
		TenantID:    user.TenantID,
	}); err != nil {
		return nil, fmt.Errorf("enroll 2fa: update: %w", err)
	}
	return qrPNG, nil
}

// Complete2FASetup validates the TOTP code against the stored (encrypted)
// secret and sets totp_enabled=true if valid.
func (s *Service) Complete2FASetup(ctx context.Context, user db.User, code string) error {
	if user.TotpSecret == nil {
		return ErrNoTOTPSecret
	}
	if !ValidateTOTP(code, *user.TotpSecret, s.key) {
		return ErrInvalidTOTP
	}
	if _, err := s.queries.UpdateUserTOTP(ctx, db.UpdateUserTOTPParams{
		ID:          user.ID,
		TotpSecret:  user.TotpSecret,
		TotpEnabled: true,
		TenantID:    user.TenantID,
	}); err != nil {
		return fmt.Errorf("complete 2fa setup: %w", err)
	}
	s.log.InfoContext(ctx, "totp_enrolled", "user_id", user.ID)
	return nil
}

// Verify2FA validates the TOTP code for a user with totp_enabled=true.
func (s *Service) Verify2FA(ctx context.Context, user db.User, code string) error {
	if user.TotpSecret == nil {
		return ErrNoTOTPSecret
	}
	if !ValidateTOTP(code, *user.TotpSecret, s.key) {
		return ErrInvalidTOTP
	}
	return nil
}

// CreateSession generates a session token, stores the hash in the DB, and
// returns the raw token (for the cookie) and the session row.
func (s *Service) CreateSession(ctx context.Context, userID, tenantID pgtype.UUID) (rawToken string, session db.Session, err error) {
	raw, hash, err := GenerateSessionToken()
	if err != nil {
		return "", db.Session{}, fmt.Errorf("create session: %w", err)
	}
	sessionID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(s.sessionTTL), Valid: true}
	session, err = s.queries.CreateSession(ctx, db.CreateSessionParams{
		ID:        sessionID,
		TenantID:  tenantID,
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", db.Session{}, fmt.Errorf("create session: insert: %w", err)
	}
	return raw, session, nil
}

// Logout revokes a session by ID.
func (s *Service) Logout(ctx context.Context, sessionID, tenantID pgtype.UUID) error {
	if err := s.queries.RevokeSessionByID(ctx, db.RevokeSessionByIDParams{
		ID:       sessionID,
		TenantID: tenantID,
	}); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	s.log.InfoContext(ctx, "session_revoked", "session_id", sessionID)
	return nil
}

// Sentinel errors for auth operations.
var (
	ErrInvalidCredentials = fmt.Errorf("email ou senha invalidos")
	ErrAccountLocked      = fmt.Errorf("conta bloqueada, tente novamente em 15 minutos")
	ErrInvalidTOTP        = fmt.Errorf("codigo TOTP invalido")
	ErrNoTOTPSecret       = fmt.Errorf("nenhum segredo TOTP configurado")
)
