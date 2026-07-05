package resourceimport

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
)

// ErrManifestTooLarge is returned when a manifest exceeds MaxManifestResources.
var ErrManifestTooLarge = fmt.Errorf("manifest exceeds the maximum of %d resources", MaxManifestResources)

// ErrValidationFailed is returned by Import when one or more rows are invalid
// (all-or-nothing: nothing is written). The report carries the per-row detail.
var ErrValidationFailed = errors.New("manifest validation failed")

// ParseError wraps a manifest parse failure so the handler can render a 422.
type ParseError struct{ err error }

func (e *ParseError) Error() string { return e.err.Error() }
func (e *ParseError) Unwrap() error { return e.err }

// resourceGateway is the subset of ResourceService the importer needs.
type resourceGateway interface {
	CreateResource(ctx context.Context, payload *dto.CreateResourcePayload) (*domain.Resource, error)
	DeleteResource(ctx context.Context, resourceID string) error
	ListAll(ctx context.Context) ([]*domain.Resource, error)
}

// componentGateway resolves/creates components by name (non-secret; auto-created).
type componentGateway interface {
	List(ctx context.Context, limit, offset int) ([]*domain.Component, error)
	Create(ctx context.Context, c *domain.Component) (*domain.Component, error)
}

// channelGateway lists notification channels for name-existence checks.
type channelGateway interface {
	List(ctx context.Context, limit, offset int) ([]*domain.NotificationChannel, error)
}

// Service orchestrates manifest validation, import, and export.
type Service struct {
	resources  resourceGateway
	components componentGateway
	channels   channelGateway
}

// NewService builds the importer/exporter service.
func NewService(resources resourceGateway, components componentGateway, channels channelGateway) *Service {
	return &Service{resources: resources, components: components, channels: channels}
}

// DryRun parses + validates a manifest and returns a report without any write.
func (s *Service) DryRun(ctx context.Context, raw []byte, opts dtoV1.ImportOptions) (*dtoV1.ImportReport, error) {
	m, err := s.parseAndCap(raw)
	if err != nil {
		return nil, err
	}
	existing, channels, err := s.lookups(ctx)
	if err != nil {
		return nil, err
	}
	rows := Validate(m, existing, channels, normalizePolicy(opts.DuplicatePolicy))
	return buildReport(true, rows), nil
}

// Import validates then creates all valid rows via ResourceService, all-or-nothing.
// If any row is invalid (or a duplicate under the "error" policy), nothing is
// written and ErrValidationFailed is returned with the per-row report.
func (s *Service) Import(ctx context.Context, raw []byte, opts dtoV1.ImportOptions) (*dtoV1.ImportReport, error) {
	policy := normalizePolicy(opts.DuplicatePolicy)
	m, err := s.parseAndCap(raw)
	if err != nil {
		return nil, err
	}
	existing, channels, err := s.lookups(ctx)
	if err != nil {
		return nil, err
	}
	rows := Validate(m, existing, channels, policy)

	if opts.DryRun {
		return buildReport(true, rows), nil
	}

	// All-or-nothing: any errored row blocks the entire import.
	for _, row := range rows {
		if row.Action == dtoV1.RowActionError {
			report := buildReport(false, rows)
			return report, ErrValidationFailed
		}
	}

	// Resolve/auto-create components referenced by the create rows.
	componentIDs, err := s.resolveComponents(ctx, m, rows)
	if err != nil {
		return nil, err
	}

	// Create each create-row; compensate (delete) on any failure to preserve
	// all-or-nothing observable semantics (research D6).
	created := make([]string, 0, len(rows))
	for i, row := range rows {
		if row.Action != dtoV1.RowActionCreate {
			continue
		}
		payload := toPayload(mergeDefaults(m.Resources[i], m.Defaults), componentIDs)
		res, cerr := s.resources.CreateResource(ctx, payload)
		if cerr != nil {
			s.compensate(ctx, created)
			return nil, fmt.Errorf("failed to create resource %q (row %d): %w", row.Name, i, cerr)
		}
		created = append(created, res.ID)
	}

	return buildReport(false, rows), nil
}

