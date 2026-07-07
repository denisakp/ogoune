package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

type EscalationRepository struct {
	mu   sync.RWMutex
	byID map[string]*domain.EscalationPolicy
}

func NewEscalationRepository() *EscalationRepository {
	return &EscalationRepository{byID: make(map[string]*domain.EscalationPolicy)}
}

func (r *EscalationRepository) clone(p *domain.EscalationPolicy) *domain.EscalationPolicy {
	cp := *p
	cp.Steps = append([]domain.EscalationStep(nil), p.Steps...)
	return &cp
}

func (r *EscalationRepository) Create(ctx context.Context, p *domain.EscalationPolicy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.EnsureID()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = p.CreatedAt
	}
	if _, exists := r.byID[p.ID]; exists {
		return ErrDuplicate
	}
	if p.IsActive {
		for _, q := range r.byID {
			if q.IsActive && q.Priority == p.Priority {
				return ErrDuplicate
			}
		}
	}
	r.byID[p.ID] = r.clone(p)
	return nil
}

func (r *EscalationRepository) FindByID(ctx context.Context, id string) (*domain.EscalationPolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return r.clone(p), nil
}

func (r *EscalationRepository) List(ctx context.Context) ([]*domain.EscalationPolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.EscalationPolicy, 0, len(r.byID))
	for _, p := range r.byID {
		out = append(out, r.clone(p))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Priority < out[j].Priority })
	return out, nil
}

func (r *EscalationRepository) Update(ctx context.Context, p *domain.EscalationPolicy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[p.ID]; !ok {
		return repository.ErrNotFound
	}
	p.UpdatedAt = time.Now()
	r.byID[p.ID] = r.clone(p)
	return nil
}

func (r *EscalationRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.byID, id)
	return nil
}

func (r *EscalationRepository) Reorder(ctx context.Context, order []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, id := range order {
		p, ok := r.byID[id]
		if !ok {
			return repository.ErrNotFound
		}
		p.Priority = i + 1
		p.UpdatedAt = time.Now()
	}
	return nil
}

func (r *EscalationRepository) NextPriority(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	max := 0
	for _, p := range r.byID {
		if p.IsActive && p.Priority > max {
			max = p.Priority
		}
	}
	return max + 1, nil
}
