package crypto

import (
	"os"
	"strings"
	"testing"
)

func setKey(t *testing.T) {
	t.Helper()
	t.Setenv("APP_SECRET_KEY", testKey)
}

func TestChannel_RoundTrip(t *testing.T) {
	setKey(t)
	cases := []string{
		`{"webhook_url":"https://example.invalid"}`,
		"",
		"unicode 🌍 résumé",
		strings.Repeat("x", 4096),
	}
	for _, plain := range cases {
		cipher, err := EncryptChannelConfig(plain)
		if err != nil {
			t.Fatalf("EncryptChannelConfig(%q) failed: %v", plain, err)
		}
		got, err := DecryptChannelConfig(cipher)
		if err != nil {
			t.Fatalf("DecryptChannelConfig failed: %v", err)
		}
		if got != plain {
			t.Errorf("round-trip mismatch: want %q got %q", plain, got)
		}
	}
}

// TestChannel_OracleAgainstGeneric verifies the typed wrappers and the
// generic primitive interoperate. This is the safety net against silent
// divergence inside the wrappers' bodies.
func TestChannel_OracleAgainstGeneric(t *testing.T) {
	setKey(t)
	plain := `{"webhook_url":"https://example.invalid","token":"sekret"}`

	// Generic encrypt → typed decrypt.
	cipherFromGeneric, err := Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	got, err := DecryptChannelConfig(cipherFromGeneric)
	if err != nil {
		t.Fatalf("DecryptChannelConfig(generic ciphertext) failed: %v", err)
	}
	if got != plain {
		t.Errorf("generic→typed mismatch: want %q got %q", plain, got)
	}

	// Typed encrypt → generic decrypt.
	cipherFromTyped, err := EncryptChannelConfig(plain)
	if err != nil {
		t.Fatalf("EncryptChannelConfig failed: %v", err)
	}
	got2, err := Decrypt(cipherFromTyped)
	if err != nil {
		t.Fatalf("Decrypt(typed ciphertext) failed: %v", err)
	}
	if got2 != plain {
		t.Errorf("typed→generic mismatch: want %q got %q", plain, got2)
	}
}

func TestChannel_MalformedCiphertext(t *testing.T) {
	setKey(t)
	if _, err := DecryptChannelConfig("not-a-real-ciphertext"); err == nil {
		t.Error("expected error on malformed ciphertext, got nil")
	}
}

// reference os import to avoid unused dep warnings.
var _ = os.Getenv
