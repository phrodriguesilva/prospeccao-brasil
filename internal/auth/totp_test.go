package auth

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

// totpCode generates a valid TOTP code for the given secret at the given time.
// Used by auth_test.go for integration tests.
func totpCode(secret string, now time.Time) string {
	code, err := totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    6,
		Algorithm: 0,
	})
	if err != nil {
		return ""
	}
	return code
}

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate test key: %v", err)
	}
	return key
}

func TestGenerateTOTPSecret(t *testing.T) {
	secret, qrPNG, err := GenerateTOTPSecret("admin@prospeccao.com.br")
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	if secret == "" {
		t.Error("secret should not be empty")
	}
	if len(qrPNG) == 0 {
		t.Error("qrPNG should not be empty")
	}
}

func TestEncryptDecryptSecretRoundTrip(t *testing.T) {
	key := testKey(t)
	original := "JBSWY3DPEHPK3PXP"
	encrypted, err := EncryptSecret(original, key)
	if err != nil {
		t.Fatalf("EncryptSecret: %v", err)
	}
	if encrypted == original {
		t.Error("encrypted should differ from plaintext")
	}
	decrypted, err := DecryptSecret(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptSecret: %v", err)
	}
	if decrypted != original {
		t.Errorf("round-trip: got %q, want %q", decrypted, original)
	}
}

func TestDecryptSecretWrongKey(t *testing.T) {
	key1 := testKey(t)
	key2 := testKey(t)
	encrypted, _ := EncryptSecret("secret", key1)
	_, err := DecryptSecret(encrypted, key2)
	if err == nil {
		t.Error("DecryptSecret with wrong key should fail")
	}
}

func TestEncryptSecretWrongKeySize(t *testing.T) {
	_, err := EncryptSecret("test", []byte("short"))
	if err == nil {
		t.Error("EncryptSecret with short key should fail")
	}
}

func TestValidateTOTPValidCode(t *testing.T) {
	key := testKey(t)
	// Generate a real TOTP secret
	key2, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Test",
		AccountName: "test@prospeccao.com.br",
	})
	if err != nil {
		t.Fatalf("totp.Generate: %v", err)
	}
	encrypted, err := EncryptSecret(key2.Secret(), key)
	if err != nil {
		t.Fatalf("EncryptSecret: %v", err)
	}
	// Generate a valid code for now
	code := key2.Secret()
	_ = code
	// Use totp.GenerateCode with the current time
	validCode, err := totp.GenerateCodeCustom(key2.Secret(), time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    6,
		Algorithm: 0,
	})
	if err != nil {
		t.Fatalf("totp.GenerateCodeCustom: %v", err)
	}
	if !ValidateTOTP(validCode, encrypted, key) {
		t.Error("ValidateTOTP with valid code should return true")
	}
}

func TestValidateTOTPInvalidCode(t *testing.T) {
	key := testKey(t)
	encrypted, err := EncryptSecret("JBSWY3DPEHPK3PXP", key)
	if err != nil {
		t.Fatalf("EncryptSecret: %v", err)
	}
	if ValidateTOTP("000000", encrypted, key) {
		t.Error("ValidateTOTP with invalid code should return false")
	}
}

func TestValidateTOTPInvalidEncryptedSecret(t *testing.T) {
	key := testKey(t)
	// Pass invalid encrypted secret -- DecryptSecret will fail, ValidateTOTP returns false
	if ValidateTOTP("123456", "not-valid-base64!!!", key) {
		t.Error("ValidateTOTP with invalid encrypted secret should return false")
	}
}

func TestDecryptSecretInvalidBase64(t *testing.T) {
	key := testKey(t)
	_, err := DecryptSecret("!!!not-base64!!!", key)
	if err == nil {
		t.Error("DecryptSecret with invalid base64 should fail")
	}
}

func TestDecryptSecretShortCiphertext(t *testing.T) {
	key := testKey(t)
	// Valid base64 but too short to contain nonce
	short := "AAAA"
	_, err := DecryptSecret(short, key)
	if err == nil {
		t.Error("DecryptSecret with short ciphertext should fail")
	}
}

func TestDecryptSecretWrongKeySize(t *testing.T) {
	_, err := DecryptSecret("AAAA", []byte("short"))
	if err == nil {
		t.Error("DecryptSecret with short key should fail")
	}
}