// compensate rolls back already-created resources after a mid-loop failure.
func (s *Service) compensate(ctx context.Context, ids []string) {
	for _, id := range ids {
		_ = s.resources.DeleteResource(ctx, id)
	}
}

// resolveComponents ensures every component name referenced by a create row
// exists, creating missing ones, and returns a name→ID map.
func (s *Service) resolveComponents(ctx context.Context, m *Manifest, rows []dtoV1.RowResult) (map[string]string, error) {
	existing, err := s.components.List(ctx, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list components: %w", err)
	}
	byName := make(map[string]string, len(existing))
	for _, c := range existing {
		if c != nil {
			byName[c.Name] = c.ID
		}
	}
	for i, row := range rows {
		if row.Action != dtoV1.RowActionCreate {
			continue
		}
		name := strings.TrimSpace(m.Resources[i].Component)
		if name == "" || byName[name] != "" {
			continue
		}
		created, cerr := s.components.Create(ctx, &domain.Component{Name: name})
		if cerr != nil {
			return nil, fmt.Errorf("failed to create component %q: %w", name, cerr)
		}
		byName[name] = created.ID
	}
	return byName, nil
}

// parseAndCap parses the manifest and enforces the size cap.
func (s *Service) parseAndCap(raw []byte) (*Manifest, error) {
	m, err := Parse(raw)
	if err != nil {
		return nil, &ParseError{err: err}
	}
	if len(m.Resources) > MaxManifestResources {
		return nil, ErrManifestTooLarge
	}
	return m, nil
}

// lookups gathers existing resource names + channel names for validation.
func (s *Service) lookups(ctx context.Context) (map[string]bool, map[string]bool, error) {
	all, err := s.resources.ListAll(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list resources: %w", err)
	}
	existing := make(map[string]bool, len(all))
	for _, r := range all {
		if r != nil {
			existing[r.Name] = true
		}
	}
	channelsList, err := s.channels.List(ctx, 10000, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list notification channels: %w", err)
	}
	channels := make(map[string]bool, len(channelsList))
	for _, c := range channelsList {
		if c != nil {
			channels[c.Name] = true
		}
	}
	return existing, channels, nil
}

// toPayload maps a defaults-merged declaration to a CreateResourcePayload.
func toPayload(decl ResourceDecl, componentIDs map[string]string) *dto.CreateResourcePayload {
	p := &dto.CreateResourcePayload{
		Name:                     decl.Name,
		Type:                     domain.ResourceType(decl.Type),
		Target:                   decl.Target,
		Interval:                 derefInt(decl.Interval),
		Timeout:                  derefInt(decl.Timeout),
		Tags:                     decl.Tags,
		NotificationChannelNames: decl.NotificationChannels,
		Keyword:                  decl.Keyword,
		KeywordMode:              decl.KeywordMode,
		ProtocolType:             decl.ProtocolType,
		ProtocolPort:             decl.ProtocolPort,
		HeartbeatInterval:        decl.HeartbeatInterval,
		HeartbeatGrace:           decl.HeartbeatGrace,
		ConfirmationChecks:       decl.ConfirmationChecks,
		ConfirmationInterval:     decl.ConfirmationInterval,
	}
	if name := strings.TrimSpace(decl.Component); name != "" {
		if id := componentIDs[name]; id != "" {
			cid := id
			p.ComponentID = &cid
		}
	}
	return p
}

func derefInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func normalizePolicy(p dtoV1.DuplicatePolicy) dtoV1.DuplicatePolicy {
	if p == dtoV1.DuplicatePolicyError {
		return dtoV1.DuplicatePolicyError
	}
	return dtoV1.DuplicatePolicySkip
}

// buildReport tallies row actions into an ImportReport.
func buildReport(dryRun bool, rows []dtoV1.RowResult) *dtoV1.ImportReport {
	report := &dtoV1.ImportReport{DryRun: dryRun, Total: len(rows), Rows: rows}
	for _, row := range rows {
		switch row.Action {
		case dtoV1.RowActionCreate:
			if !dryRun {
				report.Created++
			}
		case dtoV1.RowActionSkip:
			report.Skipped++
		case dtoV1.RowActionError:
			report.Failed++
		}
	}
	return report
}
