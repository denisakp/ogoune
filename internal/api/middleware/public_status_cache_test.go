package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicStatusCache_SetsCacheControl(t *testing.T) {
	h := PublicStatusCache(15, 30, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	h := PublicStatusCache(0, -5, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	got := rec.Header().Get("Cache-Control")
	if !strings.Contains(got, "max-age=15") || !strings.Contains(got, "stale-while-revalidate=0") {
		t.Errorf("default Cache-Control: got %q", got)
	}
}

type stubRecorder struct{ hits, misses int }

func (s *stubRecorder) RecordHit()  { s.hits++ }
func (s *stubRecorder) RecordMiss() { s.misses++ }

func TestPublicStatusCache_HitMissCounters(t *testing.T) {
	rec := &stubRecorder{}
	h := PublicStatusCache(15, 30, rec)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))

	// Fresh request → miss.
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.misses != 1 || rec.hits != 0 {
		t.Fatalf("after miss: hits=%d misses=%d", rec.hits, rec.misses)
	}

	// Conditional request → hit.
	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", `"etag"`)
	h.ServeHTTP(rr, req)
	if rec.hits != 1 || rec.misses != 1 {
		t.Fatalf("after hit: hits=%d misses=%d", rec.hits, rec.misses)
	}
}
