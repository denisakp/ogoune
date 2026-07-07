package integrations

import (
	"fmt"
	"sort"

	domain "github.com/denisakp/ogoune/internal/domain"
	"gopkg.in/yaml.v3"
)

// Prometheus rules file shape (deterministic key order via struct fields;
// yaml.v3 sorts map keys, so labels/annotations are stable too).
type promRuleFile struct {
	Groups []promRuleGroup `yaml:"groups"`
}

type promRuleGroup struct {
	Name  string     `yaml:"name"`
	Rules []promRule `yaml:"rules"`
}

type promRule struct {
	Alert       string            `yaml:"alert"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// derivedForSeconds is the latency before Ogoune declares a resource down:
// max(60, interval × confirmation). Zero/absent values floor to 60s.
func derivedForSeconds(r *domain.Resource) int {
	n := r.ConfirmationChecks
	if n <= 0 {
		n = 1
	}
	s := r.Interval * n
	if s < 60 {
		s = 60
	}
	return s
}

// representativeForSeconds is the median of derivedForSeconds across resources
// (0 when there are none). Even count → average of the two central values.
func representativeForSeconds(resources []*domain.Resource) int {
	if len(resources) == 0 {
		return 0
	}
	vals := make([]int, 0, len(resources))
	for _, r := range resources {
		vals = append(vals, derivedForSeconds(r))
	}
	sort.Ints(vals)
	n := len(vals)
	if n%2 == 1 {
		return vals[n/2]
	}
	return (vals[n/2-1] + vals[n/2]) / 2
}

// formatPromDuration renders seconds as a compact Prometheus duration.
func formatPromDuration(seconds int) string {
	switch {
	case seconds%3600 == 0:
		return fmt.Sprintf("%dh", seconds/3600)
	case seconds%60 == 0:
		return fmt.Sprintf("%dm", seconds/60)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// clampThreshold defaults 0/unset to 99 and bounds to [1,100].
func clampThreshold(t int) int {
	if t <= 0 {
		return 99
	}
	if t > 100 {
		return 100
	}
	return t
}

// BuildAlertRules renders config-derived Prometheus alerting rules as YAML.
// Deterministic: resources sorted by name; yaml.v3 sorts map keys.
func BuildAlertRules(resources []*domain.Resource, uptimeThreshold int) (string, error) {
	threshold := clampThreshold(uptimeThreshold)

	rs := append([]*domain.Resource(nil), resources...)
	sort.Slice(rs, func(i, j int) bool { return rs[i].Name < rs[j].Name })

	down := promRule{
		Alert:  "OgouneResourceDown",
		Expr:   "ogoune_resource_up == 0",
		Labels: map[string]string{"severity": "critical"},
		Annotations: map[string]string{
			"summary":     "{{ $labels.name }} is down",
			"description": "Ogoune declares a resource down after interval × confirmation per monitor; the for: below is the representative (median) window across monitors.",
		},
	}
	if rep := representativeForSeconds(rs); rep > 0 {
		down.For = formatPromDuration(rep)
	}

	rules := []promRule{
		down,
		{
			Alert:       "OgouneLowUptime24h",
			Expr:        fmt.Sprintf(`ogoune_uptime_ratio{window="24h"} < %d`, threshold),
			For:         "10m",
			Labels:      map[string]string{"severity": "warning"},
			Annotations: map[string]string{"summary": fmt.Sprintf("{{ $labels.name }} uptime below %d%% (24h)", threshold)},
		},
		{
			Alert:       "OgouneActiveIncident",
			Expr:        "ogoune_incidents_active > 0",
			For:         "1m",
			Labels:      map[string]string{"severity": "warning"},
			Annotations: map[string]string{"summary": "Open incident on {{ $labels.name }}"},
		},
		{
			Alert:       "OgouneHighFailureRate",
			Expr:        `sum by (name) (rate(ogoune_checks_total{status="failure"}[5m])) > 0`,
			For:         "10m",
			Labels:      map[string]string{"severity": "warning"},
			Annotations: map[string]string{"summary": "{{ $labels.name }} is failing checks"},
		},
	}

	out, err := yaml.Marshal(promRuleFile{Groups: []promRuleGroup{{Name: "ogoune", Rules: rules}}})
	if err != nil {
		return "", fmt.Errorf("marshal alert rules: %w", err)
	}
	header := "# Ogoune — Prometheus alerting rules, generated from your configuration.\n" +
		"# Evaluated by Prometheus; firing alerts are routed by Alertmanager.\n"
	return header + string(out), nil
}
