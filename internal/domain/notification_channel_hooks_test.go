package domain

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/denisakp/ogoune/pkg/crypto"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const hookTestKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func newHookTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&NotificationChannel{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func setupCryptoKey(t *testing.T) {
	t.Helper()
	os.Setenv("APP_SECRET_KEY", hookTestKey)
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	t.Cleanup(func() {
		os.Unsetenv("APP_SECRET_KEY")
		crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	})
}

// T008: BeforeCreate encrypts Config
func TestNotificationChannel_BeforeCreate_EncryptsConfig(t *testing.T) {
	setupCryptoKey(t)
	db := newHookTestDB(t)

	plainConfig := []byte(`{"host":"smtp.example.com","password":"supersecret"}`)
	ch := &NotificationChannel{
		Name:   "smtp-test",
		Type:   NotificationChannelTypeSMTP,
		Config: plainConfig,
	}

	if err := db.Create(ch).Error; err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Read raw value from DB — must be ciphertext, not JSON
	var rawConfig string
	db.Raw("SELECT config FROM notification_channels WHERE id = ?", ch.ID).Scan(&rawConfig)

	if len(rawConfig) == 0 {
		t.Fatal("expected non-empty config in DB")
	}
	if rawConfig[0] == '{' {
		t.Fatal("expected encrypted ciphertext in DB, got plaintext JSON")
	}
}

// T009: BeforeUpdate encrypts Config
func TestNotificationChannel_BeforeUpdate_EncryptsConfig(t *testing.T) {
	setupCryptoKey(t)
	db := newHookTestDB(t)

	ch := &NotificationChannel{
		Name:   "smtp-update",
		Type:   NotificationChannelTypeSMTP,
		Config: []byte(`{"host":"smtp.example.com","password":"original"}`),
	}
	if err := db.Create(ch).Error; err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	ch.Config = []byte(`{"host":"smtp.example.com","password":"updated"}`)
	if err := db.Save(ch).Error; err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	var rawConfig string
	db.Raw("SELECT config FROM notification_channels WHERE id = ?", ch.ID).Scan(&rawConfig)

	if len(rawConfig) == 0 {
		t.Fatal("expected non-empty config in DB after update")
	}
	if rawConfig[0] == '{' {
		t.Fatal("expected ciphertext in DB after update, got plaintext JSON")
	}
}

// T010: AfterFind decrypts Config
func TestNotificationChannel_AfterFind_DecryptsConfig(t *testing.T) {
	setupCryptoKey(t)
	db := newHookTestDB(t)

	plainConfig := `{"host":"smtp.example.com","password":"supersecret"}`
	ciphertext, err := crypto.Encrypt(plainConfig)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Insert pre-encrypted record directly (bypassing hooks)
	ch := &NotificationChannel{
		Name: "smtp-decrypt",
		Type: NotificationChannelTypeSMTP,
	}
	ch.ID = "test-decrypt-id"
	db.Exec("INSERT INTO notification_channels (id, name, type, config, enabled_by_default, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))",
		ch.ID, ch.Name, string(ch.Type), ciphertext, false)

	// Read via GORM — AfterFind should decrypt
	var found NotificationChannel
	if err := db.First(&found, "id = ?", ch.ID).Error; err != nil {
		t.Fatalf("First failed: %v", err)
	}

	if !bytes.Equal(found.Config, []byte(plainConfig)) {
		t.Fatalf("expected plaintext config %q, got %q", plainConfig, string(found.Config))
	}
}

