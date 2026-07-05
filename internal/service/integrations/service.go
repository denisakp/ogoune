// Package integrations builds config-derived observability artifacts (Prometheus
// alerting rules + Grafana dashboard) from Ogoune's resources/components.
// Pure builders (alertrules.go, dashboard.go) behind a thin read-only service.
package integrations

import (
	"context"

	"github.com/denisakp/ogoune/internal/port"
)

// IntegrationsService reads config and delegates to the pure builders.
type IntegrationsService struct {
	resources  port.ResourceRepository
	components port.ComponentRepository
}

func NewIntegrationsService(resources port.ResourceRepository, components port.ComponentRepository) *IntegrationsService {
	return &IntegrationsService{resources: resources, components: components}
}

// AlertRulesYAML returns config-derived Prometheus alerting rules as YAML.
func (s *IntegrationsService) AlertRulesYAML(ctx context.Context, uptimeThreshold int) (string, error) {
	res, err := s.resources.List(ctx, 10000, 0)
	if err != nil {
		return "", err
	}
	return BuildAlertRules(res, uptimeThreshold)
}

// GrafanaDashboard returns a config-derived Grafana dashboard model (JSON-serializable).
func (s *IntegrationsService) GrafanaDashboard(ctx context.Context) (any, error) {
	res, err := s.resources.List(ctx, 10000, 0)
	if err != nil {
		return nil, err
	}
	comps, err := s.components.List(ctx, 10000, 0)
	if err != nil {
		return nil, err
	}
	return BuildDashboard(res, comps), nil
}
