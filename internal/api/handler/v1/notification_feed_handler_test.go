package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
)

type stubFeedService struct {
	items     []*domain.FeedNotification
	total     int64
	markErr   error
	markedAll int64
	lastCat   *string
}

func (s *stubFeedService) ListForUser(_ context.Context, _ string, category *string, _, _ int) ([]*domain.FeedNotification, int64, error) {
	s.lastCat = category
	return s.items, s.total, nil
}
func (s *stubFeedService) MarkRead(_ context.Context, _ string) error { return s.markErr }
func (s *stubFeedService) MarkAllRead(_ context.Context, _ string, _ time.Time) (int64, error) {
	return s.markedAll, nil
}

func TestNotifHandler_List_OK(t *testing.T) {
	desc := "d"
	link := "/incidents/x"
	h := NewNotificationFeedHandler(&stubFeedService{
		items: []*domain.FeedNotification{{
			Base: domain.Base{ID: "n1"}, Category: "incident", Severity: "error",
			Title: "down", Description: &desc, DeepLink: &link, OccurredAt: time.Now(),
		}},
		total: 1,
	})
	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{`"id":"n1"`, `"unread":true`, `"deepLink":"/incidents/x"`, `"total":1`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q: %s", want, body)
		}
	}
}

func TestNotifHandler_List_PaginationBounds(t *testing.T) {
	h := NewNotificationFeedHandler(&stubFeedService{})
	// invalid per_page
	req := httptest.NewRequest(http.MethodGet, "/notifications?per_page=0", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid per_page: got %d want 422", rec.Code)
	}
	// invalid category
	req2 := httptest.NewRequest(http.MethodGet, "/notifications?category=bogus", nil)
	rec2 := httptest.NewRecorder()
	h.List(rec2, req2)
	if rec2.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid category: got %d want 422", rec2.Code)
	}
}

func TestNotifHandler_MarkRead_NotFound(t *testing.T) {
	h := NewNotificationFeedHandler(&stubFeedService{markErr: service.ErrNotificationNotFound})
	req := httptest.NewRequest(http.MethodPost, "/notifications/x/read", nil)
	rec := httptest.NewRecorder()
	// inject chi URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "x")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	h.MarkRead(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d want 404", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "NOTIFICATION_NOT_FOUND") {
		t.Fatalf("missing code: %s", rec.Body.String())
	}
}

func TestNotifHandler_MarkRead_OK(t *testing.T) {
	h := NewNotificationFeedHandler(&stubFeedService{})
	req := httptest.NewRequest(http.MethodPost, "/notifications/x/read", nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "x")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	h.MarkRead(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("got %d want 204", rec.Code)
	}
}

func TestNotifHandler_MarkAllRead(t *testing.T) {
	h := NewNotificationFeedHandler(&stubFeedService{markedAll: 4})
	req := httptest.NewRequest(http.MethodPost, "/notifications/read-all", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	h.MarkAllRead(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"marked":4`) {
		t.Fatalf("missing marked count: %s", rec.Body.String())
	}
}