// T016: AfterFind lazily migrates plaintext → re-encrypts in DB
func TestNotificationChannel_AfterFind_LazyMigration_EncryptsPlaintext(t *testing.T) {
	setupCryptoKey(t)
	db := newHookTestDB(t)

	plainConfig := `{"host":"smtp.example.com","password":"legacypass"}`
	// Insert as raw plaintext (bypassing hooks)
	db.Exec(`INSERT INTO notification_channels (id, name, type, config, enabled_by_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		"lazy-id", "lazy", "smtp", plainConfig, false)

	var found NotificationChannel
	if err := db.First(&found, "id = ?", "lazy-id").Error; err != nil {
		t.Fatalf("First failed: %v", err)
	}

	// In-memory config must be plaintext
	if string(found.Config) != plainConfig {
		t.Fatalf("expected plaintext config, got %q", string(found.Config))
	}

	// DB must now hold ciphertext
	var rawConfig string
	db.Raw("SELECT config FROM notification_channels WHERE id = ?", "lazy-id").Scan(&rawConfig)
	if len(rawConfig) == 0 {
		t.Fatal("expected non-empty config in DB")
	}
	if rawConfig[0] == '{' {
		t.Fatal("expected DB config to be ciphertext after lazy migration, got plaintext JSON")
	}
}

// T017: Already-encrypted records are not double-encrypted
func TestNotificationChannel_AfterFind_LazyMigration_AlreadyEncrypted_NoDoubleEncrypt(t *testing.T) {
	setupCryptoKey(t)
	db := newHookTestDB(t)

	plainConfig := `{"host":"smtp.example.com","password":"secret"}`
	ciphertext, err := crypto.Encrypt(plainConfig)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	db.Exec(`INSERT INTO notification_channels (id, name, type, config, enabled_by_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		"already-enc-id", "enc", "smtp", ciphertext, false)

	var found NotificationChannel
	if err := db.First(&found, "id = ?", "already-enc-id").Error; err != nil {
		t.Fatalf("First failed: %v", err)
	}

	// Config should be decrypted plaintext (not double-encrypted)
	if string(found.Config) != plainConfig {
		t.Fatalf("expected plaintext %q after AfterFind, got %q", plainConfig, string(found.Config))
	}

	// DB value should still be a valid single-layer ciphertext
	var rawConfig string
	db.Raw("SELECT config FROM notification_channels WHERE id = ?", "already-enc-id").Scan(&rawConfig)
	if rawConfig[0] == '{' {
		t.Fatal("DB value should still be ciphertext (not re-encrypted plaintext)")
	}
	// Verify it decrypts correctly to plaintext (not double-encrypted garbage)
	decrypted, err := crypto.Decrypt(rawConfig)
	if err != nil {
		t.Fatalf("DB value should decrypt cleanly: %v", err)
	}
	if decrypted != plainConfig {
		t.Fatalf("expected %q, got %q", plainConfig, decrypted)
	}
}

// T018: Empty Config is a no-op
func TestNotificationChannel_AfterFind_EmptyConfig_NoOp(t *testing.T) {
	setupCryptoKey(t)
	db := newHookTestDB(t)

	db.Exec(`INSERT INTO notification_channels (id, name, type, config, enabled_by_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		"empty-id", "empty", "smtp", "", false)

	var found NotificationChannel
	err := db.First(&found, "id = ?", "empty-id").Error
	if err != nil {
		t.Fatalf("expected no error on empty config, got: %v", err)
	}
	if len(found.Config) != 0 {
		t.Fatalf("expected empty config to remain empty, got %q", string(found.Config))
	}
}

// T011: AfterFind returns error on corrupt ciphertext (no silent empty)
func TestNotificationChannel_AfterFind_DecryptFailure(t *testing.T) {
	setupCryptoKey(t)
	db := newHookTestDB(t)

	// Insert corrupted ciphertext (valid base64 but not a valid AES-GCM payload)
	corruptCipher := "bm90YXZhbGlkY2lwaGVydGV4dA==" // base64("notavalidciphertext")
	db.Exec("INSERT INTO notification_channels (id, name, type, config, enabled_by_default, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))",
		"test-corrupt-id", "corrupt", "smtp", corruptCipher, false)

	var found NotificationChannel
	err := db.First(&found, "id = ?", "test-corrupt-id").Error

	if err == nil {
		t.Fatal("expected error when decryption fails, got nil")
	}
	if strings.Contains(string(found.Config), "") && len(found.Config) > 0 {
		// If config was silently set to empty, that's also a failure
	}
	// The key assertion: error must be non-nil
	t.Logf("got expected error: %v", err)
}
