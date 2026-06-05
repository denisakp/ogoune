package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
)

// Spec 059 FR-023..FR-026a · contracts/escalation-api.md.
var (
	ErrEscalationStepsRange     = errors.New("policy must declare between 1 and 5 steps")
	ErrEscalationChannelsEmpty  = errors.New("each step must declare at least one channel")
	ErrEscalationDelayRange     = errors.New("step delay must be between 1 and 1440 minutes")
	ErrEscalationScopeInvalid   = errors.New("scope.kind must be component or tag")
	ErrEscalationReorderMissing = errors.New("reorder payload missing active policy IDs")
	ErrEscalationReorderUnknown = errors.New("reorder payload contains unknown or inactive policy IDs")
)

const (
	escalationMinSteps   = 1
	escalationMaxSteps   = 5
	escalationMinDelay   = 1
	escalationMaxDelayMn = 1440
)

type EscalationService struct {
	repo         port.EscalationRepository
	resourceRepo port.ResourceRepository
}

func NewEscalationService(repo port.EscalationRepository, resourceRepo port.ResourceRepository) *EscalationService {
	return &EscalationService{repo: repo, resourceRepo: resourceRepo}
}

func validatePolicy(p *domain.EscalationPolicy) error {
	if p.Scope.Kind != domain.EscalationScopeComponent && p.Scope.Kind != domain.EscalationScopeTag {
		return ErrEscalationScopeInvalid
	}
	if len(p.Steps) < escalationMinSteps || len(p.Steps) > escalationMaxSteps {
		return ErrEscalationStepsRange
	}
	for _, s := range p.Steps {
		if s.DelayMinutes < escalationMinDelay || s.DelayMinutes > escalationMaxDelayMn {
			return ErrEscalationDelayRange
		}
		if len(s.ChannelIDs) == 0 {
			return ErrEscalationChannelsEmpty
		}
	}
	return nil
}

func (s *EscalationService) Create(ctx context.Context, p *domain.EscalationPolicy) (*domain.EscalationPolicy, error) {
	if err := validatePolicy(p); err != nil {
		return nil, err
	}
	// Assign next-priority among existing.
	next, err := s.repo.NextPriority(ctx)
	if err != nil {
		return nil, fmt.Errorf("next priority: %w", err)
	}
	p.Priority = next
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, p.ID)
}

func (s *EscalationService) List(ctx context.Context) ([]*domain.EscalationPolicy, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Priority < rows[j].Priority })
	return rows, nil
}

func (s *EscalationService) Get(ctx context.Context, id string) (*domain.EscalationPolicy, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *EscalationService) Update(ctx context.Context, p *domain.EscalationPolicy) (*domain.EscalationPolicy, error) {
	if err := validatePolicy(p); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, p.ID)
}

func (s *EscalationService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Reorder validates that `order` covers exactly the active policy IDs (no
// missing, no unknown, no inactive) then delegates the atomic update.
func (s *EscalationService) Reorder(ctx context.Context, order []string) error {
	all, err := s.repo.List(ctx)
	if err != nil {
		return err
	}
	active := make(map[string]struct{})
	for _, p := range all {
		if p.IsActive {
			active[p.ID] = struct{}{}
		}
	}
	if len(order) != len(active) {
		return ErrEscalationReorderMissing
	}
	seen := make(map[string]struct{}, len(order))
	for _, id := range order {
		if _, ok := active[id]; !ok {
			return ErrEscalationReorderUnknown
		}
		if _, dup := seen[id]; dup {
			return ErrEscalationReorderUnknown
		}
		seen[id] = struct{}{}
	}
	return s.repo.Reorder(ctx, order)
}

// MatchForResource returns the lowest-priority active policy whose scope
// matches the resource (by component or by any tag). Returns (nil, nil)
// when no policy matches.
func (s *EscalationService) MatchForResource(ctx context.Context, resourceID string) (*domain.EscalationPolicy, error) {
	if s.resourceRepo == nil {
		return nil, errors.New("resource repository required for matching")
	}
	res, err := s.resourceRepo.FindByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	policies, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	// Already sorted ascending by priority.
	tagSet := make(map[string]struct{}, len(res.Tags))
	for _, t := range res.Tags {
		if t == nil {
			continue
		}
		tagSet[t.ID] = struct{}{}
		tagSet[t.Name] = struct{}{}
	}
	for _, p := range policies {
		if !p.IsActive {
			continue
		}
		switch p.Scope.Kind {
		case domain.EscalationScopeComponent:
			if res.ComponentID != nil && *res.ComponentID == p.Scope.Value {
				return p, nil
			}
		case domain.EscalationScopeTag:
			if _, ok := tagSet[p.Scope.Value]; ok {
				return p, nil
			}
		}
	}
	return nil, nil
}
