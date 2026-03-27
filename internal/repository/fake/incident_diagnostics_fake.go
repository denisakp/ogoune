package fake

import (
	"context"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository"
)

// IncidentDiagnosticsFake is a fake implementation of IncidentDiagnosticsRepository for testing
type IncidentDiagnosticsFake struct {
	diagnostics map[string]*domain.IncidentDiagnostics
	idCounter   int
}

// NewIncidentDiagnosticsFake creates a new fake diagnostics repository
func NewIncidentDiagnosticsFake() repository.IncidentDiagnosticsRepository {
	return &IncidentDiagnosticsFake{
		diagnostics: make(map[string]*domain.IncidentDiagnostics),
		idCounter:   0,
	}
}

// Create stores a new incident diagnostics record
func (f *IncidentDiagnosticsFake) Create(ctx context.Context, d *domain.IncidentDiagnostics) (*domain.IncidentDiagnostics, error) {
	if d == nil {
		return nil, repository.ErrInvalidInput
	}

	if d.ID == "" {
		f.idCounter++
		d.ID = "diag-" + string(rune(f.idCounter))
	}

	f.diagnostics[d.ID] = d
	return d, nil
}

// FindByIncidentID retrieves diagnostics for an incident
func (f *IncidentDiagnosticsFake) FindByIncidentID(ctx context.Context, incidentID string) (*domain.IncidentDiagnostics, error) {
	if incidentID == "" {
		return nil, repository.ErrInvalidInput
	}

	for _, d := range f.diagnostics {
		if d.IncidentID == incidentID {
			return d, nil
		}
	}

	return nil, repository.ErrNotFound
}

// Update modifies an existing diagnostics record
func (f *IncidentDiagnosticsFake) Update(ctx context.Context, d *domain.IncidentDiagnostics) error {
	if d == nil || d.ID == "" {
		return repository.ErrInvalidInput
	}

	if _, exists := f.diagnostics[d.ID]; !exists {
		return repository.ErrNotFound
	}

	f.diagnostics[d.ID] = d
	return nil
}

// Delete removes a diagnostics record
func (f *IncidentDiagnosticsFake) Delete(ctx context.Context, id string) error {
	if id == "" {
		return repository.ErrInvalidInput
	}

	if _, exists := f.diagnostics[id]; !exists {
		return repository.ErrNotFound
	}

	delete(f.diagnostics, id)
	return nil
}
