package store

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/pkg/crypto"
)

const credentialTestKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func setupCredentialCryptoKey(t *testing.T) {
	t.Helper()
	t.Setenv("APP_SECRET_KEY", credentialTestKey)
	crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	t.Cleanup(func() {
		_ = os.Unsetenv("APP_SECRET_KEY")
		crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})
	})
}

func seedResourceForCred(t *testing.T, repo *ResourceCredentialRepository, resourceID string) {
	t.Helper()
	if err := repo.db.Exec(
		`INSERT INTO resources (id, name, type, target, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`,
		resourceID, "test", "protocol", "redis://localhost:6379",
	).Error; err != nil {
		t.Fatalf("seed resource: %v", err)
	}
}

func TestResourceCredentialRepository_UpsertAndGet(t *testing.T) {
	setupCredentialCryptoKey(t)
	db := internaltest.GetTestDB(t)
	repo := NewResourceCredentialRepository(db)
	ctx := context.Background()
	const resourceID = "01HXYZRESOURCE0000000000010"
	seedResourceForCred(t, repo, resourceID)

	cred := &domain.ResourceCredential{
		ResourceID: resourceID,
		Username:   "monitor",
		Password:   []byte("s3cret!"),
	}
	if err := repo.Upsert(ctx, cred); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := repo.Get(ctx, resourceID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(got.Password) != "s3cret!" || got.Username != "monitor" {
		t.Fatalf("decrypted mismatch: %+v", got)
	}
}

func TestResourceCredentialRepository_UpsertAtomicReplace(t *testing.T) {
	setupCredentialCryptoKey(t)
	db := internaltest.GetTestDB(t)
	repo := NewResourceCredentialRepository(db)
	ctx := context.Background()
	const resourceID = "01HXYZRESOURCE0000000000011"
	seedResourceForCred(t, repo, resourceID)

	if err := repo.Upsert(ctx, &domain.ResourceCredential{ResourceID: resourceID, Username: "old", Password: []byte("old-pass")}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := repo.Upsert(ctx, &domain.ResourceCredential{ResourceID: resourceID, Username: "new", Password: []byte("new-pass")}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	var rowCount int64
	if err := db.Model(&domain.ResourceCredential{}).Where("resource_id = ?", resourceID).Count(&rowCount).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("expected exactly 1 row after replace, got %d", rowCount)
	}

	got, err := repo.Get(ctx, resourceID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Username != "new" || string(got.Password) != "new-pass" {
		t.Fatalf("replace did not apply: %+v", got)
	}
}

func TestResourceCredentialRepository_DeleteCascade(t *testing.T) {
	setupCredentialCryptoKey(t)
	db := internaltest.GetTestDB(t)
	repo := NewResourceCredentialRepository(db)
	ctx := context.Background()
	const resourceID = "01HXYZRESOURCE0000000000012"
	seedResourceForCred(t, repo, resourceID)

	if err := repo.Upsert(ctx, &domain.ResourceCredential{ResourceID: resourceID, Password: []byte("p")}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Delete the parent resource — cascade should wipe the credential.
	if err := db.Exec("DELETE FROM resources WHERE id = ?", resourceID).Error; err != nil {
		t.Fatalf("delete resource: %v", err)
	}

	_, err := repo.Get(ctx, resourceID)
	if !errors.Is(err, repository.ErrCredentialNotFound) {
		t.Fatalf("expected ErrCredentialNotFound after cascade, got %v", err)
	}
}

func TestResourceCredentialRepository_DeleteAndExists(t *testing.T) {
	setupCredentialCryptoKey(t)
	db := internaltest.GetTestDB(t)
	repo := NewResourceCredentialRepository(db)
	ctx := context.Background()
	const resourceID = "01HXYZRESOURCE0000000000013"
	seedResourceForCred(t, repo, resourceID)

	exists, err := repo.Exists(ctx, resourceID)
	if err != nil || exists {
		t.Fatalf("Exists before upsert: exists=%v err=%v", exists, err)
	}

	if err := repo.Upsert(ctx, &domain.ResourceCredential{ResourceID: resourceID, Password: []byte("p")}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	exists, err = repo.Exists(ctx, resourceID)
	if err != nil || !exists {
		t.Fatalf("Exists after upsert: exists=%v err=%v", exists, err)
	}

	if err := repo.Delete(ctx, resourceID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := repo.Delete(ctx, resourceID); !errors.Is(err, repository.ErrCredentialNotFound) {
		t.Fatalf("second delete: expected ErrCredentialNotFound, got %v", err)
	}
}
