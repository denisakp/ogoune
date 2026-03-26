package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

// APIKeyRepositoryImpl implements API key persistence using GORM.
type APIKeyRepositoryImpl struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates an API key repository.
func NewAPIKeyRepository(db *gorm.DB) repository.APIKeyRepository {
	return &APIKeyRepositoryImpl{db: db}
}

func (r *APIKeyRepositoryImpl) Create(ctx context.Context, key *domain.APIKey) error {
	if err := r.db.WithContext(ctx).Create(key).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return repository.ErrDuplicate
		}
		return fmt.Errorf("failed to create api key: %w", err)
	}
	return nil
}

func (r *APIKeyRepositoryImpl) FindByID(ctx context.Context, id, userID string) (*domain.APIKey, error) {
	var key domain.APIKey
	err := r.db.WithContext(ctx).First(&key, "id = ? AND user_id = ?", id, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find api key by id: %w", err)
	}
	return &key, nil
}

func (r *APIKeyRepositoryImpl) FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	var key domain.APIKey
	err := r.db.WithContext(ctx).First(&key, "key_hash = ?", keyHash).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find api key by hash: %w", err)
	}
	return &key, nil
}

func (r *APIKeyRepositoryImpl) ListByUserID(ctx context.Context, userID string) ([]domain.APIKey, error) {
	var keys []domain.APIKey
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}
	return keys, nil
}

func (r *APIKeyRepositoryImpl) UpdateLastUsed(ctx context.Context, id string, at time.Time, ip string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.APIKey{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"last_used_at": at, "last_used_ip": ip})
	if result.Error != nil {
		return fmt.Errorf("failed to update last used metadata: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *APIKeyRepositoryImpl) Revoke(ctx context.Context, id, userID string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.APIKey{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("failed to revoke api key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *APIKeyRepositoryImpl) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.APIKey{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count api keys: %w", err)
	}
	return count, nil
}
