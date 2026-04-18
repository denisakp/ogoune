package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v1 "github.com/denisakp/ogoune/internal/api/handler/v1"
	"github.com/denisakp/ogoune/internal/api/middleware"
	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock notification channel service ---

type mockChannelService struct {
	channels  []*domain.NotificationChannel
	channel   *domain.NotificationChannel
	listErr   error
	getErr    error
	createErr error
	updateErr error
	deleteErr error
}

func (m *mockChannelService) ListNotificationChannels(_ context.Context, limit, offset int) ([]*domain.NotificationChannel, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	end := offset + limit
	if end > len(m.channels) {
		end = len(m.channels)
	}
	if offset > len(m.channels) {
		return []*domain.NotificationChannel{}, nil
	}
	return m.channels[offset:end], nil
}

func (m *mockChannelService) GetNotificationChannel(_ context.Context, _ string) (*domain.NotificationChannel, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.channel, nil
}

func (m *mockChannelService) CreateNotificationChannel(_ context.Context, payload *dto.CreateNotificationChannelPayload) (*domain.NotificationChannel, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &domain.NotificationChannel{
		Base: domain.Base{ID: "new-ch", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name: payload.Name,
		Type: payload.Type,
	}, nil
}

func (m *mockChannelService) UpdateNotificationChannel(_ context.Context, id string, _ *dto.UpdateNotificationChannelPayload) (*domain.NotificationChannel, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return &domain.NotificationChannel{Base: domain.Base{ID: id, CreatedAt: time.Now(), UpdatedAt: time.Now()}}, nil
}

func (m *mockChannelService) DeleteNotificationChannel(_ context.Context, _ string) error {
	return m.deleteErr
}

func newChannelRouter(svc v1.ChannelV1ServiceInterface) *chi.Mux {
	r := chi.NewRouter()
	h := v1.NewNotificationChannelHandler(svc)
	r.Get("/api/v1/notification-channels", h.List)
	r.With(middleware.RequireReadWrite).Post("/api/v1/notification-channels", h.Create)
	r.Get("/api/v1/notification-channels/{id}", h.Get)
	r.With(middleware.RequireReadWrite).Put("/api/v1/notification-channels/{id}", h.Update)
	r.With(middleware.RequireReadWrite).Delete("/api/v1/notification-channels/{id}", h.Delete)
	return r
}

// T025: Notification channel scope tests

func TestChannelHandler_ScopeEnforcement_ReadKey_Returns403OnWrite(t *testing.T) {
	svc := &mockChannelService{}
	router := newChannelRouter(svc)

	cases := []struct {
		method string
		path   string
		body   []byte
	}{
		{"POST", "/api/v1/notification-channels", []byte(`{"name":"x","type":"smtp","config":{}}`)},
		{"PUT", "/api/v1/notification-channels/abc", []byte(`{}`)},
		{"DELETE", "/api/v1/notification-channels/abc", nil},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			var body *bytes.Reader
			if tc.body != nil {
				body = bytes.NewReader(tc.body)
			} else {
				body = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(tc.method, tc.path, body)
			req = injectReadScope(req)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusForbidden, rr.Code, "%s %s with read scope should be 403", tc.method, tc.path)
		})
	}
}

func TestChannelHandler_ScopeEnforcement_ReadWriteKey_NotForbiddenOnWrite(t *testing.T) {
	ch := &domain.NotificationChannel{
		Base:   domain.Base{ID: "ch-1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name:   "Test",
		Type:   domain.NotificationChannelTypeSMTP,
		Config: []byte(`{}`),
	}
	svc := &mockChannelService{channels: []*domain.NotificationChannel{ch}, channel: ch}
	router := newChannelRouter(svc)

	cases := []struct {
		method string
		path   string
		body   []byte
	}{
		{"POST", "/api/v1/notification-channels", []byte(`{"name":"x","type":"smtp","config":{}}`)},
		{"PUT", "/api/v1/notification-channels/ch-1", []byte(`{}`)},
		{"DELETE", "/api/v1/notification-channels/ch-1", nil},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			var body *bytes.Reader
			if tc.body != nil {
				body = bytes.NewReader(tc.body)
			} else {
				body = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(tc.method, tc.path, body)
			req = injectReadWriteScope(req)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.NotEqual(t, http.StatusForbidden, rr.Code)
		})
	}
}

func TestChannelHandler_Response_DoesNotExposePassword(t *testing.T) {
	ch := &domain.NotificationChannel{
		Base:   domain.Base{ID: "ch-1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name:   "SMTP",
		Type:   domain.NotificationChannelTypeSMTP,
		Config: []byte(`{"host":"smtp.example.com","port":587,"username":"user","password":"secret"}`),
	}
	svc := &mockChannelService{channel: ch}
	router := newChannelRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notification-channels/ch-1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.NotContains(t, body, "secret", "response should not expose password")

	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	data := out["data"].(map[string]interface{})
	config, ok := data["config"].(map[string]interface{})
	if ok {
		_, hasPassword := config["password"]
		assert.False(t, hasPassword, "config should not include password field")
	}
}
