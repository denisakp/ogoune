package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/dto"
)

type stubPublicStatus struct {
	out       *dto.PublicStatus
	err       error
	incidents *dto.PublicIncidentsResponse
	incErr    error
}

func (s *stubPublicStatus) GetCurrent(context.Context) (*dto.PublicStatus, error) {
	return s.out, s.err
}
func (s *stubPublicStatus) GetIncidents(context.Context, time.Time, time.Time, string) (*dto.PublicIncidentsResponse, error) {
	return s.incidents, s.incErr
}
func (s *stubPublicStatus) GetUptime(context.Context, string, time.Time, time.Time) (*dto.PublicUptimeResponse, error) {
	return &dto.PublicUptimeResponse{}, nil
}

func TestPublicStatusHandler_HappyPath(t *testing.T) {
	stub := &stubPublicStatus{
		out: &dto.PublicStatus{
			GeneratedAt: time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC),
			Verdict:     dto.PublicVerdict{Status: dto.VerdictOperational, Label: "All Systems Operational", Color: "green"},
			Components:  []dto.PublicComponent{},
		},
	}
	h := NewPublicStatusHandler(stub)
	rec := httptest.NewRecorder()
	h.GetCurrent(rec, httptest.NewRequest(http.MethodGet, "/status", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var got dto.PublicStatus
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, dto.VerdictOperational, got.Verdict.Status)
}

func TestPublicStatusHandler_ErrorReturnsProblemJSON(t *testing.T) {
	h := NewPublicStatusHandler(&stubPublicStatus{err: errors.New("boom")})
	rec := httptest.NewRecorder()
	h.GetCurrent(rec, httptest.NewRequest(http.MethodGet, "/status", nil))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, float64(http.StatusInternalServerError), body["status"])
	assert.Equal(t, "internal_error", body["code"])
}

func TestPublicStatusHandler_GetIncidents_HappyPath(t *testing.T) {
	stub := &stubPublicStatus{
		incidents: &dto.PublicIncidentsResponse{
			GeneratedAt: time.Now().UTC(),
			Total:       0,
			Months:      []dto.PublicIncidentMonth{},
		},
	}
	h := NewPublicStatusHandler(stub)
	rec := httptest.NewRecorder()
	h.GetIncidents(rec, httptest.NewRequest(http.MethodGet, "/status/incidents", nil))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestPublicStatusHandler_GetIncidents_FromAfterTo422(t *testing.T) {
	h := NewPublicStatusHandler(&stubPublicStatus{})
	rec := httptest.NewRecorder()
	h.GetIncidents(rec, httptest.NewRequest(http.MethodGet, "/status/incidents?from=2026-06-01&to=2026-05-01", nil))
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	assert.Equal(t, "invalid_range", body["code"])
}

func TestPublicStatusHandler_GetIncidents_InvalidFrom422(t *testing.T) {
	h := NewPublicStatusHandler(&stubPublicStatus{})
	rec := httptest.NewRecorder()
	h.GetIncidents(rec, httptest.NewRequest(http.MethodGet, "/status/incidents?from=garbage", nil))
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestPublicStatusHandler_GetUptime_HappyPath(t *testing.T) {
	h := NewPublicStatusHandler(&stubPublicStatus{})
	rec := httptest.NewRecorder()
	h.GetUptime(rec, httptest.NewRequest(http.MethodGet, "/status/uptime?from=2026-05-01&to=2026-06-01", nil))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestPublicStatusHandler_GetUptime_RangeTooLong422(t *testing.T) {
	h := NewPublicStatusHandler(&stubPublicStatus{})
	rec := httptest.NewRecorder()
	h.GetUptime(rec, httptest.NewRequest(http.MethodGet, "/status/uptime?from=2024-01-01&to=2026-01-01", nil))
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	assert.Equal(t, "range_too_long", body["code"])
}
