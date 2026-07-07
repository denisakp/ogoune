package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
)

type stubDashboardService struct {
	list    []*domain.Dashboard
	one     *domain.Dashboard
	err     error
}

func (s *stubDashboardService) List(context.Context, int, int) ([]*domain.Dashboard, error) {
	return s.list, s.err
}
func (s *stubDashboardService) Get(context.Context, string) (*domain.Dashboard, error) {
	return s.one, s.err
}
func (s *stubDashboardService) Create(_ context.Context, _ string, d *domain.Dashboard) (*domain.Dashboard, error) {
	if s.err != nil {
		return nil, s.err
	}
	return d, nil
}
func (s *stubDashboardService) Update(context.Context, string, string, service.DashboardUpdate) (*domain.Dashboard, error) {
	return s.one, s.err
}
func (s *stubDashboardService) SaveLayout(context.Context, string, string, []domain.WidgetInstance) (*domain.Dashboard, error) {
	return s.one, s.err
}
func (s *stubDashboardService) Delete(context.Context, string, string) error { return s.err }

func withID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func sampleDash() *domain.Dashboard {
	return &domain.Dashboard{
		Base: domain.Base{ID: "d1"}, OwnerID: "alice", OwnerName: "Alice", Name: "Prod",
		Scope: domain.DashboardScope{Mode: domain.DashboardScopeModeTag},
		Widgets: []domain.WidgetInstance{{ID: "w1", WidgetTypeID: domain.WidgetTypeUptimeStat, Position: 0}},
		DefaultTimeRange: "24h", RefreshInterval: "1m", Visibility: "team",
	}
}

func TestDashHandler_List_OK(t *testing.T) {
	h := NewDashboardHandler(&stubDashboardService{list: []*domain.Dashboard{sampleDash()}})
	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/dashboards", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200; %s", rec.Code, rec.Body.String())
	}
	for _, want := range []string{`"id":"d1"`, `"ownerName":"Alice"`, `"defaultTimeRange":"24h"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("missing %q: %s", want, rec.Body.String())
		}
	}
}

func TestDashHandler_Get_NotFound(t *testing.T) {
	h := NewDashboardHandler(&stubDashboardService{err: service.ErrDashboardNotFound})
	rec := httptest.NewRecorder()
	h.Get(rec, withID(httptest.NewRequest(http.MethodGet, "/dashboards/x", nil), "x"))
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "DASHBOARD_NOT_FOUND") {
		t.Fatalf("got %d / %s", rec.Code, rec.Body.String())
	}
}

func TestDashHandler_Create_OK_and_Validation(t *testing.T) {
	h := NewDashboardHandler(&stubDashboardService{})
	body := `{"name":"Prod","scope":{"mode":"tag","payload":{}},"widgets":[],"defaultTimeRange":"24h","refreshInterval":"1m","visibility":"team"}`
	rec := httptest.NewRecorder()
	h.Create(rec, httptest.NewRequest(http.MethodPost, "/dashboards", strings.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d, want 201; %s", rec.Code, rec.Body.String())
	}

	h2 := NewDashboardHandler(&stubDashboardService{err: service.ErrDashboardValidation})
	rec2 := httptest.NewRecorder()
	h2.Create(rec2, httptest.NewRequest(http.MethodPost, "/dashboards", strings.NewReader(body)))
	if rec2.Code != http.StatusUnprocessableEntity {
		t.Fatalf("validation: got %d, want 422", rec2.Code)
	}
}

func TestDashHandler_Update_Forbidden(t *testing.T) {
	h := NewDashboardHandler(&stubDashboardService{err: service.ErrDashboardForbidden})
	rec := httptest.NewRecorder()
	req := withID(httptest.NewRequest(http.MethodPatch, "/dashboards/d1", strings.NewReader(`{"name":"x"}`)), "d1")
	h.Update(rec, req)
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "FORBIDDEN") {
		t.Fatalf("got %d / %s", rec.Code, rec.Body.String())
	}
}

func TestDashHandler_SaveLayout_OK(t *testing.T) {
	h := NewDashboardHandler(&stubDashboardService{one: sampleDash()})
	rec := httptest.NewRecorder()
	body := `{"widgets":[{"id":"w1","widgetTypeId":"uptime-stat","position":0}]}`
	req := withID(httptest.NewRequest(http.MethodPut, "/dashboards/d1/layout", strings.NewReader(body)), "d1")
	h.SaveLayout(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200; %s", rec.Code, rec.Body.String())
	}
}

func TestDashHandler_Delete(t *testing.T) {
	h := NewDashboardHandler(&stubDashboardService{})
	rec := httptest.NewRecorder()
	h.Delete(rec, withID(httptest.NewRequest(http.MethodDelete, "/dashboards/d1", nil), "d1"))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("got %d, want 204", rec.Code)
	}

	h2 := NewDashboardHandler(&stubDashboardService{err: service.ErrDashboardForbidden})
	rec2 := httptest.NewRecorder()
	h2.Delete(rec2, withID(httptest.NewRequest(http.MethodDelete, "/dashboards/d1", nil), "d1"))
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("non-owner delete: got %d, want 403", rec2.Code)
	}
}
