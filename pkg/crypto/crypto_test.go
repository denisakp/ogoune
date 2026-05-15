package crypto

import (
	"os"
	"testing"
)

const testKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
const testKey2 = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

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

// ---- T004: KeyProvider tests ----

func TestEnvKeyProvider_ValidKey(t *testing.T) {
	os.Setenv("APP_SECRET_KEY", testKey)
	defer os.Unsetenv("APP_SECRET_KEY")

	p := &EnvKeyProvider{}
	key, err := p.GetEncryptionKey()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-byte key, got %d bytes", len(key))
	}
}

func TestEnvKeyProvider_MissingKey(t *testing.T) {
	os.Unsetenv("APP_SECRET_KEY")

	p := &EnvKeyProvider{}
	_, err := p.GetEncryptionKey()
	if err == nil {
		t.Fatal("expected error for missing APP_SECRET_KEY")
	}
}

func TestEnvKeyProvider_WrongLength(t *testing.T) {
	os.Setenv("APP_SECRET_KEY", "tooshort")
	defer os.Unsetenv("APP_SECRET_KEY")

	p := &EnvKeyProvider{}
	_, err := p.GetEncryptionKey()
	if err == nil {
		t.Fatal("expected error for wrong-length APP_SECRET_KEY")
	}
}

func TestEnvKeyProvider_InvalidHex(t *testing.T) {
	os.Setenv("APP_SECRET_KEY", "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	defer os.Unsetenv("APP_SECRET_KEY")

	p := &EnvKeyProvider{}
	_, err := p.GetEncryptionKey()
	if err == nil {
		t.Fatal("expected error for non-hex APP_SECRET_KEY")
	}
}

func TestGlobalProvider_SetAndGet(t *testing.T) {
	original := GlobalProvider()
	defer SetGlobalProvider(original)

	custom := &EnvKeyProvider{}
	SetGlobalProvider(custom)
	if GlobalProvider() != custom {
		t.Fatal("expected GlobalProvider to return the custom provider")
	}
}

func TestValidateKey_Valid(t *testing.T) {
	os.Setenv("APP_SECRET_KEY", testKey)
	defer os.Unsetenv("APP_SECRET_KEY")

	SetGlobalProvider(&EnvKeyProvider{})
	if err := ValidateKey(); err != nil {
		t.Fatalf("expected no error with valid key, got: %v", err)
	}
}

func TestValidateKey_Missing(t *testing.T) {
	os.Unsetenv("APP_SECRET_KEY")
	SetGlobalProvider(&EnvKeyProvider{})

	if err := ValidateKey(); err == nil {
		t.Fatal("expected error with missing key")
	}
}

// ---- T005: wrong-key decryption failure ----

func TestDecrypt_WrongKey(t *testing.T) {
	// Encrypt with testKey
	os.Setenv("APP_SECRET_KEY", testKey)
	SetGlobalProvider(&EnvKeyProvider{})
	ciphertext, err := Encrypt("my-secret")
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Attempt decrypt with testKey2
	os.Setenv("APP_SECRET_KEY", testKey2)
	SetGlobalProvider(&EnvKeyProvider{})
	_, err = Decrypt(ciphertext)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}

	// Restore
	os.Unsetenv("APP_SECRET_KEY")
	SetGlobalProvider(&EnvKeyProvider{})
}
