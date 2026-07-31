package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("test123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$2a$") {
		t.Errorf("hash should start with $2a$, got %q", hash[:4])
	}
}

func TestHashPasswordNonDeterministic(t *testing.T) {
	h1, _ := HashPassword("test123")
	h2, _ := HashPassword("test123")
	if h1 == h2 {
		t.Error("HashPassword should be non-deterministic (different salts)")
	}
}

func TestVerifyPasswordCorrect(t *testing.T) {
	hash, _ := HashPassword("test123")
	if err := VerifyPassword(hash, "test123"); err != nil {
		t.Errorf("VerifyPassword with correct password: %v", err)
	}
}

func TestVerifyPasswordWrong(t *testing.T) {
	hash, _ := HashPassword("test123")
	if err := VerifyPassword(hash, "wrong"); err == nil {
		t.Error("VerifyPassword with wrong password should fail")
	}
}
