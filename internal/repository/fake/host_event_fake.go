package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// HostEventFake is an in-memory HostEventRepository for tests.
type HostEventFake struct {
	mu     sync.RWMutex
	events []*domain.HostEvent

	// CreateErr, when set, makes Create fail — so an ingestion path can be tested
	// against a storage failure without a database.
	CreateErr error
	// ListInWindowErr, when set, makes ListInWindow fail, so a correlation
	// failure path can be tested without a database.
	ListInWindowErr error
	// ListInWindowCalls counts the reads. A test proving the zero-cost guarantee
	// asserts this stays at zero for a monitor with no host attached.
	ListInWindowCalls int
}

func NewHostEventFake() *HostEventFake {
	return &HostEventFake{}
}

func (r *HostEventFake) Create(_ context.Context, e *domain.HostEvent) error {
	if r.CreateErr != nil {
		return r.CreateErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	e.EnsureID()
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	cp := *e
	r.events = append(r.events, &cp)
	return nil
}

func (r *HostEventFake) ListByHost(_ context.Context, hostID string, limit int) ([]*domain.HostEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.HostEvent, 0)
	for _, e := range r.events {
		if e.HostID == hostID {
			cp := *e
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OccurredAt.After(out[j].OccurredAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// ListInWindow mirrors the real repository's half-open [from, to) semantics.
// Deliberately not looser: a fake that accepts what the database rejects lets a
// boundary bug pass every service test and fail only in production.
func (r *HostEventFake) ListInWindow(_ context.Context, hostID string, from, to time.Time, limit int) ([]*domain.HostEvent, error) {
	r.mu.Lock()
	r.ListInWindowCalls++
	err := r.ListInWindowErr
	r.mu.Unlock()
	if err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if !to.After(from) {
		return nil, nil
	}

	out := make([]*domain.HostEvent, 0)
	for _, e := range r.events {
		if e.HostID != hostID {
			continue
		}
		if e.OccurredAt.Before(from) || !e.OccurredAt.Before(to) {
			continue
		}
		cp := *e
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OccurredAt.After(out[j].OccurredAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *HostEventFake) DeleteOlderThan(_ context.Context, cutoff time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	kept := r.events[:0]
	var deleted int64
	for _, e := range r.events {
		if e.OccurredAt.Before(cutoff) {
			deleted++
			continue
		}
		kept = append(kept, e)
	}
	r.events = kept
	return deleted, nil
}

func (r *HostEventFake) DeleteByHost(_ context.Context, hostID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	kept := r.events[:0]
	for _, e := range r.events {
		if e.HostID != hostID {
			kept = append(kept, e)
		}
	}
	r.events = kept
	return nil
}

// Count exposes how many events are held, for tests asserting a bound.
func (r *HostEventFake) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.events)
}
