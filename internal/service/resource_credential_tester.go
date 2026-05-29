package service

import (
	"context"
	"errors"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
)

// ResourceCredentialTester runs a live check using a supplied credential without
// persisting it. It loads the resource, attaches the transient credential, and
// dispatches to the matching CheckStrategy.
type ResourceCredentialTester struct {
	resourceRepo port.ResourceRepository
	strategies   map[domain.ResourceType]domain.CheckStrategy
}

func NewResourceCredentialTester(resourceRepo port.ResourceRepository, strategies map[domain.ResourceType]domain.CheckStrategy) *ResourceCredentialTester {
	return &ResourceCredentialTester{resourceRepo: resourceRepo, strategies: strategies}
}

// Test executes the protocol strategy with the provided credential, never persisting it.
// Returns ErrResourceNotFound if the resource id is unknown.
func (t *ResourceCredentialTester) Test(ctx context.Context, resourceID string, username string, password []byte, options []byte) (domain.CheckResult, error) {
	resource, err := t.resourceRepo.FindByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.CheckResult{}, ErrResourceNotFound
		}
		return domain.CheckResult{}, err
	}

	// Shallow-copy so we don't mutate the cached resource attached to the repo.
	r := *resource
	r.Credential = &domain.ResourceCredential{
		ResourceID: resourceID,
		Username:   username,
		Password:   password,
		Options:    options,
	}

	strategy, ok := t.strategies[r.Type]
	if !ok {
		cause := domain.InvalidConfiguration
		return domain.CheckResult{
			Status:       string(domain.StatusDown),
			ResponseData: "no strategy registered for this resource type",
			Cause:        &cause,
		}, nil
	}
	return strategy.Execute(ctx, &r)
}
