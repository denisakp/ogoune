package integrations

import (
	"sort"
	"strings"

	domain "github.com/denisakp/ogoune/internal/domain"
)

// NOTE (D1): this 5-panel dashboard shape is duplicated with the frontend static
// fallback (web/src/views/metrics/integrations.ts GRAFANA_DASHBOARD). If the
// panels/queries change here, update the frontend fallback too, and vice versa.

const dsPrometheus = "${DS_PROMETHEUS}"

func distinctSorted(vals []string) []string {
	seen := make(map[string]bool, len(vals))
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

// customVar builds a Grafana "custom" template variable with options seeded from
// real config values (so the picker lists the operator's own resources).
func customVar(name string, values []string) map[string]any {
	options := []any{map[string]any{"text": "All", "value": "$__all", "selected": true}}
	for _, v := range values {
		options = append(options, map[string]any{"text": v, "value": v, "selected": false})
	}
	return map[string]any{
		"name":       name,
		"label":      name,
		"type":       "custom",
		"query":      strings.Join(values, ","),
		"options":    options,
		"current":    map[string]any{"text": "All", "value": "$__all"},
		"includeAll": true,
		"multi":      true,
	}
}

func panel(id int, title, ptype string, x, y, w, h int, expr, legend, unit string) map[string]any {
	target := map[string]any{"expr": expr, "refId": "A"}
	if legend != "" {
		target["legendFormat"] = legend
	}
	p := map[string]any{
		"id":         id,
		"title":      title,
		"type":       ptype,
		"datasource": dsPrometheus,
		"gridPos":    map[string]any{"x": x, "y": y, "w": w, "h": h},
		"targets":    []any{target},
	}
	if unit != "" {
		p["fieldConfig"] = map[string]any{"defaults": map[string]any{"unit": unit}}
	}
	return p
}

// BuildDashboard renders a config-derived Grafana dashboard model (JSON-serializable),
// with $resource/$component/$type variables seeded from real values. Deterministic
// (distinct sorted values; encoding/json sorts map keys).
func BuildDashboard(resources []*domain.Resource, components []*domain.Component) any {
	names := make([]string, 0, len(resources))
	types := make([]string, 0, len(resources))
	for _, r := range resources {
		names = append(names, r.Name)
		types = append(types, string(r.Type))
	}
	comps := make([]string, 0, len(components))
	for _, c := range components {
		comps = append(comps, c.Name)
	}

	return map[string]any{
		"__inputs": []any{map[string]any{
			"name":        "DS_PROMETHEUS",
			"label":       "Prometheus",
			"description": "Prometheus data source scraping the Ogoune /metrics endpoint",
			"type":        "datasource",
			"pluginId":    "prometheus",
			"pluginName":  "Prometheus",
		}},
		"__requires": []any{
			map[string]any{"type": "grafana", "id": "grafana", "name": "Grafana", "version": "10.0.0"},
			map[string]any{"type": "datasource", "id": "prometheus", "name": "Prometheus", "version": "1.0.0"},
		},
		"title":         "Ogoune — Uptime & Monitoring",
		"uid":           "ogoune-overview",
		"tags":          []any{"ogoune", "uptime"},
		"schemaVersion": 39,
		"version":       1,
		"time":          map[string]any{"from": "now-24h", "to": "now"},
		"refresh":       "30s",
		"templating": map[string]any{"list": []any{
			customVar("resource", distinctSorted(names)),
			customVar("component", distinctSorted(comps)),
			customVar("type", distinctSorted(types)),
		}},
		"panels": []any{
			panel(1, "Resources up", "stat", 0, 0, 6, 4, "sum(ogoune_resource_up)", "", ""),
			panel(2, "Active incidents", "stat", 6, 0, 6, 4, "sum(ogoune_incidents_active)", "", ""),
			panel(3, "Uptime % (24h) by resource", "timeseries", 12, 0, 12, 8,
				`ogoune_uptime_ratio{window="24h", name=~"$resource"}`, "{{name}}", "percent"),
			panel(4, "Check success rate", "timeseries", 0, 8, 12, 8,
				`sum(rate(ogoune_checks_total{status="success"}[5m])) / clamp_min(sum(rate(ogoune_checks_total[5m])), 1)`, "", "percentunit"),
			panel(5, "Check duration p95 (s)", "timeseries", 12, 8, 12, 8,
				`histogram_quantile(0.95, sum(rate(ogoune_check_duration_seconds_bucket[5m])) by (le, name))`, "{{name}}", "s"),
		},
	}
}
