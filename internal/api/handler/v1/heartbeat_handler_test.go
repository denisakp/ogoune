package v1_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v1 "github.com/denisakp/ogoune/internal/api/handler/v1"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// --- mock heartbeat service ---

type mockHeartbeatService struct {
	resource    *domain.Resource
	getErr      error
	pingErr     error
	recoveryErr error
}

func (m *mockHeartbeatService) GetResourceByHeartbeatSlug(_ context.Context, _ string) (*domain.Resource, error) {
	return m.resource, m.getErr
}

func (m *mockHeartbeatService) MarkHeartbeatPing(_ context.Context, _ string, _ time.Time) error {
	return m.pingErr
}

func (m *mockHeartbeatService) HandleHeartbeatRecovery(_ context.Context, _ *domain.Resource) error {
	return m.recoveryErr
}

func newHeartbeatRouter(svc v1.HeartbeatV1ServiceInterface) *chi.Mux {
	r := chi.NewRouter()
	h := v1.NewHeartbeatV1Handler(svc)
	r.Post("/api/v1/heartbeat/ping/{slug}", h.Ping)
	return r
}

// T035: Known slug returns 200 with received_at
func TestHeartbeatV1Handler_Ping_KnownSlug_Returns200(t *testing.T) {
	res := &domain.Resource{
		Base:     domain.Base{ID: "r1"},
		Name:     "My Monitor",
		IsActive: true,
		Status:   domain.StatusUp,
	}
	svc := &mockHeartbeatService{resource: res}
	router := newHeartbeatRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/heartbeat/ping/abc-123", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "received_at")
}

// T035: Unknown slug returns 404
func TestHeartbeatV1Handler_Ping_UnknownSlug_Returns404(t *testing.T) {
	svc := &mockHeartbeatService{getErr: service.ErrResourceNotFound}
	router := newHeartbeatRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/heartbeat/ping/unknown-slug", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// T035: Paused monitor returns 403
func TestHeartbeatV1Handler_Ping_PausedMonitor_Returns403(t *testing.T) {
	res := &domain.Resource{
		Base:     domain.Base{ID: "r2"},
		Name:     "Paused Monitor",
		IsActive: false,
	}
	svc := &mockHeartbeatService{resource: res}
	router := newHeartbeatRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/heartbeat/ping/paused-slug", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// T035: Response is in v1 envelope format (data.received_at, meta null)
func TestHeartbeatV1Handler_Ping_ResponseEnvelope(t *testing.T) {
	res := &domain.Resource{
		Base:     domain.Base{ID: "r3"},
		Name:     "My Monitor",
		IsActive: true,
		Status:   domain.StatusUp,
	}
	svc := &mockHeartbeatService{resource: res}
	router := newHeartbeatRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/heartbeat/ping/test-slug", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, `"data"`)
	assert.Contains(t, body, `"received_at"`)
	// meta should be null
	assert.Contains(t, body, `"meta":null`)
}

// T035: MarkPing failure returns 500
func TestHeartbeatV1Handler_Ping_MarkPingFailure_Returns500(t *testing.T) {
	res := &domain.Resource{
		Base:     domain.Base{ID: "r4"},
		Name:     "My Monitor",
		IsActive: true,
		Status:   domain.StatusUp,
	}
	svc := &mockHeartbeatService{resource: res, pingErr: errors.New("db error")}
	router := newHeartbeatRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/heartbeat/ping/test-slug", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
