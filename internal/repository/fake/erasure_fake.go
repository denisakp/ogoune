package fake

import (
	"context"
	"sort"
	"sync"

	"github.com/denisakp/ogoune/internal/domain"
)

// ErasureFake records the plans it is asked to apply (spec 095). It does not
// replay them onto other fakes: service tests assert on the plan itself.
type ErasureFake struct {
	mu      sync.Mutex
	Applied []domain.ErasurePlan
	records []domain.ErasureRecord
	// ApplyErr, when set, fails Apply and records nothing.
	ApplyErr error
}

func NewErasureFake() *ErasureFake { return &ErasureFake{} }

func (f *ErasureFake) Apply(_ context.Context, plan domain.ErasurePlan) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ApplyErr != nil {
		return f.ApplyErr
	}
	plan.Record.EnsureID()
	f.Applied = append(f.Applied, plan)
	f.records = append(f.records, plan.Record)
	return nil
}

func (f *ErasureFake) FindRecordsByFingerprint(_ context.Context, fp string) ([]domain.ErasureRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []domain.ErasureRecord{}
	for _, r := range f.records {
		if r.SubjectFingerprint == fp {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
