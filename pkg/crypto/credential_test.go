package crypto

import (
	"testing"
)

func TestCredentialPassword_RoundTrip(t *testing.T) {
	t.Setenv("APP_SECRET_KEY", testKey)
	plain := "s3cr3t-p@ssw0rd"

	cipher, err := EncryptCredentialPassword(plain)
	if err != nil {
		t.Fatalf("EncryptCredentialPassword failed: %v", err)
	}
	got, err := DecryptCredentialPassword(cipher)
	if err != nil {
		t.Fatalf("DecryptCredentialPassword failed: %v", err)
	}
	if got != plain {
		t.Errorf("password round-trip mismatch: want %q got %q", plain, got)
	}
}

func TestCredentialOptions_RoundTrip(t *testing.T) {
	t.Setenv("APP_SECRET_KEY", testKey)
	plain := `{"tls":true,"db":0,"ca_cert":"-----BEGIN CERTIFICATE-----..."}`

	cipher, err := EncryptCredentialOptions(plain)
	if err != nil {
		t.Fatalf("EncryptCredentialOptions failed: %v", err)
	}
	got, err := DecryptCredentialOptions(cipher)
	if err != nil {
		t.Fatalf("DecryptCredentialOptions failed: %v", err)
	}
	if got != plain {
		t.Errorf("options round-trip mismatch: want %q got %q", plain, got)
	}
}

func TestCredential_OracleAgainstGeneric(t *testing.T) {
	t.Setenv("APP_SECRET_KEY", testKey)
	plainPwd := "p@ssword"
	plainOpts := `{"db":3}`

	// Password: generic encrypt → typed decrypt.
	pwdCipher, err := Encrypt(plainPwd)
	if err != nil {
		t.Fatalf("Encrypt(password) failed: %v", err)
	}
	gotPwd, err := DecryptCredentialPassword(pwdCipher)
	if err != nil {
		t.Fatalf("DecryptCredentialPassword(generic) failed: %v", err)
	}
	if gotPwd != plainPwd {
		t.Errorf("password generic→typed mismatch: want %q got %q", plainPwd, gotPwd)
	}

	// Options: typed encrypt → generic decrypt.
	optsCipher, err := EncryptCredentialOptions(plainOpts)
	if err != nil {
		t.Fatalf("EncryptCredentialOptions failed: %v", err)
	}
	gotOpts, err := Decrypt(optsCipher)
	if err != nil {
		t.Fatalf("Decrypt(typed options) failed: %v", err)
	}
	if gotOpts != plainOpts {
		t.Errorf("options typed→generic mismatch: want %q got %q", plainOpts, gotOpts)
	}
}

func TestCredential_MalformedCiphertext(t *testing.T) {
	t.Setenv("APP_SECRET_KEY", testKey)
	if _, err := DecryptCredentialPassword("not-real"); err == nil {
		t.Error("expected error on malformed password ciphertext")
	}
	if _, err := DecryptCredentialOptions("not-real"); err == nil {
		t.Error("expected error on malformed options ciphertext")
	}
}
