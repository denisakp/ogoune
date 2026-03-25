package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	domain "github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
)

type IncidentEventStepFake struct {
	mu    sync.RWMutex
	steps map[string]*domain.IncidentEventStep
}

// NewIncidentEventStepFake creates a new in-memory fake implementation of IncidentEventStepRepository
func NewIncidentEventStepFake() repository.IncidentEventStepRepository {
	return &IncidentEventStepFake{
		steps: make(map[string]*domain.IncidentEventStep),
	}
}

func (f *IncidentEventStepFake) Create(ctx context.Context, s *domain.IncidentEventStep) (*domain.IncidentEventStep, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Call BeforeCreate hook like GORM does - generates ID if not set
	if err := s.BeforeCreate(nil); err != nil {
		return nil, ErrInvalidInput
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}

	if _, exists := f.steps[s.ID]; exists {
		return nil, ErrDuplicate
	}

	// Create a copy to avoid external mutations
	step := *s
	f.steps[s.ID] = &step
	return &step, nil
}

func (f *IncidentEventStepFake) FindByID(ctx context.Context, id string) (*domain.IncidentEventStep, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	step, exists := f.steps[id]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy to avoid external mutations
	result := *step
	return &result, nil
}

func (f *IncidentEventStepFake) FindLastByIncidentAndStep(ctx context.Context, incidentID string, stepType domain.IncidentEventStepType) (*domain.IncidentEventStep, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	matches := make([]*domain.IncidentEventStep, 0)
	for _, step := range f.steps {
		if step.IncidentID == incidentID && step.Step == stepType {
			matches = append(matches, step)
		}
	}
	if len(matches) == 0 {
		return nil, ErrNotFound
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].CreatedAt.After(matches[j].CreatedAt)
	})

	result := *matches[0]
	return &result, nil
}

func (f *IncidentEventStepFake) List(ctx context.Context, limit, offset int) ([]*domain.IncidentEventStep, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var steps []*domain.IncidentEventStep
	i := 0
	for _, step := range f.steps {
		if i < offset {
			i++
			continue
		}
		if len(steps) >= limit {
			break
		}
		// Return a copy to avoid external mutations
		result := *step
		steps = append(steps, &result)
		i++
	}

	return steps, nil
}

func (f *IncidentEventStepFake) Update(ctx context.Context, s *domain.IncidentEventStep) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.steps[s.ID]; !exists {
		return ErrNotFound
	}

	// Create a copy to avoid external mutations
	step := *s
	f.steps[s.ID] = &step
	return nil
}

func (f *IncidentEventStepFake) Delete(ctx context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.steps[id]; !exists {
		return ErrNotFound
	}

	delete(f.steps, id)
	return nil
}
