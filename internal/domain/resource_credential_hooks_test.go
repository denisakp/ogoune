package domain

import (
	"bytes"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newCredentialTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&ResourceCredential{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestResourceCredential_RoundTrip(t *testing.T) {
	setupCryptoKey(t)
	db := newCredentialTestDB(t)

	plainPassword := []byte("s3cret!")
	cred := &ResourceCredential{
		ResourceID: "01HXYZRESOURCE0000000000001",
		Username:   "monitor",
		Password:   bytes.Clone(plainPassword),
	}
	if err := db.Create(cred).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	// Confirm ciphertext is stored (not plaintext).
	var raw struct {
		Password []byte
	}
	if err := db.Raw("SELECT password FROM resource_credentials WHERE resource_id = ?", cred.ResourceID).Scan(&raw).Error; err != nil {
		t.Fatalf("raw scan: %v", err)
	}
	if bytes.Equal(raw.Password, plainPassword) {
		t.Fatalf("password stored in plaintext")
	}

	// AfterFind decrypts back to plaintext.
	var got ResourceCredential
	if err := db.First(&got, "resource_id = ?", cred.ResourceID).Error; err != nil {
		t.Fatalf("find: %v", err)
	}
	if !bytes.Equal(got.Password, plainPassword) {
		t.Fatalf("decrypted password mismatch: got %q want %q", got.Password, plainPassword)
	}
	if got.Username != "monitor" {
		t.Fatalf("username mismatch: got %q", got.Username)
	}
}

func TestResourceCredential_AfterFind_DecryptFailure(t *testing.T) {
	setupCryptoKey(t)
	db := newCredentialTestDB(t)

	// Insert a row with garbage ciphertext directly.
	if err := db.Exec(`INSERT INTO resource_credentials (id, resource_id, username, password, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`,
		"01HXYZGARBAGE0000000000001", "01HXYZRESOURCE0000000000002", "", []byte("not-real-ciphertext")).Error; err != nil {
		t.Fatalf("insert garbage: %v", err)
	}

	var got ResourceCredential
	err := db.First(&got, "resource_id = ?", "01HXYZRESOURCE0000000000002").Error
	if !errors.Is(err, ErrCredentialDecryption) {
		t.Fatalf("expected ErrCredentialDecryption, got %v", err)
	}
}

func TestResourceCredential_BeforeUpdate_ReEncrypts(t *testing.T) {
	setupCryptoKey(t)
	db := newCredentialTestDB(t)

	cred := &ResourceCredential{
		ResourceID: "01HXYZRESOURCE0000000000003",
		Password:   []byte("first"),
	}
	if err := db.Create(cred).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	// Re-fetch (decrypted), update plaintext, save.
	if err := db.First(cred, "resource_id = ?", cred.ResourceID).Error; err != nil {
		t.Fatalf("find: %v", err)
	}
	cred.Password = []byte("second")
	if err := db.Save(cred).Error; err != nil {
		t.Fatalf("save: %v", err)
	}

	var fresh ResourceCredential
	if err := db.First(&fresh, "resource_id = ?", cred.ResourceID).Error; err != nil {
		t.Fatalf("refind: %v", err)
	}
	if string(fresh.Password) != "second" {
		t.Fatalf("expected 'second', got %q", fresh.Password)
	}
}
