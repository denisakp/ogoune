package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubIntegrationsService struct {
	yaml      string
	dashboard any
	lastThr   int
}

func (s *stubIntegrationsService) AlertRulesYAML(_ context.Context, uptimeThreshold int) (string, error) {
	s.lastThr = uptimeThreshold
	return s.yaml, nil
}
func (s *stubIntegrationsService) GrafanaDashboard(context.Context) (any, error) {
	return s.dashboard, nil
}

func TestIntegrationsHandler_AlertRules_YAML(t *testing.T) {
	stub := &stubIntegrationsService{yaml: "groups:\n- name: ogoune\n"}
	h := NewIntegrationsHandler(stub)
	rec := httptest.NewRecorder()
	h.AlertRules(rec, httptest.NewRequest(http.MethodGet, "/integrations/alert-rules?uptimeThreshold=95", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/yaml") {
		t.Fatalf("content-type=%q, want text/yaml", ct)
	}
	if !strings.Contains(rec.Body.String(), "name: ogoune") {
		t.Fatalf("body=%s", rec.Body.String())
	}
	if stub.lastThr != 95 {
		t.Fatalf("threshold not forwarded: %d", stub.lastThr)
	}
}

func TestIntegrationsHandler_Dashboard_JSON(t *testing.T) {
	h := NewIntegrationsHandler(&stubIntegrationsService{dashboard: map[string]any{"title": "Ogoune"}})
	rec := httptest.NewRecorder()
	h.GrafanaDashboard(rec, httptest.NewRequest(http.MethodGet, "/integrations/grafana-dashboard", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content-type=%q", ct)
	}
	if !strings.Contains(rec.Body.String(), `"title":"Ogoune"`) {
		t.Fatalf("body=%s", rec.Body.String())
	}
}
