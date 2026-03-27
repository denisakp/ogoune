package fake

import (
	"context"
	"sync"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
)

// UserRepository implements the UserRepository interface with in-memory storage
type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

// NewUserRepository creates a new fake user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*domain.User),
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate email
	for _, u := range r.users {
		if u.Email == user.Email {
			return nil, repository.ErrDuplicate
		}
	}

	// Create a copy
	copy := *user
	r.users[copy.ID] = &copy
	return &copy, nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, repository.ErrNotFound
	}

	copy := *user
	return &copy, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			copy := *user
			return &copy, nil
		}
	}

	return nil, repository.ErrNotFound
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return repository.ErrNotFound
	}

	copy := *user
	copy.UpdatedAt = time.Now()
	r.users[user.ID] = &copy
	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[id]; !exists {
		return repository.ErrNotFound
	}

	delete(r.users, id)
	return nil
}

// UpdatePassword updates the password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return repository.ErrNotFound
	}

	user.HashedPassword = hashedPassword
	user.PasswordInitialized = true
	user.ForcePasswordChange = false
	user.UpdatedAt = time.Now()
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return repository.ErrNotFound
	}

	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now
	return nil
}

// UpdateTwoFactorSecret updates the 2FA secret
func (r *UserRepository) UpdateTwoFactorSecret(ctx context.Context, userID string, secret string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return repository.ErrNotFound
	}

	user.TwoFactorSecret = secret
	user.TwoFactorEnabled = enabled
	user.UpdatedAt = time.Now()
	return nil
}
