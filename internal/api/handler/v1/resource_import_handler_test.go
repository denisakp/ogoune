package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denisakp/ogoune/internal/api/middleware"
	"github.com/denisakp/ogoune/internal/domain"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/denisakp/ogoune/internal/service/resourceimport"
)

type stubImportSvc struct {
	report *dtoV1.ImportReport
	err    error
	yaml   []byte
}

func (s *stubImportSvc) Import(_ context.Context, _ []byte, _ dtoV1.ImportOptions) (*dtoV1.ImportReport, error) {
	return s.report, s.err
}

func (s *stubImportSvc) ExportYAML(_ context.Context) ([]byte, error) {
	return s.yaml, nil
}

func decodeReport(t *testing.T, body []byte) dtoV1.ImportReport {
	t.Helper()
	var env struct {
		Data dtoV1.ImportReport `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, body)
	}
	return env.Data
}

func TestImportHandler_DryRunReturnsReport(t *testing.T) {
	svc := &stubImportSvc{report: &dtoV1.ImportReport{DryRun: true, Total: 2, Rows: []dtoV1.RowResult{
		{Index: 0, Name: "A", Valid: true, Action: dtoV1.RowActionCreate},
		{Index: 1, Name: "B", Valid: true, Action: dtoV1.RowActionCreate},
	}}}
	h := NewResourceImportHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/monitors/import?dryRun=true", strings.NewReader("version: 1"))
	req.Header.Set("Content-Type", "text/yaml")
	rec := httptest.NewRecorder()
	h.Import(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	report := decodeReport(t, rec.Body.Bytes())
	if !report.DryRun || report.Total != 2 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestImportHandler_ValidationFailedReturns422WithReport(t *testing.T) {
	svc := &stubImportSvc{
		report: &dtoV1.ImportReport{Total: 1, Failed: 1, Rows: []dtoV1.RowResult{
			{Index: 0, Name: "Bad", Valid: false, Action: dtoV1.RowActionError, Errors: []string{"target is required"}},
		}},
		err: resourceimport.ErrValidationFailed,
	}
	h := NewResourceImportHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/monitors/import", strings.NewReader("version: 1"))
	req.Header.Set("Content-Type", "text/yaml")
	rec := httptest.NewRecorder()
	h.Import(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	report := decodeReport(t, rec.Body.Bytes())
	if report.Failed != 1 || len(report.Rows) != 1 {
		t.Fatalf("expected per-row report, got %+v", report)
	}
}

func TestImportHandler_EmptyBodyIsBadRequest(t *testing.T) {
	h := NewResourceImportHandler(&stubImportSvc{})
	req := httptest.NewRequest(http.MethodPost, "/monitors/import", strings.NewReader("   "))
	req.Header.Set("Content-Type", "text/yaml")
	rec := httptest.NewRecorder()
	h.Import(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestImportHandler_ReadOnlyKeyForbidden(t *testing.T) {
	h := NewResourceImportHandler(&stubImportSvc{report: &dtoV1.ImportReport{}})
	guarded := middleware.RequireReadWrite(http.HandlerFunc(h.Import))

	req := httptest.NewRequest(http.MethodPost, "/monitors/import", strings.NewReader("version: 1"))
	req.Header.Set("Content-Type", "text/yaml")
	ctx := context.WithValue(req.Context(), "auth_method", "api_key")
	ctx = context.WithValue(ctx, "api_key_scope", domain.APIKeyScopeRead)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	guarded.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestExportHandler_ReturnsYAML(t *testing.T) {
	svc := &stubImportSvc{yaml: []byte("version: 1\nresources: []\n")}
	h := NewResourceImportHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/monitors/export", nil)
	rec := httptest.NewRecorder()
	h.Export(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/yaml") {
		t.Fatalf("content-type = %q, want text/yaml", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "ogoune-monitors.yaml") {
		t.Fatalf("content-disposition = %q", cd)
	}
	if !strings.Contains(rec.Body.String(), "version: 1") {
		t.Fatalf("body missing manifest: %s", rec.Body.String())
	}
}
