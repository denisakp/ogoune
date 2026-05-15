package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/repository"
	"github.com/denisakp/ogoune/pkg/apikey"
)

const maxAPIKeysPerUser = 10

// AuthenticatedAPIKey carries validated API key and user data for middleware.
type AuthenticatedAPIKey struct {
	Key  *domain.APIKey
	User *domain.User
}

// APIKeyService orchestrates API key business logic.
type APIKeyService struct {
	repo     repository.APIKeyRepository
	userRepo repository.UserRepository
	now      func() time.Time
}

func NewAPIKeyService(repo repository.APIKeyRepository, userRepo repository.UserRepository) *APIKeyService {
	return &APIKeyService{
		repo:     repo,
		userRepo: userRepo,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// SetNow replaces the clock function used internally. Intended for tests only.
func (s *APIKeyService) SetNow(fn func() time.Time) { s.now = fn }

func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID, name string, scope domain.APIKeyScope, expiresAt *time.Time) (*dto.CreateAPIKeyResponse, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > 100 {
		return nil, ErrValidationFailed
	}
	if scope != domain.APIKeyScopeRead && scope != domain.APIKeyScopeReadWrite {
		return nil, ErrValidationFailed
	}
	if expiresAt != nil && !expiresAt.After(s.now()) {
		return nil, ErrValidationFailed
	}

	count, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= maxAPIKeysPerUser {
		return nil, ErrAPIKeyLimitReached
	}

	rawKey, keyHash, keyPrefix, err := apikey.Generate()
	if err != nil {
		return nil, err
	}

	key := &domain.APIKey{
		UserID:    userID,
		Name:      trimmedName,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		Scope:     scope,
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	if err := s.repo.Create(ctx, key); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrValidationFailed
		}
		return nil, err
	}

	return &dto.CreateAPIKeyResponse{
		ID:        key.ID,
		Name:      key.Name,
		Key:       rawKey,
		KeyPrefix: key.KeyPrefix,
		Scope:     key.Scope,
		ExpiresAt: key.ExpiresAt,
		CreatedAt: key.CreatedAt,
	}, nil
}

func (s *APIKeyService) ListAPIKeys(ctx context.Context, userID string) ([]dto.APIKeyListItem, error) {
	keys, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.APIKeyListItem, 0, len(keys))
	for _, key := range keys {
		items = append(items, dto.APIKeyListItem{
			ID:         key.ID,
			Name:       key.Name,
			KeyPrefix:  key.KeyPrefix,
			Scope:      key.Scope,
			ExpiresAt:  key.ExpiresAt,
			LastUsedAt: key.LastUsedAt,
			LastUsedIP: key.LastUsedIP,
			IsActive:   key.IsActive,
			CreatedAt:  key.CreatedAt,
		})
	}
	return items, nil
}

func (s *APIKeyService) RevokeAPIKey(ctx context.Context, userID, keyID string) error {
	if strings.TrimSpace(keyID) == "" {
		return ErrValidationFailed
	}
	if err := s.repo.Revoke(ctx, keyID, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrAPIKeyNotFound
		}
		return err
	}
	return nil
}

func (s *APIKeyService) AuthenticateAPIKey(ctx context.Context, rawKey string) (*AuthenticatedAPIKey, error) {
	if !apikey.IsAPIKeyFormat(rawKey) {
		return nil, ErrUnauthorized
	}

	keyHash := apikey.Hash(rawKey)
	key, err := s.repo.FindByKeyHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrAPIKeyInvalid
		}
		return nil, err
	}
	if !key.IsActive {
		return nil, ErrAPIKeyRevoked
	}
	if key.ExpiresAt != nil && !key.ExpiresAt.After(s.now()) {
		return nil, ErrAPIKeyExpired
	}

	user, err := s.userRepo.FindByID(ctx, key.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}

	return &AuthenticatedAPIKey{Key: key, User: user}, nil
}

func (s *APIKeyService) UpdateLastUsed(ctx context.Context, keyID, ip string) error {
	if strings.TrimSpace(keyID) == "" {
		return ErrValidationFailed
	}
	if err := s.repo.UpdateLastUsed(ctx, keyID, s.now(), strings.TrimSpace(ip)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrAPIKeyNotFound
		}
		return fmt.Errorf("update api key last used: %w", err)
	}
	return nil
}
