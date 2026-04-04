package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHeartbeatPingService struct {
	getResourceByHeartbeatSlugFunc func(ctx context.Context, slug string) (*domain.Resource, error)
	markHeartbeatPingFunc          func(ctx context.Context, resourceID string, at time.Time) error
	handleHeartbeatRecoveryFunc    func(ctx context.Context, resource *domain.Resource) error
}

func (m *mockHeartbeatPingService) GetResourceByHeartbeatSlug(ctx context.Context, slug string) (*domain.Resource, error) {
	if m.getResourceByHeartbeatSlugFunc != nil {
		return m.getResourceByHeartbeatSlugFunc(ctx, slug)
	}
	return nil, service.ErrResourceNotFound
}

func (m *mockHeartbeatPingService) MarkHeartbeatPing(ctx context.Context, resourceID string, at time.Time) error {
	if m.markHeartbeatPingFunc != nil {
		return m.markHeartbeatPingFunc(ctx, resourceID, at)
	}
	return nil
}

func (m *mockHeartbeatPingService) HandleHeartbeatRecovery(ctx context.Context, resource *domain.Resource) error {
	if m.handleHeartbeatRecoveryFunc != nil {
		return m.handleHeartbeatRecoveryFunc(ctx, resource)
	}
	return nil
}

func newPingTestRouter(h *PingHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/ping/{slug}", h.Ping)
	r.Post("/ping/{slug}", h.Ping)
	return r
}

func TestPingHandler_SuccessGETAndPOST(t *testing.T) {
	slug := "550e8400-e29b-41d4-a716-446655440111"
	mockSvc := &mockHeartbeatPingService{
		getResourceByHeartbeatSlugFunc: func(ctx context.Context, s string) (*domain.Resource, error) {
			assert.Equal(t, slug, s)
			return &domain.Resource{Base: domain.Base{ID: "hb-1"}, Name: "Nightly", Type: domain.ResourceHeartbeat, IsActive: true}, nil
		},
	}
	h := NewPingHandler(mockSvc)
	r := newPingTestRouter(h)

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/ping/"+slug, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)

			var body map[string]string
			err := json.NewDecoder(rec.Body).Decode(&body)
			require.NoError(t, err)
			assert.Equal(t, "ok", body["status"])
			assert.Equal(t, "Nightly", body["monitor"])
		})
	}
}

