package service

import (
	"context"
	"errors"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
)

// ResourceCredentialService manages optional auth credentials for protocol-aware resources.
// Encryption and decryption happen transparently in the domain layer via GORM hooks.
// Audit logging is the caller's responsibility (HTTP handler), to keep the service free
// of HTTP-context concerns per constitution principle I (Layered Boundary Integrity).
type ResourceCredentialService struct {
	repo         port.ResourceCredentialRepository
	resourceRepo port.ResourceRepository
}

func NewResourceCredentialService(repo port.ResourceCredentialRepository, resourceRepo port.ResourceRepository) *ResourceCredentialService {
	return &ResourceCredentialService{repo: repo, resourceRepo: resourceRepo}
}

// Get returns the credential for the given resource (password/options decrypted).
// Returns ErrResourceNotFound if the resource does not exist.
// Returns ErrCredentialNotFound if the resource has no credential row.
func (s *ResourceCredentialService) Get(ctx context.Context, resourceID string) (*domain.ResourceCredential, error) {
	if _, err := s.resourceRepo.FindByID(ctx, resourceID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}
	cred, err := s.repo.Get(ctx, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrCredentialNotFound) {
			return nil, ErrCredentialNotFound
		}
		return nil, err
	}
	return cred, nil
}

// Set creates or atomically replaces the credential for resourceID.
// Returns (created bool, err error) — created=true on insert, false on replace.
// Validates: password required and non-empty; username ≤ 128 chars.
func (s *ResourceCredentialService) Set(ctx context.Context, resourceID, username string, password []byte, options []byte) (bool, error) {
	if len(password) == 0 {
		return false, ErrCredentialInvalid
	}
	if len(username) > 128 {
		return false, ErrCredentialInvalid
	}
	if _, err := s.resourceRepo.FindByID(ctx, resourceID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return false, ErrResourceNotFound
		}
		return false, err
	}
	existed, err := s.repo.Exists(ctx, resourceID)
	if err != nil {
		return false, err
	}
	cred := &domain.ResourceCredential{
		ResourceID: resourceID,
		Username:   username,
		Password:   password,
		Options:    options,
	}
	if err := s.repo.Upsert(ctx, cred); err != nil {
		return false, err
	}
	return !existed, nil
}

// Delete removes the credential for resourceID.
// Returns ErrResourceNotFound if resource missing, ErrCredentialNotFound if no credential row.
func (s *ResourceCredentialService) Delete(ctx context.Context, resourceID string) error {
	if _, err := s.resourceRepo.FindByID(ctx, resourceID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrResourceNotFound
		}
		return err
	}
	if err := s.repo.Delete(ctx, resourceID); err != nil {
		if errors.Is(err, repository.ErrCredentialNotFound) {
			return ErrCredentialNotFound
		}
		return err
	}
	return nil
}
