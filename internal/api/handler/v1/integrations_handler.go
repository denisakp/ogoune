package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

// IntegrationsV1ServiceInterface is the slice of *IntegrationsService used here.
type IntegrationsV1ServiceInterface interface {
	AlertRulesYAML(ctx context.Context, uptimeThreshold int) (string, error)
	GrafanaDashboard(ctx context.Context) (any, error)
}

// IntegrationsHandler exposes /api/v1/integrations (config-derived observability assets).
type IntegrationsHandler struct {
	service IntegrationsV1ServiceInterface
}

func NewIntegrationsHandler(svc IntegrationsV1ServiceInterface) *IntegrationsHandler {
	return &IntegrationsHandler{service: svc}
}

// AlertRules handles GET /api/v1/integrations/alert-rules.
//
// @Summary  Download config-derived Prometheus alerting rules (YAML)
// @Tags     integrations
// @Security BearerAuth
// @Produce  text/yaml
// @Param    uptimeThreshold query int false "Uptime alert threshold percent (default 99)"
// @Success  200 {string} string "Prometheus rules YAML"
// @Router   /integrations/alert-rules [get]
func (h *IntegrationsHandler) AlertRules(w http.ResponseWriter, r *http.Request) {
	threshold := 99
	if v := r.URL.Query().Get("uptimeThreshold"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			threshold = n
		}
	}
	body, err := h.service.AlertRulesYAML(r.Context(), threshold)
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to build alert rules")
		return
	}
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="ogoune-alerts.rules.yml"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

// GrafanaDashboard handles GET /api/v1/integrations/grafana-dashboard.
//
// @Summary  Download config-derived Grafana dashboard (JSON)
// @Tags     integrations
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} map[string]interface{} "Grafana dashboard model"
// @Router   /integrations/grafana-dashboard [get]
func (h *IntegrationsHandler) GrafanaDashboard(w http.ResponseWriter, r *http.Request) {
	model, err := h.service.GrafanaDashboard(r.Context())
	if err != nil {
		respondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to build dashboard")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="ogoune-grafana-dashboard.json"`)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model)
}
