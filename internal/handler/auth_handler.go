package handler

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"prospeccaobrasil/internal/auth"
	"prospeccaobrasil/internal/db"
)

// AuthHandler handles auth-related HTTP endpoints: /login, /logout,
// /2fa/setup, /2fa/verify, /admin.
type AuthHandler struct {
	svc     *auth.Service
	queries *db.Queries
	tmpl    *template.Template
	log     *slog.Logger
	secure  bool
	hmacKey []byte
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc *auth.Service, queries *db.Queries, tmpl *template.Template, log *slog.Logger, secure bool, hmacKey []byte) *AuthHandler {
	return &AuthHandler{
		svc:     svc,
		queries: queries,
		tmpl:    tmpl,
		log:     log,
		secure:  secure,
		hmacKey: hmacKey,
	}
}

// LoginGET displays the login form.
func (h *AuthHandler) LoginGET(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "login.html", map[string]any{
		"Error": "",
	})
}

// LoginPOST processes login form submission.
func (h *AuthHandler) LoginPOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Rate limit
	ip := clientIPFromRequest(r)
	if !h.svc.Limiter().AllowBoth(ip, email) {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	// Get tenant ID (MVP: first active tenant)
	tenantID, err := getFirstTenantID(r.Context(), h.queries)
	if err != nil {
		h.log.ErrorContext(r.Context(), "login: get tenant", "error", err)
		h.renderTemplate(w, "login.html", map[string]any{
			"Error": "Erro interno. Tente novamente.",
		})
		return
	}

	result, err := h.svc.Login(r.Context(), email, password, tenantID)
	if err != nil {
		if err == auth.ErrAccountLocked {
			h.renderTemplate(w, "login.html", map[string]any{
				"Error": "Conta bloqueada, tente novamente em 15 minutos",
			})
			return
		}
		h.renderTemplate(w, "login.html", map[string]any{
			"Error": "Email ou senha invalidos",
		})
		return
	}

	// Set pending session cookie
	pendingCookie, err := auth.PendingSessionCookie(
		result.User.ID.String(),
		result.User.TenantID.String(),
		h.hmacKey,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, pendingCookie)

	if result.Skip2FA {
		// 2FA disabled globally -- create session directly
		rawToken, _, err := h.svc.CreateSession(r.Context(), result.User.ID, result.User.TenantID)
		if err != nil {
			h.log.ErrorContext(r.Context(), "login: create session (skip 2fa)", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		sessionCookie := auth.SessionCookie(rawToken, h.secure)
		http.SetCookie(w, sessionCookie)
		http.SetCookie(w, auth.ClearCookie(auth.PendingSessionCookieName, h.secure))
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	if result.Need2FASetup {
		http.Redirect(w, r, "/2fa/setup", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/2fa/verify", http.StatusFound)
}

// TotpSetupGET displays the QR code for TOTP enrollment.
func (h *AuthHandler) TotpSetupGET(w http.ResponseWriter, r *http.Request) {
	userID, tenantID, err := h.verifyPending(r)
	if err != nil {
		redirectToLogin(w, r)
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), db.GetUserByIDParams{
		ID:       parseUUID(userID),
		TenantID: parseUUID(tenantID),
	})
	if err != nil {
		redirectToLogin(w, r)
		return
	}

	// If already enrolled, redirect to verify
	if user.TotpEnabled {
		http.Redirect(w, r, "/2fa/verify", http.StatusFound)
		return
	}

	// Generate or re-use the secret
	var qrPNG []byte
	if user.TotpSecret == nil {
		qrPNG, err = h.svc.Enroll2FA(r.Context(), user)
		if err != nil {
			h.log.ErrorContext(r.Context(), "2fa setup: enroll", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		// Re-generate QR from existing secret (decrypt, re-generate image)
		// For simplicity, re-enroll (generates new secret). In production,
		// we'd cache the QR. MVP: re-enroll is fine.
		qrPNG, err = h.svc.Enroll2FA(r.Context(), user)
		if err != nil {
			h.log.ErrorContext(r.Context(), "2fa setup: re-enroll", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	qrB64 := base64.StdEncoding.EncodeToString(qrPNG)
	h.renderTemplate(w, "totp_setup.html", map[string]any{
		"QRCode": qrB64,
		"Error":  "",
	})
}

// TotpSetupPOST verifies the TOTP code and completes enrollment.
func (h *AuthHandler) TotpSetupPOST(w http.ResponseWriter, r *http.Request) {
	userID, tenantID, err := h.verifyPending(r)
	if err != nil {
		redirectToLogin(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	code := r.FormValue("code")

	user, err := h.queries.GetUserByID(r.Context(), db.GetUserByIDParams{
		ID:       parseUUID(userID),
		TenantID: parseUUID(tenantID),
	})
	if err != nil {
		redirectToLogin(w, r)
		return
	}

	if err := h.svc.Complete2FASetup(r.Context(), user, code); err != nil {
		h.renderTemplate(w, "totp_setup.html", map[string]any{
			"QRCode": "",
			"Error":  "Codigo TOTP invalido",
		})
		return
	}

	// Create session
	rawToken, _, err := h.svc.CreateSession(r.Context(), user.ID, user.TenantID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "2fa setup: create session", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set session cookie, clear pending cookie
	http.SetCookie(w, auth.SessionCookie(rawToken, h.secure))
	http.SetCookie(w, auth.ClearCookie(auth.PendingSessionCookieName, h.secure))
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// TotpVerifyGET displays the TOTP verification form.
func (h *AuthHandler) TotpVerifyGET(w http.ResponseWriter, r *http.Request) {
	_, _, err := h.verifyPending(r)
	if err != nil {
		redirectToLogin(w, r)
		return
	}
	h.renderTemplate(w, "totp_verify.html", map[string]any{
		"Error": "",
	})
}

// TotpVerifyPOST verifies the TOTP code for existing 2FA users.
func (h *AuthHandler) TotpVerifyPOST(w http.ResponseWriter, r *http.Request) {
	userID, tenantID, err := h.verifyPending(r)
	if err != nil {
		redirectToLogin(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	code := r.FormValue("code")

	user, err := h.queries.GetUserByID(r.Context(), db.GetUserByIDParams{
		ID:       parseUUID(userID),
		TenantID: parseUUID(tenantID),
	})
	if err != nil {
		redirectToLogin(w, r)
		return
	}

	if err := h.svc.Verify2FA(r.Context(), user, code); err != nil {
		h.renderTemplate(w, "totp_verify.html", map[string]any{
			"Error": "Codigo TOTP invalido",
		})
		return
	}

	// Create session
	rawToken, _, err := h.svc.CreateSession(r.Context(), user.ID, user.TenantID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "2fa verify: create session", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, auth.SessionCookie(rawToken, h.secure))
	http.SetCookie(w, auth.ClearCookie(auth.PendingSessionCookieName, h.secure))
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// LogoutPOST revokes the session and clears the cookie.
func (h *AuthHandler) LogoutPOST(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := SessionIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	tenantIDStr, _ := auth.TenantIDFromContext(r.Context())
	tenantID := parseUUID(tenantIDStr)

	if err := h.svc.Logout(r.Context(), sessionID, tenantID); err != nil {
		h.log.ErrorContext(r.Context(), "logout", "error", err)
	}
	http.SetCookie(w, auth.ClearCookie(auth.SessionCookieName, h.secure))
	http.Redirect(w, r, "/login", http.StatusFound)
}

// AdminGET is a placeholder for the internal system dashboard.
func (h *AuthHandler) AdminGET(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(auth.CtxUser).(*db.User)
	if !ok {
		redirectToLogin(w, r)
		return
	}
	fmt.Fprintf(w, "Authenticated as %s (role: %s)", user.Email, user.Role) //nolint:errcheck // diagnostic page
}

// renderTemplate renders a named template with data.
func (h *AuthHandler) renderTemplate(w http.ResponseWriter, name string, data any) {
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// verifyPending extracts and verifies the pending session cookie.
func (h *AuthHandler) verifyPending(r *http.Request) (userID, tenantID string, err error) {
	cookie, err := r.Cookie(auth.PendingSessionCookieName)
	if err != nil {
		return "", "", fmt.Errorf("no pending cookie: %w", err)
	}
	return auth.VerifyPendingSession(cookie.Value, h.hmacKey)
}

// parseUUID converts a string to pgtype.UUID.
func parseUUID(s string) pgtype.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

// clientIPFromRequest extracts the client IP (without port).
func clientIPFromRequest(r *http.Request) string {
	host := r.RemoteAddr
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}
