package store

import (
	"context"
	"errors"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const whereResourceID = "resource_id = ?"

// ResourceCredentialRepository persists optional auth credentials for protocol-aware resources.
type ResourceCredentialRepository struct {
	db *gorm.DB
}

func NewResourceCredentialRepository(db *gorm.DB) *ResourceCredentialRepository {
	return &ResourceCredentialRepository{db: db}
}

// Get returns the credential for the given resource. ErrCredentialNotFound when no row exists.
func (r *ResourceCredentialRepository) Get(ctx context.Context, resourceID string) (*domain.ResourceCredential, error) {
	var cred domain.ResourceCredential
	if err := r.db.WithContext(ctx).First(&cred, whereResourceID, resourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrCredentialNotFound
		}
		if errors.Is(err, domain.ErrCredentialDecryption) {
			return nil, repository.ErrCredentialDecryption
		}
		return nil, err
	}
	return &cred, nil
}

// Upsert atomically creates or replaces the credential for cred.ResourceID.
// GORM's OnConflict clause maps to "INSERT ... ON CONFLICT DO UPDATE" on Postgres
// and "INSERT OR REPLACE" semantics on SQLite via UpdateAll.
func (r *ResourceCredentialRepository) Upsert(ctx context.Context, cred *domain.ResourceCredential) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "resource_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"username", "password", "options", "updated_at"}),
	}).Create(cred).Error
}

// Delete removes the credential for the given resource. ErrCredentialNotFound when no row exists.
func (r *ResourceCredentialRepository) Delete(ctx context.Context, resourceID string) error {
	res := r.db.WithContext(ctx).Where(whereResourceID, resourceID).Delete(&domain.ResourceCredential{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrCredentialNotFound
	}
	return nil
}

// Exists returns true when a credential row exists for the given resource.
func (r *ResourceCredentialRepository) Exists(ctx context.Context, resourceID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.ResourceCredential{}).
		Where(whereResourceID, resourceID).Count(&count).Error
	return count > 0, err
}
