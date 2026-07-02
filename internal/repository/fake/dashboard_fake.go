package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

// DashboardRepository — in-memory port.DashboardRepository for tests.
// Does NOT simulate the users JOIN: OwnerName is whatever is stored on the
// struct (empty for service-created rows). owner_name correctness is verified
// by the dual-dialect contract test, not here.
type DashboardRepository struct {
	mu   sync.RWMutex
	byID map[string]*domain.Dashboard
}

func NewDashboardRepository() *DashboardRepository {
	return &DashboardRepository{byID: make(map[string]*domain.Dashboard)}
}

func (r *DashboardRepository) Create(ctx context.Context, d *domain.Dashboard) (*domain.Dashboard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d.EnsureID()
	now := time.Now()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = now
	}
	if _, exists := r.byID[d.ID]; exists {
		return nil, ErrDuplicate
	}
	cp := *d
	r.byID[d.ID] = &cp
	out := cp
	return &out, nil
}

func (r *DashboardRepository) FindByID(ctx context.Context, id string) (*domain.Dashboard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.byID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *d
	return &cp, nil
}

func (r *DashboardRepository) List(ctx context.Context, limit, offset int) ([]*domain.Dashboard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]*domain.Dashboard, 0, len(r.byID))
	for _, d := range r.byID {
		cp := *d
		all = append(all, &cp)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].UpdatedAt.After(all[j].UpdatedAt) })
	if offset >= len(all) {
		return []*domain.Dashboard{}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (r *DashboardRepository) Update(ctx context.Context, d *domain.Dashboard) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[d.ID]; !ok {
		return repository.ErrNotFound
	}
	d.UpdatedAt = time.Now()
	cp := *d
	r.byID[d.ID] = &cp
	return nil
}

func (r *DashboardRepository) UpdateWidgets(ctx context.Context, id string, widgets []domain.WidgetInstance, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.byID[id]
	if !ok {
		return repository.ErrNotFound
	}
	d.Widgets = append([]domain.WidgetInstance(nil), widgets...)
	d.UpdatedAt = at
	return nil
}

func (r *DashboardRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.byID, id)
	return nil
}
