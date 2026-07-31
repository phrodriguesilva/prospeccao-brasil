package auth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"image/png"
	"io"

	"github.com/pquerna/otp/totp"
)

// GenerateTOTPSecret generates a new TOTP secret and QR code PNG for the
// given email (used as account name). The secret is plaintext -- the caller
// must encrypt it before storing. The QR code is a base64-encoded PNG
// suitable for embedding in HTML as <img src="data:image/png;base64,...">.
func GenerateTOTPSecret(email string) (secret string, qrPNG []byte, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Prospeccao Brasil",
		AccountName: email,
	})
	if err != nil {
		return "", nil, fmt.Errorf("generate totp key: %w", err)
	}
	img, err := key.Image(200, 200)
	if err != nil {
		return "", nil, fmt.Errorf("generate totp qr: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", nil, fmt.Errorf("encode qr png: %w", err)
	}
	return key.Secret(), buf.Bytes(), nil
}

// EncryptSecret encrypts a plaintext TOTP secret using AES-256-GCM.
// The key must be 32 bytes (AES-256). Returns base64-encoded ciphertext
// (nonce prepended).
func EncryptSecret(secret string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("encrypt secret: key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("encrypt secret: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("encrypt secret: new gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("encrypt secret: nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSecret decrypts a base64-encoded AES-256-GCM ciphertext.
// The key must be 32 bytes (AES-256) and match the key used for encryption.
func DecryptSecret(encrypted string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("decrypt secret: key must be 32 bytes, got %d", len(key))
	}
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt secret: base64 decode: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("decrypt secret: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("decrypt secret: new gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("decrypt secret: ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt secret: gcm open: %w", err)
	}
	return string(plaintext), nil
}

// ValidateTOTP decrypts the encrypted secret and validates the TOTP code.
// Allows +/- 1 time step (30 seconds) for clock drift.
func ValidateTOTP(code, encryptedSecret string, key []byte) bool {
	secret, err := DecryptSecret(encryptedSecret, key)
	if err != nil {
		return false
	}
	return totp.Validate(code, secret)
}
