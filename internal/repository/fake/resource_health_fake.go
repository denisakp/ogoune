package fake

import (
	"context"
	"sync"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// ResourceHealthFake is an in-memory ResourceHealthRepository for tests. It
// mirrors the real repository's contract exactly: at most one record per monitor,
// replaced on upsert, and nil (not a zero-filled struct) when absent.
type ResourceHealthFake struct {
	mu      sync.RWMutex
	records map[string]*domain.ResourceHealth

	// UpsertErr, FindErr and DeleteErr, when set, make the matching call fail.
	// Lets a caller exercise the degraded path without a database.
	UpsertErr error
	FindErr   error
	DeleteErr error
}

func NewResourceHealthFake() *ResourceHealthFake {
	return &ResourceHealthFake{records: map[string]*domain.ResourceHealth{}}
}

func (r *ResourceHealthFake) Upsert(_ context.Context, h *domain.ResourceHealth) error {
	if r.UpsertErr != nil {
		return r.UpsertErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if h.CollectedAt.IsZero() {
		h.CollectedAt = time.Now().UTC()
	}
	cp := *h
	r.records[h.ResourceID] = &cp
	return nil
}

func (r *ResourceHealthFake) FindByResourceID(_ context.Context, resourceID string) (*domain.ResourceHealth, error) {
	if r.FindErr != nil {
		return nil, r.FindErr
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	rec, ok := r.records[resourceID]
	if !ok {
		return nil, nil
	}
	cp := *rec
	return &cp, nil
}

func (r *ResourceHealthFake) DeleteByResourceID(_ context.Context, resourceID string) error {
	if r.DeleteErr != nil {
		return r.DeleteErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.records, resourceID)
	return nil
}

// Count exposes how many records are held, so a test can assert that upsert
// replaces rather than accumulates.
func (r *ResourceHealthFake) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.records)
}
