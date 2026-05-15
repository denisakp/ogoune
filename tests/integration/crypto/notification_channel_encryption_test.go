package crypto_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/store"
	"github.com/denisakp/ogoune/pkg/crypto"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const integTestKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func newIntegDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&domain.NotificationChannel{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func setupIntegCrypto(t *testing.T) {
	t.Helper()
	os.Setenv("APP_SECRET_KEY", integTestKey)
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	t.Cleanup(func() {
		os.Unsetenv("APP_SECRET_KEY")
		crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	})
}

// T012: Create channel → raw DB value is ciphertext → read via repo → decrypted config usable
func TestNotificationChannelEncryption_EndToEnd(t *testing.T) {
	setupIntegCrypto(t)
	db := newIntegDB(t)
	repo := store.NewNotificationChannelRepository(db)
	ctx := context.Background()

	plainConfig := []byte(`{"host":"smtp.example.com","port":587,"password":"supersecret"}`)
	ch := &domain.NotificationChannel{
		Name:   "e2e-smtp",
		Type:   domain.NotificationChannelTypeSMTP,
		Config: plainConfig,
	}

	if err := repo.Create(ctx, ch); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Assert raw DB value is ciphertext (not JSON)
	var rawConfig string
	db.Raw("SELECT config FROM notification_channels WHERE id = ?", ch.ID).Scan(&rawConfig)
	if len(rawConfig) == 0 {
		t.Fatal("raw DB config must not be empty")
	}
	if rawConfig[0] == '{' {
		t.Fatal("raw DB config should be ciphertext, not plaintext JSON")
	}

	// Read via repo — AfterFind should decrypt
	found, err := repo.FindByID(ctx, ch.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if !bytes.Equal(found.Config, plainConfig) {
		t.Fatalf("expected plaintext config %q, got %q", plainConfig, found.Config)
	}
}

// T019 (integration): seed plaintext → FindByID triggers migration → raw DB now ciphertext
func TestNotificationChannelEncryption_LazyMigration_E2E(t *testing.T) {
	setupIntegCrypto(t)
	db := newIntegDB(t)
	repo := store.NewNotificationChannelRepository(db)
	ctx := context.Background()

	// Seed a legacy plaintext record directly (bypassing GORM hooks)
	plainConfig := `{"host":"legacy.example.com","password":"legacypass"}`
	db.Exec(`INSERT INTO notification_channels (id, name, type, config, enabled_by_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		"legacy-id", "legacy-smtp", "smtp", plainConfig, false)

	// Read via repo — AfterFind should lazily migrate
	found, err := repo.FindByID(ctx, "legacy-id")
	if err != nil {
		t.Fatalf("FindByID failed after lazy migration: %v", err)
	}

	// In-memory config should be usable plaintext
	if string(found.Config) != plainConfig {
		t.Fatalf("expected plaintext config %q, got %q", plainConfig, string(found.Config))
	}

	// Raw DB should now hold ciphertext
	var rawConfig string
	db.Raw("SELECT config FROM notification_channels WHERE id = ?", "legacy-id").Scan(&rawConfig)
	if len(rawConfig) == 0 {
		t.Fatal("raw DB config must not be empty after lazy migration")
	}
	if rawConfig[0] == '{' {
		t.Fatal("after lazy migration, raw DB config should be ciphertext, not plaintext JSON")
	}
}
