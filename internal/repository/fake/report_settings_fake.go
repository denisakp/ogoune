package fake

import (
	"context"
	"sync"
	"time"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

// ReportSettingsFake is an in-memory single-value store for handler/service tests.
type ReportSettingsFake struct {
	mu  sync.RWMutex
	val *domain.ReportSettings
}

func NewReportSettingsFake() *ReportSettingsFake { return &ReportSettingsFake{} }

func (f *ReportSettingsFake) Get(_ context.Context) (*domain.ReportSettings, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.val == nil {
		return nil, repository.ErrNotFound
	}
	cp := *f.val
	return &cp, nil
}

func (f *ReportSettingsFake) Upsert(_ context.Context, s *domain.ReportSettings) (*domain.ReportSettings, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s.ID = domain.ReportSettingsSingletonID
	now := time.Now()
	if f.val == nil {
		s.CreatedAt = now
	} else {
		s.CreatedAt = f.val.CreatedAt
	}
	s.UpdatedAt = now
	cp := *s
	f.val = &cp
	out := *f.val
	return &out, nil
}
