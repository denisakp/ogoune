package crypto

import (
	"os"
	"testing"
)

const testKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	os.Setenv("APP_SECRET_KEY", testKey)
	defer os.Unsetenv("APP_SECRET_KEY")

	plain := "secret-token"
	encrypted, err := Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encrypted == plain {
		t.Fatalf("expected encrypted payload to differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != plain {
		t.Fatalf("expected %q, got %q", plain, decrypted)
	}
}

func TestEncryptDecryptEmptyInput(t *testing.T) {
	os.Setenv("APP_SECRET_KEY", testKey)
	defer os.Unsetenv("APP_SECRET_KEY")

	encrypted, err := Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt empty failed: %v", err)
	}
	if encrypted != "" {
		t.Fatalf("expected empty encrypted string, got %q", encrypted)
	}

	decrypted, err := Decrypt("")
	if err != nil {
		t.Fatalf("Decrypt empty failed: %v", err)
	}
	if decrypted != "" {
		t.Fatalf("expected empty decrypted string, got %q", decrypted)
	}
}

func TestEncryptMissingKey(t *testing.T) {
	os.Unsetenv("APP_SECRET_KEY")

	if _, err := Encrypt("value"); err == nil {
		t.Fatalf("expected error when APP_SECRET_KEY is missing")
	}
}

func TestEncryptIsNonDeterministic(t *testing.T) {
	os.Setenv("APP_SECRET_KEY", testKey)
	defer os.Unsetenv("APP_SECRET_KEY")

	c1, err := Encrypt("same-plaintext")
	if err != nil {
		t.Fatalf("Encrypt #1 failed: %v", err)
	}
	c2, err := Encrypt("same-plaintext")
	if err != nil {
		t.Fatalf("Encrypt #2 failed: %v", err)
	}

	if c1 == c2 {
		t.Fatalf("expected different ciphertext outputs for identical plaintext")
	}
}
