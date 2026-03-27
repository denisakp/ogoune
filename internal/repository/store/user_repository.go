package store

import (
	"context"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

// UserRepository implements the UserRepository interface for PostgreSQL
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if err == gorm.ErrDuplicatedKey {
			return nil, repository.ErrDuplicate
		}
		return nil, err
	}
	return user, nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return err
	}
	return nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// UpdatePassword updates only the password hash and marks password as initialized
func (r *UserRepository) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"hashed_password":       hashedPassword,
			"password_initialized":  true,
			"force_password_change": false,
		}).Error
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("last_login_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

// UpdateTwoFactorSecret updates the 2FA secret and enabled flag
func (r *UserRepository) UpdateTwoFactorSecret(ctx context.Context, userID string, secret string, enabled bool) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"two_factor_secret":  secret,
			"two_factor_enabled": enabled,
		}).Error
}
