package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
)

type IncidentEventStepFake struct {
	mu    sync.RWMutex
	steps map[string]*domain.IncidentEventStep
}

// NewIncidentEventStepFake creates a new in-memory fake implementation of IncidentEventStepRepository
func NewIncidentEventStepFake() port.IncidentEventStepRepository {
	return &IncidentEventStepFake{
		steps: make(map[string]*domain.IncidentEventStep),
	}
}

func (f *IncidentEventStepFake) Create(ctx context.Context, s *domain.IncidentEventStep) (*domain.IncidentEventStep, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	s.EnsureID()
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

// FindByIncidentID returns all event steps for an incident sorted by CreatedAt ascending.
func (f *IncidentEventStepFake) FindByIncidentID(incidentID string) []*domain.IncidentEventStep {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]*domain.IncidentEventStep, 0)
	for _, step := range f.steps {
		if step.IncidentID != incidentID {
			continue
		}
		copy := *step
		result = append(result, &copy)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result
}

// CountByIncidentAndStep counts event steps of a given type for an incident.
func (f *IncidentEventStepFake) CountByIncidentAndStep(incidentID string, stepType domain.IncidentEventStepType) int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	count := 0
	for _, step := range f.steps {
		if step.IncidentID == incidentID && step.Step == stepType {
			count++
		}
	}

	return count
}
