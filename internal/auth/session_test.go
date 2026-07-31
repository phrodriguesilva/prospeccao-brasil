package auth

import (
	"crypto/rand"
	"net/http"
	"testing"
	"time"
)

func testHMACKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate hmac key: %v", err)
	}
	return key
}

func TestGenerateSessionToken(t *testing.T) {
	raw, hash, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken: %v", err)
	}
	if raw == "" {
		t.Error("raw token should not be empty")
	}
	if hash == "" {
		t.Error("hash should not be empty")
	}
	if raw == hash {
		t.Error("raw token and hash should differ")
	}
}

func TestGenerateSessionTokenUnique(t *testing.T) {
	raw1, _, _ := GenerateSessionToken()
	raw2, _, _ := GenerateSessionToken()
	if raw1 == raw2 {
		t.Error("tokens should be unique")
	}
}

func TestSessionCookieAttributes(t *testing.T) {
	c := SessionCookie("test-token", true)
	if c.Name != SessionCookieName {
		t.Errorf("Name = %q, want %q", c.Name, SessionCookieName)
	}
	if !c.HttpOnly {
		t.Error("HttpOnly should be true")
	}
	if c.SameSite != http.SameSiteStrictMode {
		t.Error("SameSite should be Strict")
	}
	if c.MaxAge != SessionMaxAge {
		t.Errorf("MaxAge = %d, want %d", c.MaxAge, SessionMaxAge)
	}
}

func TestSessionCookieSecureFlag(t *testing.T) {
	secure := SessionCookie("token", true)
	if !secure.Secure {
		t.Error("Secure should be true when secure=true")
	}
	insecure := SessionCookie("token", false)
	if insecure.Secure {
		t.Error("Secure should be false when secure=false")
	}
}

func TestClearCookie(t *testing.T) {
	c := ClearCookie(SessionCookieName, true)
	if c.MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", c.MaxAge)
	}
	if c.Value != "" {
		t.Error("Value should be empty")
	}
}

func TestPendingSessionCookieRoundTrip(t *testing.T) {
	key := testHMACKey(t)
	c, err := PendingSessionCookie("user-123", "tenant-456", key)
	if err != nil {
		t.Fatalf("PendingSessionCookie: %v", err)
	}
	if c.Name != PendingSessionCookieName {
		t.Errorf("Name = %q, want %q", c.Name, PendingSessionCookieName)
	}
	uid, tid, err := VerifyPendingSession(c.Value, key)
	if err != nil {
		t.Fatalf("VerifyPendingSession: %v", err)
	}
	if uid != "user-123" {
		t.Errorf("userID = %q, want user-123", uid)
	}
	if tid != "tenant-456" {
		t.Errorf("tenantID = %q, want tenant-456", tid)
	}
}

func TestVerifyPendingSessionTampered(t *testing.T) {
	key := testHMACKey(t)
	c, _ := PendingSessionCookie("user-123", "tenant-456", key)
	// Tamper with the cookie value
	tampered := c.Value[:len(c.Value)-1] + "x"
	_, _, err := VerifyPendingSession(tampered, key)
	if err == nil {
		t.Error("VerifyPendingSession with tampered cookie should fail")
	}
}

func TestVerifyPendingSessionWrongKey(t *testing.T) {
	key1 := testHMACKey(t)
	key2 := testHMACKey(t)
	c, _ := PendingSessionCookie("user-123", "tenant-456", key1)
	_, _, err := VerifyPendingSession(c.Value, key2)
	if err == nil {
		t.Error("VerifyPendingSession with wrong key should fail")
	}
}

func TestIsSecure(t *testing.T) {
	if !IsSecure("https://prospeccao.com.br") {
		t.Error("https:// should be secure")
	}
	if IsSecure("http://localhost:8080") {
		t.Error("http:// should not be secure")
	}
}

func TestSessionExpiry(t *testing.T) {
	expiry := SessionExpiry()
	if !expiry.After(time.Now()) {
		t.Error("SessionExpiry should be in the future")
	}
}
