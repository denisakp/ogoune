package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

type IncidentUpdateRepository struct {
	mu   sync.RWMutex
	rows map[string]*domain.IncidentUpdate
}

func NewIncidentUpdateRepository() *IncidentUpdateRepository {
	return &IncidentUpdateRepository{rows: map[string]*domain.IncidentUpdate{}}
}

func (r *IncidentUpdateRepository) Create(_ context.Context, u *domain.IncidentUpdate) (*domain.IncidentUpdate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u.EnsureID()
	now := time.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.PostedAt.IsZero() {
		u.PostedAt = now
	}
	u.UpdatedAt = now
	cp := *u
	r.rows[u.ID] = &cp
	return &cp, nil
}

func (r *IncidentUpdateRepository) FindByID(_ context.Context, id string) (*domain.IncidentUpdate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	row, ok := r.rows[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *row
	return &cp, nil
}

func (r *IncidentUpdateRepository) ListByIncident(_ context.Context, incidentID string) ([]*domain.IncidentUpdate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []*domain.IncidentUpdate{}
	for _, row := range r.rows {
		if row.IncidentID == incidentID {
			cp := *row
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PostedAt.After(out[j].PostedAt) })
	return out, nil
}

func (r *IncidentUpdateRepository) Update(_ context.Context, u *domain.IncidentUpdate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.rows[u.ID]; !ok {
		return repository.ErrNotFound
	}
	u.UpdatedAt = time.Now()
	cp := *u
	r.rows[u.ID] = &cp
	return nil
}

func (r *IncidentUpdateRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rows, id)
	return nil
}
