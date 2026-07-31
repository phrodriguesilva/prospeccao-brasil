package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// SessionCookieName is the name of the main session cookie.
	SessionCookieName = "session"
	// PendingSessionCookieName is the name of the short-lived 2FA pending cookie.
	PendingSessionCookieName = "pending_session"
	// SessionMaxAge is the session lifetime in seconds (24 hours).
	SessionMaxAge = 86400
	// PendingSessionMaxAge is the pending session lifetime (5 minutes).
	PendingSessionMaxAge = 300
)

// GenerateSessionToken generates a 256-bit random session token and its
// SHA-256 hash. The raw token is stored in the cookie; the hash is stored
// in the database.
func GenerateSessionToken() (raw, hash string, err error) {
	b := make([]byte, 32) // 256 bits
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate session token: %w", err)
	}
	raw = base64.URLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(sum[:])
	return raw, hash, nil
}

// SessionCookie creates the main session cookie with HttpOnly + SameSite=Strict.
// The Secure flag is set based on the secure parameter (true in production,
// false in dev).
func SessionCookie(token string, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   SessionMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}
}

// ClearCookie creates a cookie that clears the named cookie (MaxAge=0).
func ClearCookie(name string, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}
}

// PendingSessionCookie creates a short-lived HMAC-signed cookie for the 2FA
// flow. The cookie value is "userID:tenantID:signature" where the signature
// is HMAC-SHA256(userID + ":" + tenantID, key).
func PendingSessionCookie(userID, tenantID string, key []byte) (*http.Cookie, error) {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write([]byte(userID + ":" + tenantID)); err != nil {
		return nil, fmt.Errorf("pending session cookie: hmac: %w", err)
	}
	sig := hex.EncodeToString(mac.Sum(nil))
	value := fmt.Sprintf("%s:%s:%s", userID, tenantID, sig)
	return &http.Cookie{
		Name:     PendingSessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   PendingSessionMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}, nil
}

// VerifyPendingSession verifies the HMAC signature of a pending session cookie
// and returns the userID and tenantID.
func VerifyPendingSession(cookieValue string, key []byte) (userID, tenantID string, err error) {
	parts := strings.SplitN(cookieValue, ":", 3)
	if len(parts) != 3 {
		return "", "", fmt.Errorf("verify pending session: invalid format")
	}
	userID, tenantID, sig := parts[0], parts[1], parts[2]
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write([]byte(userID + ":" + tenantID)); err != nil {
		return "", "", fmt.Errorf("verify pending session: hmac: %w", err)
	}
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", "", fmt.Errorf("verify pending session: invalid signature")
	}
	return userID, tenantID, nil
}

// IsSecure checks if the APP_BASE_URL indicates HTTPS (production).
func IsSecure(appBaseURL string) bool {
	return len(appBaseURL) >= 8 && appBaseURL[:8] == "https://"
}

// SessionExpiry returns the expiry time for a new session (now + SessionMaxAge).
func SessionExpiry() time.Time {
	return time.Now().Add(SessionMaxAge * time.Second)
}
