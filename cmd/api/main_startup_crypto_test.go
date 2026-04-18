package main

import (
	"os"
	"testing"

	"github.com/denisakp/ogoune/pkg/crypto"
)

const validTestKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// T021: Missing APP_SECRET_KEY causes ValidateKey to return error.
func TestStartup_MissingAppSecretKey(t *testing.T) {
	os.Unsetenv("APP_SECRET_KEY")
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})

	if err := crypto.ValidateKey(); err == nil {
		t.Fatal("expected error when APP_SECRET_KEY is missing, got nil")
	}
}

// T022: Malformed APP_SECRET_KEY (too short) causes ValidateKey to return error.
func TestStartup_MalformedAppSecretKey(t *testing.T) {
	os.Setenv("APP_SECRET_KEY", "tooshort")
	defer os.Unsetenv("APP_SECRET_KEY")
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})

	if err := crypto.ValidateKey(); err == nil {
		t.Fatal("expected error when APP_SECRET_KEY is malformed, got nil")
	}
}

// T023: Valid APP_SECRET_KEY allows ValidateKey to return nil (startup proceeds).
func TestStartup_ValidAppSecretKey_Proceeds(t *testing.T) {
	os.Setenv("APP_SECRET_KEY", validTestKey)
	defer os.Unsetenv("APP_SECRET_KEY")
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})

	if err := crypto.ValidateKey(); err != nil {
		t.Fatalf("expected no error with valid APP_SECRET_KEY, got: %v", err)
	}
}
