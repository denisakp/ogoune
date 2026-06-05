package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	domain "github.com/denisakp/ogoune/internal/domain"
)

type aggKey struct {
	resourceID string
	day        string
}

type UptimeDailyAggRepository struct {
	mu   sync.RWMutex
	rows map[aggKey]*domain.UptimeDailyAgg
}

func NewUptimeDailyAggRepository() *UptimeDailyAggRepository {
	return &UptimeDailyAggRepository{rows: map[aggKey]*domain.UptimeDailyAgg{}}
}

func truncDayUTC(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func (r *UptimeDailyAggRepository) Upsert(_ context.Context, a *domain.UptimeDailyAgg) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	day := truncDayUTC(a.Day)
	cp := *a
	cp.Day = day
	r.rows[aggKey{a.ResourceID, day.Format("2006-01-02")}] = &cp
	return nil
}

func (r *UptimeDailyAggRepository) FindRange(_ context.Context, resourceIDs []string, from, to time.Time) ([]*domain.UptimeDailyAgg, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fromD := truncDayUTC(from)
	toD := truncDayUTC(to)
	idSet := map[string]struct{}{}
	for _, id := range resourceIDs {
		idSet[id] = struct{}{}
	}
	out := []*domain.UptimeDailyAgg{}
	for _, row := range r.rows {
		if _, ok := idSet[row.ResourceID]; !ok {
			continue
		}
		if row.Day.Before(fromD) || row.Day.After(toD) {
			continue
		}
		cp := *row
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Day.Before(out[j].Day) })
	return out, nil
}

func (r *UptimeDailyAggRepository) FindForResource(ctx context.Context, resourceID string, from, to time.Time) ([]*domain.UptimeDailyAgg, error) {
	return r.FindRange(ctx, []string{resourceID}, from, to)
}
