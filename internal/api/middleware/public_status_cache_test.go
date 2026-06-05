package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicStatusCache_SetsCacheControl(t *testing.T) {
	h := PublicStatusCache(15, 30)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/status/foo", nil))
	got := rec.Header().Get("Cache-Control")
	if !strings.Contains(got, "max-age=15") || !strings.Contains(got, "stale-while-revalidate=30") {
		t.Errorf("Cache-Control: got %q want max-age=15 + swr=30", got)
	}
	if rec.Header().Get("Vary") != "Accept-Encoding" {
		t.Errorf("Vary: got %q want Accept-Encoding", rec.Header().Get("Vary"))
	}
}

func TestPublicStatusCache_AppliesDefaults(t *testing.T) {
	h := PublicStatusCache(0, -5)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	got := rec.Header().Get("Cache-Control")
	if !strings.Contains(got, "max-age=15") || !strings.Contains(got, "stale-while-revalidate=0") {
		t.Errorf("default Cache-Control: got %q", got)
	}
}
