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
	out *dto.PublicStatus
	err error
}

func (s *stubPublicStatus) GetCurrent(context.Context) (*dto.PublicStatus, error) {
	return s.out, s.err
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
