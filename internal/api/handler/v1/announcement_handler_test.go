package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/go-chi/chi/v5"
)

type stubAnnouncementService struct {
	list      []*domain.Announcement
	created   *domain.Announcement
	createErr error
	deleteErr error
}

func (s *stubAnnouncementService) ListActive(context.Context) ([]*domain.Announcement, error) {
	return s.list, nil
}
func (s *stubAnnouncementService) Create(_ context.Context, in *domain.Announcement) (*domain.Announcement, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	in.Base.ID = "a1"
	s.created = in
	return in, nil
}
func (s *stubAnnouncementService) Delete(context.Context, string) error { return s.deleteErr }

func withAnnID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestAnnHandler_List_EmptyEnvelope(t *testing.T) {
	h := NewAnnouncementHandler(&stubAnnouncementService{})
	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/announcements", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "[]") {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAnnHandler_Create_DefaultsDismissible(t *testing.T) {
	stub := &stubAnnouncementService{}
	h := NewAnnouncementHandler(stub)
	rec := httptest.NewRecorder()
	h.Create(rec, httptest.NewRequest(http.MethodPost, "/announcements", strings.NewReader(`{"severity":"info","title":"Hi"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	if stub.created == nil || !stub.created.Dismissible {
		t.Fatalf("dismissible should default true: %+v", stub.created)
	}
}

func TestAnnHandler_Delete_204(t *testing.T) {
	h := NewAnnouncementHandler(&stubAnnouncementService{})
	rec := httptest.NewRecorder()
	h.Delete(rec, withAnnID(httptest.NewRequest(http.MethodDelete, "/announcements/a1", nil), "a1"))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("code=%d", rec.Code)
	}
}
