package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

// ReportHistoryFake is an in-memory generated-report store for tests.
type ReportHistoryFake struct {
	mu   sync.RWMutex
	rows map[string]*domain.ReportHistory // keyed by period (unique)
}

func NewReportHistoryFake() *ReportHistoryFake {
	return &ReportHistoryFake{rows: map[string]*domain.ReportHistory{}}
}

func (f *ReportHistoryFake) Create(_ context.Context, h *domain.ReportHistory) (*domain.ReportHistory, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.rows[h.Period]; exists {
		return nil, repository.ErrDuplicate
	}
	h.EnsureID()
	now := time.Now()
	if h.CreatedAt.IsZero() {
		h.CreatedAt = now
	}
	if h.SentAt.IsZero() {
		h.SentAt = now
	}
	cp := *h
	f.rows[h.Period] = &cp
	out := *f.rows[h.Period]
	return &out, nil
}

func (f *ReportHistoryFake) ListRecent(_ context.Context, limit int) ([]*domain.ReportHistory, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	all := make([]*domain.ReportHistory, 0, len(f.rows))
	for _, r := range f.rows {
		cp := *r
		all = append(all, &cp)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].SentAt.After(all[j].SentAt) })
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (f *ReportHistoryFake) FindByPeriod(_ context.Context, period string) (*domain.ReportHistory, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	r, ok := f.rows[period]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *r
	return &cp, nil
}
