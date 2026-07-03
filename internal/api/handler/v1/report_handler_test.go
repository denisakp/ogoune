package v1

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/service"
)

type stubReportService struct {
	settings   *domain.ReportSettings
	saveErr    error
	history    []*domain.ReportHistory
	preview    *domain.ReportHistory
	lastSaved  *domain.ReportSettings
}

func (s *stubReportService) GetSettings(context.Context) (*domain.ReportSettings, error) {
	if s.settings != nil {
		return s.settings, nil
	}
	return &domain.ReportSettings{Schedule: domain.ReportScheduleMonthly1st, Scope: domain.ReportScopeAllResources}, nil
}
func (s *stubReportService) SaveSettings(_ context.Context, in *domain.ReportSettings) (*domain.ReportSettings, error) {
	if s.saveErr != nil {
		return nil, s.saveErr
	}
	s.lastSaved = in
	return in, nil
}
func (s *stubReportService) ListHistory(context.Context, int) ([]*domain.ReportHistory, error) {
	return s.history, nil
}
func (s *stubReportService) GeneratePreview(context.Context) (*domain.ReportHistory, error) {
	return s.preview, nil
}

func TestReportHandler_GetSettings_DefaultEnvelope(t *testing.T) {
	h := NewReportHandler(&stubReportService{})
	rec := httptest.NewRecorder()
	h.GetSettings(rec, httptest.NewRequest(http.MethodGet, "/reports/settings", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"schedule":"monthly-1st"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestReportHandler_UpdateSettings_OK(t *testing.T) {
	stub := &stubReportService{}
	h := NewReportHandler(stub)
	rec := httptest.NewRecorder()
	body := `{"enabled":true,"recipientEmail":"ops@example.com","schedule":"monthly-1st","scope":"all-resources"}`
	h.UpdateSettings(rec, httptest.NewRequest(http.MethodPut, "/reports/settings", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d body=%s", rec.Code, rec.Body.String())
	}
	if stub.lastSaved == nil || stub.lastSaved.RecipientEmail != "ops@example.com" {
		t.Fatalf("not forwarded: %+v", stub.lastSaved)
	}
}

func TestReportHandler_UpdateSettings_ValidationIs422(t *testing.T) {
	h := NewReportHandler(&stubReportService{saveErr: fmt.Errorf("%w: recipientEmail", service.ErrReportValidation)})
	rec := httptest.NewRecorder()
	h.UpdateSettings(rec, httptest.NewRequest(http.MethodPut, "/reports/settings", strings.NewReader(`{"enabled":true,"recipientEmail":""}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("code = %d, want 422", rec.Code)
	}
}

func TestReportHandler_History_EmptyEnvelope(t *testing.T) {
	h := NewReportHandler(&stubReportService{history: nil})
	rec := httptest.NewRecorder()
	h.History(rec, httptest.NewRequest(http.MethodGet, "/reports/history?limit=5", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `[]`) {
		t.Fatalf("empty history should be []: %s", rec.Body.String())
	}
}

func TestReportHandler_Preview_Object(t *testing.T) {
	h := NewReportHandler(&stubReportService{preview: &domain.ReportHistory{Period: "2026-07", Status: domain.ReportStatusPending}})
	rec := httptest.NewRecorder()
	h.Preview(rec, httptest.NewRequest(http.MethodGet, "/reports/preview", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"period":"2026-07"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}
