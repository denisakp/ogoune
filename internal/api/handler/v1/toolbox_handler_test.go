package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denisakp/ogoune/internal/service"
)

// stubToolboxService implements ToolboxV1ServiceInterface with canned outputs.
type stubToolboxService struct {
	dnsRes  service.ToolboxDNSResult
	portRes service.ToolboxPortScanResult
	sslRes  service.ToolboxSSLResult
	whoRes  service.ToolboxWhoisResult
	err     error
}

func (s *stubToolboxService) DNS(_ context.Context, _ service.ToolboxDNSQuery) (service.ToolboxDNSResult, error) {
	return s.dnsRes, s.err
}
func (s *stubToolboxService) PortScan(_ context.Context, _ service.ToolboxPortScanQuery) (service.ToolboxPortScanResult, error) {
	return s.portRes, s.err
}
func (s *stubToolboxService) SSL(_ context.Context, _ service.ToolboxSSLQuery) (service.ToolboxSSLResult, error) {
	return s.sslRes, s.err
}
func (s *stubToolboxService) WHOIS(_ context.Context, _ string) (service.ToolboxWhoisResult, error) {
	return s.whoRes, s.err
}

func doPost(h http.HandlerFunc, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestToolboxHandler_DNS_OK(t *testing.T) {
	h := NewToolboxHandler(&stubToolboxService{dnsRes: service.ToolboxDNSResult{
		Records:      []service.ToolboxDNSRecord{{Type: "A", Value: "93.184.216.34"}},
		ResolverUsed: "1.1.1.1",
	}})
	rec := doPost(h.DNS, `{"domain":"example.com","record_types":["A"],"resolver":"cloudflare"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "93.184.216.34") {
		t.Fatalf("missing record in body: %s", rec.Body.String())
	}
}

func TestToolboxHandler_InvalidJSON(t *testing.T) {
	h := NewToolboxHandler(&stubToolboxService{})
	rec := doPost(h.DNS, `{not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", rec.Code)
	}
}

func TestToolboxHandler_ErrorMapping(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		call   func(h *ToolboxHandler) http.HandlerFunc
		body   string
		status int
		code   string
	}{
		{"dns validation", service.ErrToolboxValidation, func(h *ToolboxHandler) http.HandlerFunc { return h.DNS }, `{"domain":""}`, http.StatusBadRequest, "VALIDATION_FAILED"},
		{"port not registered", service.ErrToolboxTargetNotRegistered, func(h *ToolboxHandler) http.HandlerFunc { return h.PortScan }, `{"target":"x","ports":[80]}`, http.StatusForbidden, "TARGET_NOT_REGISTERED"},
		{"port blocked", service.ErrToolboxTargetBlocked, func(h *ToolboxHandler) http.HandlerFunc { return h.PortScan }, `{"target":"x","ports":[80]}`, http.StatusForbidden, "TARGET_BLOCKED"},
		{"ssl cert unavailable", service.ErrToolboxCertUnavailable, func(h *ToolboxHandler) http.HandlerFunc { return h.SSL }, `{"domain":"x"}`, http.StatusUnprocessableEntity, "CERT_UNAVAILABLE"},
		{"whois no data", service.ErrToolboxWhoisNoData, func(h *ToolboxHandler) http.HandlerFunc { return h.WHOIS }, `{"domain":"x"}`, http.StatusUnprocessableEntity, "WHOIS_NO_DATA"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewToolboxHandler(&stubToolboxService{err: tc.err})
			rec := doPost(tc.call(h), tc.body)
			if rec.Code != tc.status {
				t.Fatalf("got %d, want %d; body=%s", rec.Code, tc.status, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tc.code) {
				t.Fatalf("body missing code %q: %s", tc.code, rec.Body.String())
			}
		})
	}
}

func TestToolboxHandler_PortScan_OK(t *testing.T) {
	h := NewToolboxHandler(&stubToolboxService{portRes: service.ToolboxPortScanResult{
		Results:      []service.ToolboxPortResult{{Port: 22, Service: "ssh", Status: "open"}},
		OpenCount:    1,
		ScannedCount: 1,
	}})
	rec := doPost(h.PortScan, `{"target":"db-01","ports":[22],"timeout_ms":500}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"open_count":1`) {
		t.Fatalf("missing open_count: %s", rec.Body.String())
	}
}