func TestPingHandler_ErrorMappings(t *testing.T) {
	validSlug := "550e8400-e29b-41d4-a716-446655440112"

	t.Run("422 invalid slug", func(t *testing.T) {
		h := NewPingHandler(&mockHeartbeatPingService{})
		r := newPingTestRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/ping/not-a-slug", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("404 unknown monitor", func(t *testing.T) {
		h := NewPingHandler(&mockHeartbeatPingService{
			getResourceByHeartbeatSlugFunc: func(ctx context.Context, slug string) (*domain.Resource, error) {
				return nil, service.ErrResourceNotFound
			},
		})
		r := newPingTestRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/ping/"+validSlug, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("403 paused monitor", func(t *testing.T) {
		h := NewPingHandler(&mockHeartbeatPingService{
			getResourceByHeartbeatSlugFunc: func(ctx context.Context, slug string) (*domain.Resource, error) {
				return &domain.Resource{Base: domain.Base{ID: "hb-1"}, Type: domain.ResourceHeartbeat, IsActive: false}, nil
			},
		})
		r := newPingTestRouter(h)
		req := httptest.NewRequest(http.MethodGet, "/ping/"+validSlug, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestPingHandler_RateLimit429(t *testing.T) {
	slug := "550e8400-e29b-41d4-a716-446655440113"
	h := NewPingHandler(&mockHeartbeatPingService{
		getResourceByHeartbeatSlugFunc: func(ctx context.Context, slug string) (*domain.Resource, error) {
			return &domain.Resource{Base: domain.Base{ID: "hb-1"}, Name: "Batch", Type: domain.ResourceHeartbeat, IsActive: true}, nil
		},
	})
	r := newPingTestRouter(h)

	for i := 1; i <= 101; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping/"+slug, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if i <= 100 {
			assert.Equal(t, http.StatusOK, rec.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, rec.Code)
		}
	}
}

func TestPingHandler_ConcurrentLastWriteWins(t *testing.T) {
	slug := "550e8400-e29b-41d4-a716-446655440114"
	base := time.Unix(1700000000, 0).UTC()
	var seq int64
	var mu sync.Mutex
	var calls int
	var lastAt time.Time

	h := NewPingHandler(&mockHeartbeatPingService{
		getResourceByHeartbeatSlugFunc: func(ctx context.Context, slug string) (*domain.Resource, error) {
			return &domain.Resource{Base: domain.Base{ID: "hb-1"}, Name: "Concurrent", Type: domain.ResourceHeartbeat, IsActive: true}, nil
		},
		markHeartbeatPingFunc: func(ctx context.Context, resourceID string, at time.Time) error {
			mu.Lock()
			defer mu.Unlock()
			calls++
			if at.After(lastAt) {
				lastAt = at
			}
			return nil
		},
	})
	h.now = func() time.Time {
		n := atomic.AddInt64(&seq, 1)
		return base.Add(time.Duration(n) * time.Nanosecond)
	}
	r := newPingTestRouter(h)

	var wg sync.WaitGroup
	statuses := make([]int, 2)
	errCh := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/ping/"+slug, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			statuses[idx] = rec.Code
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	assert.Equal(t, http.StatusOK, statuses[0])
	assert.Equal(t, http.StatusOK, statuses[1])
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 2, calls)
	assert.True(t, lastAt.Equal(base.Add(2*time.Nanosecond)))
}

func TestPingHandler_RecoveryCalledWhenMonitorIsDown(t *testing.T) {
	slug := "550e8400-e29b-41d4-a716-446655440116"
	recoveryCalled := false

	h := NewPingHandler(&mockHeartbeatPingService{
		getResourceByHeartbeatSlugFunc: func(ctx context.Context, s string) (*domain.Resource, error) {
			return &domain.Resource{
				Base:     domain.Base{ID: "hb-down"},
				Name:     "Backup",
				Type:     domain.ResourceHeartbeat,
				IsActive: true,
				Status:   domain.StatusDown,
			}, nil
		},
		handleHeartbeatRecoveryFunc: func(ctx context.Context, resource *domain.Resource) error {
			recoveryCalled = true
			return nil
		},
	})
	r := newPingTestRouter(h)
	req := httptest.NewRequest(http.MethodPost, "/ping/"+slug, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, recoveryCalled, "HandleHeartbeatRecovery must be called when monitor is down")
}

func TestPingHandler_RecoveryNotCalledWhenMonitorIsUp(t *testing.T) {
	slug := "550e8400-e29b-41d4-a716-446655440117"
	recoveryCalled := false

	h := NewPingHandler(&mockHeartbeatPingService{
		getResourceByHeartbeatSlugFunc: func(ctx context.Context, s string) (*domain.Resource, error) {
			return &domain.Resource{
				Base:     domain.Base{ID: "hb-up"},
				Name:     "Nightly",
				Type:     domain.ResourceHeartbeat,
				IsActive: true,
				Status:   domain.StatusUp,
			}, nil
		},
		handleHeartbeatRecoveryFunc: func(ctx context.Context, resource *domain.Resource) error {
			recoveryCalled = true
			return nil
		},
	})
	r := newPingTestRouter(h)
	req := httptest.NewRequest(http.MethodPost, "/ping/"+slug, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, recoveryCalled, "HandleHeartbeatRecovery must NOT be called when monitor is already up")
}

func TestPingHandler_RecoveryFailureIsNonFatal(t *testing.T) {
	slug := "550e8400-e29b-41d4-a716-446655440118"

	h := NewPingHandler(&mockHeartbeatPingService{
		getResourceByHeartbeatSlugFunc: func(ctx context.Context, s string) (*domain.Resource, error) {
			return &domain.Resource{
				Base:     domain.Base{ID: "hb-fail"},
				Name:     "Failing",
				Type:     domain.ResourceHeartbeat,
				IsActive: true,
				Status:   domain.StatusDown,
			}, nil
		},
		handleHeartbeatRecoveryFunc: func(ctx context.Context, resource *domain.Resource) error {
			return errors.New("incident service unavailable")
		},
	})
	r := newPingTestRouter(h)
	req := httptest.NewRequest(http.MethodPost, "/ping/"+slug, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Recovery failure is non-fatal; ping still returns 200
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPingHandler_LastPingAtPersisted(t *testing.T) {
	slug := "550e8400-e29b-41d4-a716-446655440119"
	before := time.Now().UTC()

	var capturedAt time.Time
	h := NewPingHandler(&mockHeartbeatPingService{
		getResourceByHeartbeatSlugFunc: func(ctx context.Context, s string) (*domain.Resource, error) {
			return &domain.Resource{
				Base:     domain.Base{ID: "hb-persist"},
				Name:     "Persistence Test",
				Type:     domain.ResourceHeartbeat,
				IsActive: true,
				Status:   domain.StatusUp,
			}, nil
		},
		markHeartbeatPingFunc: func(ctx context.Context, resourceID string, at time.Time) error {
			capturedAt = at
			return nil
		},
	})
	r := newPingTestRouter(h)
	req := httptest.NewRequest(http.MethodPost, "/ping/"+slug, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	after := time.Now().UTC()
	require.Equal(t, http.StatusOK, rec.Code)

	// SC-006: last_ping_at must be set to a real timestamp — not nil, not zero
	require.False(t, capturedAt.IsZero(), "last_ping_at must be written with a real timestamp on ping receipt")
	assert.True(t, capturedAt.After(before) || capturedAt.Equal(before), "last_ping_at must be >= request time")
	assert.True(t, capturedAt.Before(after) || capturedAt.Equal(after), "last_ping_at must be <= response time")
}

func TestIsHeartbeatWaiting_NonNilTimestampIsNotWaiting(t *testing.T) {
	// FR-004: a non-nil last_ping_at MUST NOT keep a monitor in waiting state,
	// even if the pointer points to a zero-value time.Time.
	zeroTime := time.Time{}
	realTime := time.Now().UTC()

	t.Run("nil last_ping_at is waiting", func(t *testing.T) {
		r := &domain.Resource{Type: domain.ResourceHeartbeat, LastPingAt: nil}
		assert.True(t, r.IsHeartbeatWaiting(), "nil LastPingAt should mean waiting=true")
	})

	t.Run("non-nil last_ping_at with zero value is NOT waiting", func(t *testing.T) {
		r := &domain.Resource{Type: domain.ResourceHeartbeat, LastPingAt: &zeroTime}
		assert.False(t, r.IsHeartbeatWaiting(), "non-nil LastPingAt (even zero-value) should mean waiting=false")
	})

	t.Run("non-nil last_ping_at with real timestamp is NOT waiting", func(t *testing.T) {
		r := &domain.Resource{Type: domain.ResourceHeartbeat, LastPingAt: &realTime}
		assert.False(t, r.IsHeartbeatWaiting(), "non-nil LastPingAt with real timestamp should mean waiting=false")
	})

	t.Run("non-heartbeat type is never waiting", func(t *testing.T) {
		r := &domain.Resource{Type: domain.ResourceHTTP, LastPingAt: nil}
		assert.False(t, r.IsHeartbeatWaiting(), "non-heartbeat resource should never be waiting")
	})
}

func TestPingHandler_InternalError(t *testing.T) {
	slug := "550e8400-e29b-41d4-a716-446655440115"
	h := NewPingHandler(&mockHeartbeatPingService{
		getResourceByHeartbeatSlugFunc: func(ctx context.Context, slug string) (*domain.Resource, error) {
			return nil, errors.New("db down")
		},
	})
	r := newPingTestRouter(h)
	req := httptest.NewRequest(http.MethodGet, "/ping/"+slug, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
