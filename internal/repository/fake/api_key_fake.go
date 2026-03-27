package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

// APIKeyRepository is an in-memory API key repository for tests.
type APIKeyRepository struct {
	mu     sync.RWMutex
	byID   map[string]*domain.APIKey
	byHash map[string]string
}

func NewAPIKeyRepository() *APIKeyRepository {
	return &APIKeyRepository{
		byID:   make(map[string]*domain.APIKey),
		byHash: make(map[string]string),
	}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := key.BeforeCreate(nil); err != nil {
		return ErrInvalidInput
	}
	if _, exists := r.byID[key.ID]; exists {
		return ErrDuplicate
	}
	if _, exists := r.byHash[key.KeyHash]; exists {
		return ErrDuplicate
	}

	copy := *key
	r.byID[key.ID] = &copy
	r.byHash[key.KeyHash] = key.ID
	return nil
}

func (r *APIKeyRepository) FindByID(ctx context.Context, id, userID string) (*domain.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key, ok := r.byID[id]
	if !ok || key.UserID != userID {
		return nil, ErrNotFound
	}
	copy := *key
	return &copy, nil
}

func (r *APIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byHash[keyHash]
	if !ok {
		return nil, ErrNotFound
	}
	key, ok := r.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *key
	return &copy, nil
}

func (r *APIKeyRepository) ListByUserID(ctx context.Context, userID string) ([]domain.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]domain.APIKey, 0)
	for _, key := range r.byID {
		if key.UserID == userID {
			keys = append(keys, *key)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].CreatedAt.After(keys[j].CreatedAt)
	})
	return keys, nil
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id string, at time.Time, ip string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key, ok := r.byID[id]
	if !ok {
		return ErrNotFound
	}
	key.LastUsedAt = &at
	key.LastUsedIP = ip
	key.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key, ok := r.byID[id]
	if !ok || key.UserID != userID {
		return ErrNotFound
	}
	key.IsActive = false
	key.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *APIKeyRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, key := range r.byID {
		if key.UserID == userID {
			count++
		}
	}
	return count, nil
}
