package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeStaticFile(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
}

func executeStaticRequest(t *testing.T, handler http.Handler, path string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	body, err := io.ReadAll(rr.Result().Body)
	require.NoError(t, err)
	return rr.Code, string(body)
}

func TestServeStaticFiles_StatusRoutesServeStatusEntryWhenAvailable(t *testing.T) {
	staticDir := t.TempDir()
	writeStaticFile(t, staticDir, "index.html", "INDEX_ENTRY")
	writeStaticFile(t, staticDir, "status.html", "STATUS_ENTRY")

	router := chi.NewRouter()
	serveStaticFiles(router, staticDir)

	code, body := executeStaticRequest(t, router, "/status")
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, body, "STATUS_ENTRY")

	code, body = executeStaticRequest(t, router, "/status/resource-1")
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, body, "STATUS_ENTRY")
}

func TestServeStaticFiles_StatusRoutesFallbackToIndexWhenStatusEntryMissing(t *testing.T) {
	staticDir := t.TempDir()
	writeStaticFile(t, staticDir, "index.html", "INDEX_ENTRY")

	router := chi.NewRouter()
	serveStaticFiles(router, staticDir)

	code, body := executeStaticRequest(t, router, "/status")
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, body, "INDEX_ENTRY")

	code, body = executeStaticRequest(t, router, "/status/resource-1")
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, body, "INDEX_ENTRY")
}

func TestServeStaticFiles_UnmatchedAPIPathsReturn404AndNeverStaticHTML(t *testing.T) {
	staticDir := t.TempDir()
	writeStaticFile(t, staticDir, "index.html", "INDEX_ENTRY")
	writeStaticFile(t, staticDir, "status.html", "STATUS_ENTRY")

	router := chi.NewRouter()
	serveStaticFiles(router, staticDir)

	code, body := executeStaticRequest(t, router, "/api/non-existent-route")
	assert.Equal(t, http.StatusNotFound, code)
	assert.NotContains(t, body, "INDEX_ENTRY")
	assert.NotContains(t, body, "STATUS_ENTRY")
}

func TestServeStaticFiles_NonStatusUnknownRoutesFallbackToIndex(t *testing.T) {
	staticDir := t.TempDir()
	writeStaticFile(t, staticDir, "index.html", "INDEX_ENTRY")
	writeStaticFile(t, staticDir, "status.html", "STATUS_ENTRY")

	router := chi.NewRouter()
	serveStaticFiles(router, staticDir)

	code, body := executeStaticRequest(t, router, "/some-dashboard-deep-link")
	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, body, "INDEX_ENTRY")
}
