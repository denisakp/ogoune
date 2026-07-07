package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/dto"
)

type stubStatusSnapshot struct {
	out *dto.PublicStatus
	err error
}

func (s *stubStatusSnapshot) GetCurrent(context.Context) (*dto.PublicStatus, error) {
	return s.out, s.err
}
func (s *stubStatusSnapshot) GetIncidents(context.Context, time.Time, time.Time, string) (*dto.PublicIncidentsResponse, error) {
	return nil, nil
}
func (s *stubStatusSnapshot) GetUptime(context.Context, string, time.Time, time.Time) (*dto.PublicUptimeResponse, error) {
	return nil, nil
}
func (s *stubStatusSnapshot) GetIncidentDetail(context.Context, string) (*dto.PublicIncidentDetail, error) {
	return nil, nil
}
func (s *stubStatusSnapshot) GetResourceWindows(context.Context, string) (*dto.PublicResourceWindowsResponse, error) {
	return nil, nil
}

func writeStatusHTML(t *testing.T, dir, body string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "status.html"), []byte(body), 0o644))
}

func TestStaticStatus_InjectsTitleAndMeta(t *testing.T) {
	dir := t.TempDir()
	writeStatusHTML(t, dir, `<html><head><title>placeholder</title></head><body><div id="app"></div></body></html>`)

	stub := &stubStatusSnapshot{out: &dto.PublicStatus{
		Branding: dto.PublicBranding{Name: "Acme"},
		Verdict:  dto.PublicVerdict{Status: dto.VerdictOperational, Label: "All Systems Operational"},
	}}
	h := NewStaticStatusHandler(dir, stub)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rec.Body.String()

	assert.Contains(t, body, "<title>Acme — All Systems Operational</title>")
	assert.Contains(t, body, `name="x-ogoune-license"`)
}

func TestStaticStatus_FallsBackToGenericTitleOnError(t *testing.T) {
	dir := t.TempDir()
	writeStatusHTML(t, dir, `<html><head><title>placeholder</title></head></html>`)
	stub := &stubStatusSnapshot{err: assert.AnError}
	h := NewStaticStatusHandler(dir, stub)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Contains(t, rec.Body.String(), "<title>Status Page</title>")
}

func TestStaticStatus_AssetRequestStreams(t *testing.T) {
	dir := t.TempDir()
	writeStatusHTML(t, dir, "<html></html>")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "favicon.ico"), []byte("ICO_BYTES"), 0o644))
	h := NewStaticStatusHandler(dir, &stubStatusSnapshot{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/favicon.ico", nil))
	assert.Equal(t, "ICO_BYTES", rec.Body.String())
}

func TestStaticStatus_SPANavigationFallback(t *testing.T) {
	dir := t.TempDir()
	writeStatusHTML(t, dir, `<html><head><title>x</title></head></html>`)
	h := NewStaticStatusHandler(dir, &stubStatusSnapshot{})
	rec := httptest.NewRecorder()
	// /history doesn't exist on disk → fall back to status.html.
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/history", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<title>")
}
