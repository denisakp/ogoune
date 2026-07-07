package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

// AnnouncementFake is an in-memory operator-banner store for tests.
type AnnouncementFake struct {
	mu   sync.RWMutex
	rows map[string]*domain.Announcement
}

func NewAnnouncementFake() *AnnouncementFake {
	return &AnnouncementFake{rows: map[string]*domain.Announcement{}}
}

func (f *AnnouncementFake) Create(_ context.Context, a *domain.Announcement) (*domain.Announcement, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a.EnsureID()
	now := time.Now()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	cp := *a
	f.rows[a.ID] = &cp
	out := *f.rows[a.ID]
	return &out, nil
}

func (f *AnnouncementFake) ListActive(_ context.Context) ([]*domain.Announcement, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]*domain.Announcement, 0, len(f.rows))
	for _, a := range f.rows {
		if a.Active {
			cp := *a
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (f *AnnouncementFake) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.rows[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.rows, id)
	return nil
}
